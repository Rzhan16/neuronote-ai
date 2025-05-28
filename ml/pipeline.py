import uuid
import magic
import psycopg2
from typing import Dict, List, Optional
from fastapi import UploadFile
from keybert import KeyBERT
from .ocr_service import extract_layout
from .asr_service import transcribe
from .summarise_service import summarise
from .qg_service import generate_qa

# Initialize KeyBERT model
keybert_model = KeyBERT()

# Database connection
DB_CONFIG = {
    'dbname': 'neuronote',
    'user': 'postgres',
    'password': 'postgres',
    'host': 'postgres',
    'port': '5432'
}

def get_db_connection():
    """Get a database connection."""
    return psycopg2.connect(**DB_CONFIG)

def detect_mimetype(file_content: bytes) -> str:
    """Detect file MIME type."""
    return magic.from_buffer(file_content, mime=True)

def extract_keyphrases(text: str, top_n: int = 5) -> List[str]:
    """Extract key phrases from text using KeyBERT."""
    keywords = keybert_model.extract_keywords(
        text,
        keyphrase_ngram_range=(1, 2),
        stop_words='english',
        top_n=top_n
    )
    return [keyword for keyword, _ in keywords]

async def process_note(file: UploadFile) -> Dict[str, str]:
    """Process a note through the ML pipeline.
    
    Args:
        file: Uploaded file (image or audio)
        
    Returns:
        Dict with note_id and processing status
    """
    # Generate unique ID for the note
    note_id = str(uuid.uuid4())
    
    # Read file content
    content = await file.read()
    
    # Detect file type
    mimetype = detect_mimetype(content)
    
    # Process based on file type
    if mimetype.startswith('image/'):
        # OCR processing
        ocr_result = extract_layout(content)
        text = ' '.join(block['text'] for block in ocr_result['blocks'])
        blocks = [
            (block['text'], block['bbox'])
            for block in ocr_result['blocks']
        ]
    elif mimetype.startswith('audio/'):
        # ASR processing
        audio_result = transcribe(content)
        text = audio_result['text']
        blocks = None
    else:
        raise ValueError(f"Unsupported file type: {mimetype}")
    
    # Generate summary
    summary = summarise(text)['summary']
    
    # Extract keyphrases
    tags = extract_keyphrases(text)
    
    # Generate quiz cards
    qa_pairs = generate_qa(text)
    
    # Store results in database
    with get_db_connection() as conn:
        with conn.cursor() as cur:
            # Insert note
            cur.execute("""
                INSERT INTO notes (id, content, summary)
                VALUES (%s, %s, %s)
            """, (note_id, text, summary))
            
            # Insert OCR blocks if present
            if blocks:
                for block_text, bbox in blocks:
                    cur.execute("""
                        INSERT INTO ocr_blocks (note_id, text, bbox)
                        VALUES (%s, %s, %s)
                    """, (note_id, block_text, {'coords': bbox}))
            
            # Insert quiz cards
            for qa in qa_pairs:
                cur.execute("""
                    INSERT INTO quiz_cards (note_id, question, answer)
                    VALUES (%s, %s, %s)
                """, (note_id, qa['q'], qa['a']))
            
            # Insert tags
            for tag in tags:
                cur.execute("""
                    INSERT INTO tags (note_id, tag)
                    VALUES (%s, %s)
                """, (note_id, tag))
        
        conn.commit()
    
    return {"note_id": note_id} 