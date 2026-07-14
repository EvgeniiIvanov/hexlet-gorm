.PHONY: help
help:
	@echo "Docker & Database:"
	@echo "  docker-up          - start PostgreSQL with Docker Compose"
	@echo "  docker-down        - stop and remove PostgreSQL containers"
	@echo "  docker-clean       - remove containers and volumes (DESTROYS DATA)"

# Docker Compose command (try both docker-compose and docker compose)
DOCKER_COMPOSE := $(shell command -v docker-compose 2> /dev/null)
ifndef DOCKER_COMPOSE
	DOCKER_COMPOSE := docker compose
endif

# Docker commands
.PHONY: docker-up
docker-up:
	$(DOCKER_COMPOSE) --profile tools up -d
	@echo "Waiting for PostgreSQL to be ready..."
	@sleep 3
	@$(DOCKER_COMPOSE) ps

.PHONY: docker-down
docker-down:
	$(DOCKER_COMPOSE) down

.PHONY: docker-clean
docker-clean:
	@echo "WARNING: This will remove all containers and volumes (all data will be lost)"
	@read -p "Are you sure? [y/N] " -n 1 -r; \
	echo; \
	if [[ $$REPLY =~ ^[Yy]$$ ]]; then \
		$(DOCKER_COMPOSE) down -v; \
		echo "Containers and volumes removed."; \
	else \
		echo "Cancelled."; \
	fi