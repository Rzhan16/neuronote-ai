# NeuroNote AI — Living Spec (v0)

Multimodal note-summariser & adaptive study-planner.

| Layer | Tech                     |
|-------|--------------------------|
| FE    | React + Vite + Tailwind  |
| API   | Go (Fiber)               |
| ML    | FastAPI + Hugging Face   |
| Data  | Postgres, Redis          |

## DB schema (v0)
```sql
-- users
id UUID PK
email TEXT UNIQUE
hashed_password TEXT
created_at TIMESTAMPTZ

-- ocr_blocks
id UUID PK
note_id UUID FK(notes.id)
text TEXT
bbox JSONB  -- [x1, y1, x2, y2]
created_at TIMESTAMPTZ DEFAULT now()

-- audio_notes
id UUID PK
note_id UUID FK(notes.id)
transcript TEXT
created_at TIMESTAMPTZ DEFAULT now()
```

## API Endpoints (v0)

### ML Service (Port 8000)
- `GET /health` - Health check endpoint
  - Response: `{"ok": true}`
- `POST /ocr` - Extract text from images
  - Input: `multipart/form-data` with file
  - Response: `{"blocks": [{"text": string, "confidence": float, "bbox": [float]}]}`
- `POST /asr` - Convert speech to text
  - Input: `multipart/form-data` with WAV file
  - Response: `{"transcript": string}`
  - Caches results in Redis for 1 hour
  - Uses Whisper-small model for transcription
- `POST /pipeline` - Process file through ML pipeline
  - Input: `multipart/form-data` with file
  - Response: `{"note_id": string}`

## Milestone Log

### M1: Initial Scaffold (2024-03-19)
- ✅ Created monorepo structure with frontend/, gateway/, ml/, infra/, docs/
- ✅ Set up React + Vite + TypeScript in frontend/
- ✅ Added Python FastAPI dependencies in ml/
- ✅ Set up Go module with Fiber in gateway/
- ✅ Added polyglot .gitignore
- ✅ Created initial documentation

### M1.1: Docker Setup (2024-03-19)
- ✅ Created Docker Compose configuration with five services:
  - Frontend (Node.js): Port 5173
  - Gateway (Go): Port 8080 with healthcheck
  - ML Service (Python): Port 8000 with healthcheck
  - PostgreSQL: With persistent volume
  - Redis: With persistent volume
- ✅ Set up service dependencies and networking
- ✅ Added health endpoints for gateway and ML services
- ✅ Configured development environment with hot-reload

### M2: FastAPI Base Implementation (2024-03-19)
- ✅ Created FastAPI app with health check
- ✅ Added CORS middleware
- ✅ Defined Pydantic models for responses
- ✅ Set up endpoint stubs for OCR, ASR, and pipeline

### M2.1: OCR Implementation (2024-03-19)
- ✅ Implemented Tesseract OCR with layout analysis
- ✅ Added OCR block extraction and confidence scoring
- ✅ Created test infrastructure with sample image generation
- ✅ Added ocr_blocks table to DB schema

### M2.2: ASR Implementation (2024-03-19)
- ✅ Implemented Whisper-small model for transcription
- ✅ Added Redis caching with SHA256 hashing
- ✅ Created test infrastructure with sample audio generation
- ✅ Added audio_notes table to DB schema
- ✅ Added proper error handling for invalid audio files

### Development Environment (2024-03-19)
- ✅ Go 1.24.3
- ✅ Node.js 23.11.0
- ✅ Python 3.11.12
- ✅ Fly CLI 0.3.132 (authenticated)
- ✅ Git (2.49.0)
