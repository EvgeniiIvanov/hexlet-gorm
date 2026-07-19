.PHONY: help
help:
	@echo "Docker & Database:"
	@echo "  docker-up          - start PostgreSQL with Docker Compose"
	@echo "  docker-down        - stop and remove PostgreSQL containers"
	@echo "  docker-clean       - remove containers, volumes, and networks (DESTROYS DATA)"
	@echo "  docker-clean-force - same as docker-clean but without confirmation"
	@echo ""
	@echo "Testing:"
	@echo "  test               - run all tests"
	@echo "  test-verbose       - run all tests with verbose output"
	@echo "  test-cover         - run tests with coverage report"
	@echo "  test-n-plus-one    - check for N+1 query problems"
	@echo "  bench              - run performance benchmarks"
	@echo "  bench-compare      - compare baseline vs optimized benchmarks"

# Docker Compose command (try both docker-compose and docker compose)
DOCKER_COMPOSE := $(shell command -v docker-compose 2> /dev/null)
ifndef DOCKER_COMPOSE
	DOCKER_COMPOSE := docker compose
endif

# Docker commands
.PHONY: docker-up
docker-up:
	$(DOCKER_COMPOSE) --profile tools up -d || $(DOCKER_COMPOSE) --profile tools up -d
	@echo "Waiting for PostgreSQL to be ready..."
	@sleep 3
	@$(DOCKER_COMPOSE) ps

.PHONY: docker-down
docker-down:
	$(DOCKER_COMPOSE) down

.PHONY: docker-clean
docker-clean:
	@echo "WARNING: This will remove all containers, volumes, networks, and images (all data will be lost)"
	@printf "Are you sure? [y/N] "; \
	read -r reply; \
	if [ "$$reply" = "y" ] || [ "$$reply" = "Y" ]; then \
		$(DOCKER_COMPOSE) down -v --remove-orphans --rmi local; \
		echo "All resources cleaned up."; \
	else \
		echo "Cancelled."; \
	fi

.PHONY: docker-clean-force
docker-clean-force:
	@echo "Removing all containers, volumes, networks, and images..."
	$(DOCKER_COMPOSE) down -v --remove-orphans --rmi local
	@echo "All resources cleaned up."

# Test commands
.PHONY: test
test:
	go test ./internal/requests/...

.PHONY: test-verbose
test-verbose:
	go test -v ./internal/requests/...

.PHONY: test-cover
test-cover:
	go test -cover ./internal/requests/...
	@echo ""
	@echo "For detailed HTML coverage report, run:"
	@echo "  go test -coverprofile=coverage.out ./internal/requests/..."
	@echo "  go tool cover -html=coverage.out"

.PHONY: test-n-plus-one
test-n-plus-one:
	go test -v -run TestNPlusOneQueries ./internal/requests/

.PHONY: bench
bench:
	go test -bench=. -benchmem ./internal/requests/

.PHONY: bench-compare
bench-compare:
	@echo "Running baseline benchmarks..."
	go test -bench=. -benchmem -benchtime=100x ./internal/requests/ > baseline_results.txt
	@echo "Baseline saved to baseline_results.txt"
	@echo ""
	@echo "To compare after optimization:"
	@echo "  go test -bench=. -benchmem -benchtime=100x ./internal/requests/ > optimized_results.txt"
	@echo "  benchstat baseline_results.txt optimized_results.txt"