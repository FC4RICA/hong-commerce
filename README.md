# Hong Commerce
A project demonstrating a simple ecommerce built with a microservice architecture.

## Architecture Overview
```
                  Frontend
                      |
                 API Gateway
                      |
       ┌──────────────┼────────────────┐
       ▼              ▼                ▼
Users Service   Catalog Service   Order Service
                                       |
                                       |
                              ┌────────┴────────┐
                              ▼                 ▼
                      Inventory Service   Payment Service
                              |
                              ▼
                        Message Queue
```

## Repository Structure
```
hong-commerce/
│
├── services/
│   ├── gateway/
│   ├── user-service/
│   ├── catalog-service/
│   ├── inventory-service/
│   ├── order-service/
│   └── payment-service/
│
├── pkg/                # shared Go packages
│
├── docker-compose.yml  # local development environment
│
└── README.md
```

## Docker & Local Development

This project uses **Docker** to run each service in its own isolated container. All services are orchestrated together using **Docker Compose**.

### Prerequisites
- [Docker](https://www.docker.com/) and [Docker Compose](https://docs.docker.com/compose/) installed
- `make` for easier setup
Optionally install `make` for easier setup

```bash
# Windows
choco install make
# Mac
brew install make
```

### Running the Project
```bash
make dev-up      # start all services
make dev-down    # stop all services
make dev-fresh   # reset everything (wipes db data)
```

Without Make
```bash
docker compose -f compose.dev.yaml up --build    # start
docker compose -f compose.dev.yaml down          # stop
docker compose -f compose.dev.yaml down -v       # reset (wipes db data)
```

Each service has its own `Dockerfile` within its directory. The root `docker-compose.yml` ties all services together, handling networking, environment variables, and dependencies between containers.

## Contributing

### Branching Strategy

Create a new branch for every change. Branch names follow this convention:
```
<service>/<type>/<short-description>
```

- **`<service>`** — the service being changed (e.g. `user-service`, `catalog-service`, `gateway`, `order-service`, `inventory-service`, `payment-service`). Use `shared` for changes to `pkg/` or project-wide config.
- **`<type>`** — the kind of change:
  | Type | When to use |
  |------|-------------|
  | `feature` | Adding new functionality |
  | `fix` | Bug fixes |
  | `refactor` | Code restructuring with no behavior change |
  | `docs` | Documentation only |
  | `test` | Adding or updating tests |
  | `chore` | Maintenance tasks (deps, config, CI) |
- **`<short-description>`** — a brief kebab-case summary of the change

**Examples:**
```
user-service/feature/jwt-authentication
order-service/refactor/checkout-flow
shared/chore/update-go-dependencies
gateway/docs/api-endpoint-descriptions
```

### Pull Requests

1. **Rebase onto `main` before opening a PR** — never merge `main` into your branch.
```bash
   git fetch origin
   git rebase origin/main
```
   Resolve any conflicts, then force-push your branch:
```bash
   git push --force-with-lease origin user-service/feature/jwt-authentication
```
2. **Fill out the PR description** with:
   - What changed and why
   - Related task link  
3. **Merge method — Rebase or Squash only.** When merging, always select **"Rebase and merge"** for small PR or **"Squash and merge"** for PR with multiple commits (not "Create a merge commit"). This keeps the commit history linear and readable.

### General Workflow
```bash
# 1. Sync with main
git checkout main
git pull origin main

# 2. Create your branch
git checkout -b user-service/feature/jwt-authentication

# 3. Make your changes, then commit
git add .
git commit -m "user-service: add JWT authentication middleware"

# 4. Rebase onto latest main before pushing
git fetch origin
git rebase origin/main

# 5. Push and open a PR
git push origin user-service/feature/jwt-authentication
# If you've already pushed before and rebased, use:
git push --force-with-lease origin user-service/feature/jwt-authentication
```

> **Commit messages** should follow the same `<service>: <short description>` pattern to keep the git history readable.
