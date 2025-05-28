from fastapi import FastAPI, File, UploadFile
from fastapi.middleware.cors import CORSMiddleware
from pydantic import BaseModel
from typing import List

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
    # TODO: Implement OCR logic
    return OCRResponse(blocks=[
        Block(text="Sample text", confidence=0.95, bbox=[0, 0, 100, 100])
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
