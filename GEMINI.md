# Hong Commerce: Developer Guide & Conventions

This document serves as the primary source of truth for development practices, architectural patterns, and workflows within the `hong-commerce` repository.

## Project Overview

`hong-commerce` is a microservices-based ecommerce platform implemented in Go. It utilizes an API Gateway pattern for unified entry and security.

### Core Architecture
- **API Gateway (`/gateway`):** Entry point for all external requests. Handles authentication (JWT), logging, request tracing, and routing to downstream services.
- **Microservices:**
  - `user-service`: Manages user accounts, authentication, and profiles.
  - `catalog-service`: Product catalog management.
  - `order-service`: Order processing and management.
  - `inventory-service`: Stock tracking.
  - `payment-service`: Payment processing.
- **Communication:** Services communicate over HTTP. The Gateway acts as a reverse proxy.
- **Security:** Gateway validates JWTs and injects `X-User-ID` and `X-User-Role` headers into requests forwarded to downstream services. Internal services trust these headers.

### Technology Stack
- **Language:** Go (1.21+)
- **Routing:** `go-chi/chi`
- **Database:** PostgreSQL (using `pgx/v5`)
- **Logging:** `uber-go/zap` (Gateway), standard `log` (Services)
- **Containerization:** Docker & Docker Compose
- **Orchestration:** `Makefile` for common tasks

## Getting Started

### Prerequisites
- Docker & Docker Compose
- `make` (optional but recommended)

### Common Commands
```bash
make dev           # Start all services with hot-reload (using Air)
make dev-fresh     # Stop, wipe volumes, and restart all services
make seed-admin    # Seed an admin user (requires SEED_ADMIN_EMAIL/PASSWORD env)
```

## Development Conventions

### Branching Strategy
Branch names MUST follow the pattern: `<service>/<type>/<short-description>`

- **`<service>`**: `gateway`, `user-service`, `catalog-service`, `order-service`, `inventory-service`, `payment-service`, or `shared`.
- **`<type>`**: `feature`, `fix`, `refactor`, `docs`, `test`, `chore`.
- **Example**: `user-service/feature/jwt-authentication`

### Commit Messages
Follow the pattern: `<service>: <short description>`
- **Example**: `gateway: add request-id middleware`

### Workflow & PRs
1. **Rebase Always:** Never merge `main` into your branch. Use `git rebase origin/main`.
2. **Pull Requests:** 
   - Use **"Rebase and merge"** for single-commit or small PRs.
   - Use **"Squash and merge"** for multi-commit PRs to keep history clean.
3. **Linear History:** Maintain a linear git history. Avoid merge commits.

### Coding Standards (Go)
- **Error Handling:** Use `errors.Is` and `errors.As` for error checking.
- **Routing:** Use `chi` routers. Group routes logically (e.g., public vs. protected).
- **Service Pattern:** Follow the `Handler -> Service -> Repository` layer pattern seen in `user-service`.
- **Context:** Always pass `context.Context` through layers for cancellation and timeouts.
- **Dependencies:** Use Go modules. Keep `vendor/` directory if present (not currently used).

### API Gateway Integration
- Downstream services should not re-verify JWTs.
- Rely on headers injected by the Gateway:
  - `X-User-ID`: The unique ID of the authenticated user.
  - `X-User-Role`: The role of the user (e.g., `admin`, `user`).
- Routes requiring authentication should be placed under the `middleware.Auth` group in `gateway/router/router.go`.

## Testing Strategy
- **Unit Tests:** Place next to the code (e.g., `service_test.go`).
- **Integration Tests:** (TODO: Define standard for cross-service testing).
- Use `testify` for assertions if available in the module.
