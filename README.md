# NeuroNote AI

Multimodal note-summariser & adaptive study-planner.

## Project Structure

```
neuronote-ai/
├── frontend/    # React + Vite + TypeScript
├── gateway/     # Go API Gateway (Fiber)
├── ml/          # Python ML Service (FastAPI)
├── infra/       # Infrastructure (Docker, Fly)
└── docs/        # Documentation
```

## Prerequisites

- Node.js 16+
- Go 1.20+ (not installed yet)
- Python 3.9+
- Docker

## Quick Start

1. Frontend (React):
   ```bash
   cd frontend
   npm install
   npm run dev
   ```

2. ML Service (FastAPI):
   ```bash
   cd ml
   python -m venv venv
   source venv/bin/activate  # or `venv\Scripts\activate` on Windows
   pip install -r requirements.txt
   ```

3. Gateway (Go) - Coming soon:
   ```bash
   cd gateway
   go mod download
   go run main.go
   ```

## Development

See `neuronote-ai.md` for detailed specifications and development guidelines. 