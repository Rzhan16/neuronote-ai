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

-- notes
id UUID PK
user_id UUID FK(users.id)
title TEXT
content TEXT
summary TEXT
created_at TIMESTAMPTZ DEFAULT now()
updated_at TIMESTAMPTZ DEFAULT now()

-- quiz_cards
id UUID PK
note_id UUID FK(notes.id)
question TEXT
answer TEXT
created_at TIMESTAMPTZ DEFAULT now()

-- tags
id UUID PK
note_id UUID FK(notes.id)
tag TEXT
created_at TIMESTAMPTZ DEFAULT now()

-- study_blocks
id UUID PK
note_id UUID FK(notes.id)
start_time TIMESTAMPTZ
end_time TIMESTAMPTZ
status TEXT
created_at TIMESTAMPTZ DEFAULT now()

-- sessions
id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
user_id UUID REFERENCES users(id),
token TEXT NOT NULL,
expires_at TIMESTAMPTZ NOT NULL,
created_at TIMESTAMPTZ DEFAULT NOW()
```

## API Endpoints (v0)

### Gateway Service (Port 8080)
- `GET /health` - Health check endpoint
  - Response: `{"ok": true}`
- `POST /api/notes/upload` - Upload and process a note
  - Input: `multipart/form-data` with file
  - Response: `{"note_id": string}`
  - Forwards file to ML service for processing
  - Supports both image and audio files
- `GET /api/notes/:id` - Get note details
  - Response: `{"id": string, "content": string, "summary": string, "quiz_cards": [{"id": string, "question": string, "answer": string}], "created_at": string, "updated_at": string}`
  - Returns note with generated quiz cards
- `GET /api/schedule` - Get study schedule
  - Response: `[{"id": string, "note_id": string, "start_time": string, "end_time": string, "status": string}]`
  - Returns upcoming study blocks
  - Ordered by start time

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
- `POST /summarize` - Generate text summary
  - Input: `{"text": string, "style": "bullets"|"paragraph"}`
  - Response: `{"summary": string}`
  - Uses BART-large-CNN model
    - Docker: 4-bit quantization (GPU)
    - Local: FP32 precision (CPU)
  - Supports paragraph and bullet-point formats
  - Caches results in Redis for 1 hour
- `POST /generate-qa` - Generate questions from text
  - Input: `{"text": string, "max_questions": int}`
  - Response: `{"qa_pairs": [{"q": string, "a": string}]}`
  - Uses T5-base model for question generation
  - Generates up to max_questions (default: 5) QA pairs
  - Caches results in Redis for 1 hour
- `POST /pipeline` - Process file through ML pipeline
  - Input: `multipart/form-data` with file
  - Response: `{"note_id": string}`
  - Detects file type (image/audio)
  - Extracts text via OCR/ASR
  - Generates summary
  - Extracts keyphrases using KeyBERT
  - Generates quiz cards
  - Persists all data in PostgreSQL

## Milestone Log

### M1: Initial Scaffold (2024-03-19) ✅
- ✅ Created monorepo structure with frontend/, gateway/, ml/, infra/, docs/
- ✅ Set up React + Vite + TypeScript in frontend/
- ✅ Added Python FastAPI dependencies in ml/
- ✅ Set up Go module with Fiber in gateway/
- ✅ Added polyglot .gitignore
- ✅ Created initial documentation

### M1.1: Docker Setup (2024-03-19) ✅
- ✅ Created Docker Compose configuration with five services:
  - Frontend (Node.js): Port 5173
  - Gateway (Go): Port 8080 with healthcheck
  - ML Service (Python): Port 8000 with healthcheck
  - PostgreSQL: With persistent volume
  - Redis: With persistent volume
- ✅ Set up service dependencies and networking
- ✅ Added health endpoints for gateway and ML services
- ✅ Configured development environment with hot-reload

### M2: ML Service Implementation (2024-03-19) ✅
- ✅ Created FastAPI app with health check
- ✅ Added CORS middleware
- ✅ Defined Pydantic models for responses
- ✅ Set up endpoint stubs for OCR, ASR, and pipeline
- ✅ Implemented Tesseract OCR with layout analysis
- ✅ Added OCR block extraction and confidence scoring
- ✅ Implemented Whisper-small model for transcription
- ✅ Added Redis caching with SHA256 hashing
- ✅ Implemented BART-large-CNN for summarization
- ✅ Added T5-base for question generation
- ✅ Added KeyBERT for keyphrase extraction
- ✅ Created unified pipeline with PostgreSQL persistence
- ✅ Added comprehensive test suite for all components
- ✅ Added proper error handling and validation
- ✅ Updated database schema with all required tables

### M3: Gateway API Implementation (2024-03-19) ✅
- ✅ Created Go Fiber app with health check
- ✅ Added middleware (Logger, CORS, RequestID)
- ✅ Implemented note upload endpoint with ML service integration
- ✅ Added note retrieval with quiz cards
- ✅ Added study schedule endpoint
- ✅ Set up PostgreSQL connection
- ✅ Added proper error handling and validation
- ✅ Updated API documentation

### Development Environment (2024-03-19)
- ✅ Go 1.24.3
- ✅ Node.js 23.11.0
- ✅ Python 3.11.12
- ✅ Fly CLI 0.3.132 (authenticated)
- ✅ Git (2.49.0)

## Milestone 3.1: Authentication & Authorization

### Features Added
- JWT-based authentication with secure cookie storage
- User signup and login endpoints
- Session management with database persistence
- Protected API routes with middleware
- Password hashing with bcrypt
- OpenAPI documentation updates

### Schema Changes
```sql
CREATE TABLE sessions (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID REFERENCES users(id),
    token TEXT NOT NULL,
    expires_at TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ DEFAULT NOW()
);
```

### Security Considerations
- JWT tokens expire after 30 days
- Secure, HTTP-only cookies
- Password hashing with bcrypt
- Session tracking for token revocation
- CORS configured for secure cookie handling
