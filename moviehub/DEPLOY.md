# MovieHub Week 11–12: deploy checklist

## Repository locations
- **Go API:** `Week6-lab/moviehub/` (NOT the older `kitchen/` lab)
- **React:** `Moviehub-670710751/` on branch `week11-12-development`

## Supabase (requires your own account)
1. Create a PostgreSQL project.
2. In SQL Editor, execute `schema.sql` **once** (creates 8 tables and genres).
3. Execute `seed.sql` to load the movie records.
4. Execute `rls.sql` to enable RLS on all eight tables.
5. Copy the **session-pooler** PostgreSQL URI (port 5432); ensure `sslmode=require`. Do not put this URI in React, GitHub, screenshots or chat.

## Render Go Web Service
- Repo: `670710751-Boonyai/Week6-lab`
- Root Directory: `moviehub`
- Build Command: `go build -o server .`
- Start Command: `./server`
- Set environment variables in Render (not in repository):
  - `DATABASE_URL`: Supabase session pooler connection URI (add `sslmode=require`)
  - `JWT_SECRET`: a random string of at least 32 characters
  - `ALLOWED_ORIGINS`: the exact Vercel frontend URL, e.g. `https://example.vercel.app`
- Render supplies `PORT`; API uses port 8080 when unset.
- Health path: `/health` or `/api/movies`.

## Vercel React
- Repo: `670710751-Boonyai/Moviehub-670710751`
- Framework: Create React App
- Set `REACT_APP_API_URL=https://YOUR-RENDER-API.onrender.com`
- Redeploy after setting env vars and update Render `ALLOWED_ORIGINS` to the actual Vercel domain.
- For SPA client routes, `vercel.json` rewrites routes to `index.html`.

## Create the first admin account
Register normally in React, then use Supabase SQL Editor:
```sql
UPDATE members SET role = 'admin'
WHERE lower(email) = lower('YOUR_REGISTERED_EMAIL');
```
Log out and log back in to get a new JWT carrying the admin role.

## Verify
- Public: `GET /api/movies`, `GET /api/movies/{id}`, `GET /api/genres`.
- Member: Register, login, `GET /api/me`, reviews, voting, wishlist.
- Admin: `GET /api/admin/stats`, list members, role change, add/edit/delete movie.
- Non-admin request to `/api/admin/*` must return HTTP 403.
- Never commit `.jwt_secret`, `.env` or `DATABASE_URL`.

## Limitations
The repository is ready for code review, but end-to-end tests require a running Supabase database and the deployed services. GitHub Actions build checks are provided; green checks should be verified before hand-in.
