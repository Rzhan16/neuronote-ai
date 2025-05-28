from typing import Literal, Dict
import torch
from transformers import AutoModelForSeq2SeqLM, AutoTokenizer
import redis
import os

# Initialize tokenizer and model
tokenizer = AutoTokenizer.from_pretrained("facebook/bart-large-cnn")

# Use quantization in Docker, CPU in local development
if os.environ.get("DOCKER_ENV"):
    from transformers import BitsAndBytesConfig
    quantization_config = BitsAndBytesConfig(
        load_in_4bit=True,
        bnb_4bit_compute_dtype=torch.float16
    )
    model = AutoModelForSeq2SeqLM.from_pretrained(
        "facebook/bart-large-cnn",
        device_map="auto",
        quantization_config=quantization_config,
        torch_dtype=torch.float16
    )
else:
    model = AutoModelForSeq2SeqLM.from_pretrained(
        "facebook/bart-large-cnn",
        torch_dtype=torch.float32  # Use FP32 for better compatibility
    ).to("cpu")  # Use CPU for local development

# Initialize Redis client
redis_client = redis.Redis(
    host='redis',  # Docker service name
    port=6379,
    decode_responses=True
)

def summarise(text: str, style: Literal["bullets", "paragraph"] = "paragraph") -> Dict[str, str]:
    """Summarize text using BART model.
    
    Args:
        text: Input text to summarize
        style: Output style ("bullets" or "paragraph")
        
    Returns:
        Dict with "summary" key containing the summarized text
    """
    # Calculate hash for caching
    cache_key = f"summary:{hash(text)}:{style}"
    
    # Check cache first
    cached_result = redis_client.get(cache_key)
    if cached_result:
        return {"summary": cached_result}
    
    # Prepare input
    inputs = tokenizer(
        text,
        max_length=1024,
        truncation=True,
        return_tensors="pt"
    ).to(model.device)
    
    # Generate summary
    with torch.no_grad():
        outputs = model.generate(
            **inputs,
            max_length=150,
            min_length=40,
            num_beams=4,
            length_penalty=2.0,
            early_stopping=True
        )
    
    # Decode summary
    summary = tokenizer.decode(outputs[0], skip_special_tokens=True)
    
    # Convert to bullets if requested
    if style == "bullets":
        # Split into sentences and convert to bullets
        sentences = [s.strip() for s in summary.split('.') if s.strip()]
        summary = "\n".join(f"• {s}." for s in sentences)
    
    # Cache the result
    redis_client.set(
        cache_key,
        summary,
        ex=3600  # Cache for 1 hour
    )
    
    return {"summary": summary} 