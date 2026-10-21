# MovieHub Go Backend — Week 11 foundation

This is an **initial implementation**, not the completed Week 11–12 assignment.

Implemented:
- Go HTTP service and Render-compatible PORT
- PostgreSQL connection via DATABASE_URL (pgxpool)
- GET /health and GET /api/movies
- CORS allowlist via ALLOWED_ORIGINS

Not yet implemented: register/login, JWT, member/review/vote/wishlist APIs, movie detail, and Week 12 role-protected admin APIs.

## Local run
Use PostgreSQL 16. Load `../schema.sql` and `../seed.sql` into a fresh database, then set:
- DATABASE_URL (keep secret; never commit)
- ALLOWED_ORIGINS=http://localhost:3000
- PORT=8080

Run `go mod tidy` then `go run .` from this directory.

For Supabase follow Week 11 instructions: Session pooler port 5432, sslmode=require, enable RLS for all 8 tables before exposing the project.

Do not deploy as a finished system until auth and access controls are implemented and tested.
