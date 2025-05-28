from typing import List, Dict
import torch
from transformers import AutoModelForSeq2SeqLM, AutoTokenizer
import nltk
import redis

# Download required NLTK data
try:
    nltk.data.find('tokenizers/punkt')
except LookupError:
    nltk.download('punkt')

# Initialize tokenizer and model
tokenizer = AutoTokenizer.from_pretrained("valhalla/t5-base-qg-hl")
model = AutoModelForSeq2SeqLM.from_pretrained(
    "valhalla/t5-base-qg-hl",
    torch_dtype=torch.float32  # Use FP32 for better compatibility
).to("cpu")  # Use CPU for local development

# Initialize Redis client
redis_client = redis.Redis(
    host='redis',  # Docker service name
    port=6379,
    decode_responses=True
)

def generate_qa(text: str, max_q: int = 5) -> List[Dict[str, str]]:
    """Generate question-answer pairs from text.
    
    Args:
        text: Input text to generate questions from
        max_q: Maximum number of questions to generate
        
    Returns:
        List of dicts with "q" and "a" keys for questions and answers
    """
    # Calculate hash for caching
    cache_key = f"qa:{hash(text)}:{max_q}"
    
    # Check cache first
    cached_result = redis_client.get(cache_key)
    if cached_result:
        import json
        return json.loads(cached_result)
    
    # Split text into sentences
    sentences = nltk.sent_tokenize(text)
    
    # Generate QA pairs for each sentence
    qa_pairs = []
    for sentence in sentences[:max_q]:  # Limit to max_q sentences
        # Prepare input
        inputs = tokenizer(
            f"generate question: {sentence}",
            max_length=512,
            truncation=True,
            return_tensors="pt"
        ).to(model.device)
        
        # Generate question
        with torch.no_grad():
            question_ids = model.generate(
                **inputs,
                max_length=64,
                num_beams=4,
                length_penalty=1.5,
                early_stopping=True
            )
        
        question = tokenizer.decode(question_ids[0], skip_special_tokens=True)
        qa_pairs.append({
            "q": question,
            "a": sentence.strip()
        })
    
    # Cache the result
    import json
    redis_client.set(
        cache_key,
        json.dumps(qa_pairs),
        ex=3600  # Cache for 1 hour
    )
    
    return qa_pairs 