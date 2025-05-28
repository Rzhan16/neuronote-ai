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

## Milestone Log

### M1: Initial Scaffold (2024-03-19)
- ✅ Created monorepo structure with frontend/, gateway/, ml/, infra/, docs/
- ✅ Set up React + Vite + TypeScript in frontend/
- ✅ Added Python FastAPI dependencies in ml/
- ✅ Set up Go module with Fiber in gateway/
- ✅ Added polyglot .gitignore
- ✅ Created initial documentation

### Development Environment (2024-03-19)
- ✅ Go 1.24.3
- ✅ Node.js 23.11.0
- ✅ Python 3.11.12
- ✅ Fly CLI 0.3.132 (authenticated)
- ✅ Git (2.49.0)
