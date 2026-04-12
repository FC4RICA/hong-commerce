# user-service
 
Handles user registration, authentication, and profile retrieval. Runs on port `8081`, accessed through the gateway at `8080`.
 
## Routes
 
| Method | Path | Auth | Description |
|--------|------|------|-------------|
| `POST` | `/register` | Public | Register a new user |
| `POST` | `/login` | Public | Login and receive a JWT |
| `POST` | `/admin/register` | Admin | Register a new admin user |
| `GET` | `/me` | Gateway JWT | Get current user profile |
| `GET` | `/health` | Public | Health check |
 
## Environment Variables
 
| Variable | Required | Default | Description |
|----------|----------|---------|-------------|
| `PORT` | No | `8081` | Port to listen on |
| `DATABASE_URL` | Yes | — | Postgres connection string |
| `JWT_SECRET` | Yes | — | Shared secret with gateway |
| `SEED_ADMIN_EMAIL` | No | — | Seeded admin email |
| `SEED_ADMIN_PASSWORD` | No | — | Seeded admin password |