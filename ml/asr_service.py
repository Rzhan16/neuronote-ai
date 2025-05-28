from typing import Dict
import hashlib
import io
import numpy as np
from transformers import pipeline
import redis

# Initialize ASR pipeline with Whisper small model
asr_pipeline = pipeline(
    "automatic-speech-recognition",
    model="openai/whisper-small",
    device=-1  # CPU
)

# Initialize Redis client
redis_client = redis.Redis(
    host='redis',  # Docker service name
    port=6379,
    decode_responses=True
)

def transcribe(audio_input: np.ndarray) -> Dict:
    """Transcribe audio to text using Whisper.
    
    Args:
        audio_input: Numpy array of audio samples (normalized to [-1, 1])
        
    Returns:
        Dict with "text" key containing the transcription
    """
    # Calculate hash of the numpy array for caching
    file_hash = hashlib.sha256(audio_input.tobytes()).hexdigest()
    
    # Check cache first
    cached_result = redis_client.get(f"asr:{file_hash}")
    if cached_result:
        return {"text": cached_result}
    
    # Get transcription
    result = asr_pipeline(
        audio_input,
        max_new_tokens=256,
        chunk_length_s=30,  # Process in 30-second chunks
        batch_size=8,
        return_timestamps=False
    )
    
    # Extract text from result
    if isinstance(result, dict) and "text" in result:
        text = result["text"].strip()
    else:
        text = result[0]["text"].strip() if result else ""
    
    # Cache the result
    redis_client.set(
        f"asr:{file_hash}",
        text,
        ex=3600  # Cache for 1 hour
    )
    
    return {"text": text} 