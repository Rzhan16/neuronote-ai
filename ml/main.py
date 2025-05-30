from fastapi import FastAPI, File, UploadFile, HTTPException, Body, Form
from fastapi.middleware.cors import CORSMiddleware
from pydantic import BaseModel
from typing import List, Literal, Optional
import wave
import io
import uuid
import numpy as np
from ocr_service import extract_layout
from asr_service import transcribe
from summarise_service import summarise
from qg_service import generate_qa
from pipeline import process_note

app = FastAPI()

# Enable CORS
app.add_middleware(
    CORSMiddleware,
    allow_origins=["*"],
    allow_credentials=True,
    allow_methods=["*"],
    allow_headers=["*"],
)

# Response Models
class Block(BaseModel):
    text: str
    confidence: float
    bbox: List[float]

class OCRResponse(BaseModel):
    blocks: List[Block]

class ASRResponse(BaseModel):
    transcript: str

class SummaryRequest(BaseModel):
    text: str
    style: Literal["bullets", "paragraph"] = "paragraph"

class SummaryResponse(BaseModel):
    summary: str

class QARequest(BaseModel):
    text: str
    max_questions: int = 5

class QAPair(BaseModel):
    q: str
    a: str

class QAResponse(BaseModel):
    qa_pairs: List[QAPair]

class PipelineResponse(BaseModel):
    note_id: str

# Health Check
@app.get("/health")
async def health_check() -> dict:
    return {"ok": True}

# OCR Endpoint
@app.post("/ocr", response_model=OCRResponse)
async def ocr_endpoint(file: UploadFile = File(...)) -> OCRResponse:
    contents = await file.read()
    result = extract_layout(contents)
    return OCRResponse(blocks=[
        Block(text=block["text"], confidence=1.0, bbox=block["bbox"])
        for block in result["blocks"]
    ])

# ASR Endpoint
@app.post("/asr", response_model=ASRResponse)
async def asr_endpoint(file: UploadFile = File(...)) -> ASRResponse:
    contents = await file.read()
    
    # Convert audio file bytes to numpy array
    try:
        with wave.open(io.BytesIO(contents), 'rb') as wav_file:
            sample_rate = wav_file.getframerate()
            n_frames = wav_file.getnframes()
            audio_data = wav_file.readframes(n_frames)
            audio_array = np.frombuffer(audio_data, dtype=np.int16).astype(np.float32) / 32768.0
    except Exception as e:
        raise HTTPException(status_code=400, detail=f"Invalid WAV file: {str(e)}")
    
    result = transcribe(audio_array)
    return ASRResponse(transcript=result["text"])

# Summarization Endpoint
@app.post("/summarize", response_model=SummaryResponse)
async def summarize_endpoint(request: SummaryRequest) -> SummaryResponse:
    result = summarise(request.text, request.style)
    return SummaryResponse(summary=result["summary"])

# Question Generation Endpoint
@app.post("/generate-qa", response_model=QAResponse)
async def generate_qa_endpoint(request: QARequest) -> QAResponse:
    qa_pairs = generate_qa(request.text, request.max_questions)
    return QAResponse(qa_pairs=qa_pairs)

# Pipeline Endpoint
@app.post("/pipeline", response_model=PipelineResponse)
async def pipeline_endpoint(
    file: UploadFile = File(...),
    user_id: str = Form(default="anonymous")
) -> PipelineResponse:
    try:
        # Validate file type
        content_type = file.content_type
        if not any(t in content_type.lower() for t in ["pdf", "text", "image", "audio"]):
            raise HTTPException(
                status_code=400,
                detail=f"Unsupported file type: {content_type}. Supported types: PDF, text, images, audio."
            )

        # Read file content
        content = await file.read()
        if not content:
            raise HTTPException(
                status_code=400,
                detail="Empty file uploaded"
            )

        # Process the file
        try:
            result = await process_note(file)
            if not result or "note_id" not in result:
                raise ValueError("Processing failed to return a valid note ID")
            return PipelineResponse(**result)
        except Exception as e:
            raise HTTPException(
                status_code=500,
                detail=f"Failed to process file: {str(e)}"
            )

    except HTTPException:
        raise
    except Exception as e:
        raise HTTPException(
            status_code=500,
            detail=f"Unexpected error: {str(e)}"
        )
