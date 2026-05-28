.PHONY: dev-up dev-down dev-fresh migrate-up

ifneq (,$(wildcard ./.env))
    include .env
    export
endif

dev:
	docker compose -f compose.dev.yaml up --build

dev-fresh:
	docker compose -f compose.dev.yaml down -v
	docker compose -f compose.dev.yaml up --build

seed-admin:
	docker compose -f compose.dev.yaml run --rm \
		-e SEED_ADMIN_EMAIL=$(SEED_ADMIN_EMAIL) \
		-e SEED_ADMIN_PASSWORD=$(SEED_ADMIN_PASSWORD) \
		user-service \
		go run ./cmd/seed/main.go

seed-inventory:
	docker compose -f compose.dev.yaml exec inventory-service bun run db:seed

seed: seed-admin seed-inventory