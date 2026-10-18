# None of these produce a file of their own name. Without this, a directory or
# file that happens to match one, like `demo`, makes the target "up to date" and
# the recipe silently stops running.
.PHONY: up up-demo down-demo clean test test-backend test-frontend

# The demo is its own compose project: its own database volume, env file and config.
# Container names are fixed in docker-compose.yaml, so stop one stack before starting the other.
DEMO := docker compose -p bruttice-demo --project-directory . \
	--env-file demo/env -f docker-compose.yaml -f demo/docker-compose.yaml

# The stack from your own .env and apps/backend/config.yaml (copy the .example files first).
up:
	docker compose up --build

# The stack with example venues and events, on http://localhost:8080.
# Sign in at /admin as admin@example.com / demo-password. The data persists
# between runs, use `make clean` to start over.
up-demo:
	$(DEMO) up --build

down-demo:
	$(DEMO) down --remove-orphans

# Only the project name matters for `down`, so .env.example stands in when there is no .env.
LOCAL_ENV := $(if $(wildcard .env),.env,.env.example)

# Removes the containers and database volumes of the demo and of the local stack.
# Only projects of this repo are touched, nothing else in Docker is.
clean:
	$(DEMO) down -v --remove-orphans
	docker compose --env-file $(LOCAL_ENV) down -v --remove-orphans

test: test-backend test-frontend

# No database or containers needed.
test-backend:
	cd apps/backend && go test ./...
	cd packages/go-pkg && go test ./...

test-frontend:
	cd apps/frontend && pnpm install --frozen-lockfile && pnpm test
