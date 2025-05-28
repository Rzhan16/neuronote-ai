from fastapi import FastAPI, File, UploadFile
from fastapi.middleware.cors import CORSMiddleware
from pydantic import BaseModel
from typing import List
from .ocr_service import extract_layout

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
    # TODO: Implement ASR logic
    return ASRResponse(transcript="Sample transcript")

# Pipeline Endpoint
@app.post("/pipeline", response_model=PipelineResponse)
async def pipeline_endpoint(file: UploadFile = File(...)) -> PipelineResponse:
    # TODO: Implement pipeline logic
    return PipelineResponse(note_id="sample-note-id")
