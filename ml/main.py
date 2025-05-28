from fastapi import FastAPI, File, UploadFile, HTTPException, Body
from fastapi.middleware.cors import CORSMiddleware
from pydantic import BaseModel
from typing import List, Literal
import wave
import io
import numpy as np
from .ocr_service import extract_layout
from .asr_service import transcribe
from .summarise_service import summarise

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

# Pipeline Endpoint
@app.post("/pipeline", response_model=PipelineResponse)
async def pipeline_endpoint(file: UploadFile = File(...)) -> PipelineResponse:
    # TODO: Implement pipeline logic
    return PipelineResponse(note_id="sample-note-id")
