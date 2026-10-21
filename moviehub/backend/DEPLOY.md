# MovieHub Week 11–12 deployment checklist

## Supabase
1. Create a PostgreSQL project. Run ../schema.sql then ../seed.sql using SQL Editor.
2. Run rls.sql to enable RLS on all 8 tables.
3. Use the Session Pooler PostgreSQL connection string, port 5432, sslmode=require. Keep DATABASE_URL secret.
4. Do not put PostgreSQL credentials in React.

## Render Go web service
- Repository: Week6-lab, branch: moviehub-go-week11-12
- Root directory: moviehub/backend
- Build: go mod tidy && go build -o server .
- Start: ./server
- Environment:
  - DATABASE_URL=<Supabase session pooler URI with sslmode=require>
  - JWT_SECRET=<random secret at least 32 characters>
  - ALLOWED_ORIGINS=https://<your-vercel-domain>
- Check GET /health and GET /api/movies.

## Vercel React
- Repository: Moviehub-670710751, branch: week11-12-development
- Framework: Create React App
- Build: npm run build
- Output: build
- Environment: REACT_APP_API_URL=https://<your-render-domain>
- Add SPA fallback rewrite to index.html.
- Verify register/login, reviews, wishlist and admin with a role-enabled account.

## Initial admin account
Registration always creates role=member. Promote a trusted account in Supabase SQL Editor:
```sql
UPDATE members SET role='admin' WHERE lower(email)=lower('YOUR_ADMIN_EMAIL');
```
Log out and log in again to refresh the JWT role.

## Security and testing
- Verify GET /api/admin/stats returns 401 without token, 403 with member token, and 200 with admin token.
- Verify admin cannot change their own role.
- Verify CORS origin matches the actual Vercel deployment.
- No credentials or real .env files belong in Git.
- This deployment has not been executed or independently verified.
