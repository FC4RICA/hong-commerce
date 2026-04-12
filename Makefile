.PHONY: dev-up dev-down dev-fresh migrate-up

dev-up:
	docker compose -f compose.dev.yaml up --build

dev-down:
	docker compose -f compose.dev.yaml down

dev-fresh:
	docker compose -f compose.dev.yaml down -v
	docker compose -f compose.dev.yaml up --build

seed-admin:
	docker compose -f compose.dev.yaml run --rm \
		-e SEED_ADMIN_EMAIL=$(SEED_ADMIN_EMAIL) \
		-e SEED_ADMIN_PASSWORD=$(SEED_ADMIN_PASSWORD) \
		user-service \
		go run ./cmd/seed/main.go