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
```

## API Endpoints (v0)

### ML Service (Port 8000)
- `GET /health` - Health check endpoint
  - Response: `{"ok": true}`
- `POST /ocr` - Extract text from images
  - Input: `multipart/form-data` with file
  - Response: `{"blocks": [{"text": string, "confidence": float, "bbox": [float]}]}`
- `POST /asr` - Convert speech to text
  - Input: `multipart/form-data` with file
  - Response: `{"transcript": string}`
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

### Development Environment (2024-03-19)
- ✅ Go 1.24.3
- ✅ Node.js 23.11.0
- ✅ Python 3.11.12
- ✅ Fly CLI 0.3.132 (authenticated)
- ✅ Git (2.49.0)
