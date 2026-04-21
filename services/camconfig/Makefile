COMPOSE_FILE=docker-compose.yml
GINKGO=cd ../../tests && ginkgo -v integration/camconfig
CAMCONFIG_PID=/tmp/camconfig-test.pid
POSTGRES_CONTAINER=camconfig-test-postgres
POSTGRES_PORT=5433

.PHONY: run
run:
	docker compose -f $(COMPOSE_FILE) --profile inmemory up --build

.PHONY: run-detach
run-detach:
	docker compose -f $(COMPOSE_FILE) --profile inmemory up --build -d

.PHONY: run-filestore-docker
run-filestore-docker:
	docker compose -f $(COMPOSE_FILE) --profile filestore up --build

.PHONY: run-postgres-docker
run-postgres-docker:
	docker compose -f $(COMPOSE_FILE) --profile postgres up --build

.PHONY: stop
stop:
	docker compose -f $(COMPOSE_FILE) down

.PHONY: logs
logs:
	docker compose -f $(COMPOSE_FILE) logs -f

.PHONY: build
build:
	docker compose -f $(COMPOSE_FILE) build

.PHONY: run-local
run-local:
	go run ./cmd/camconfig

.PHONY: run-filestore
run-filestore:
	DB_URI=file://./data/camconfig.json go run ./cmd/camconfig

.PHONY: build-local
build-local:
	go build -o ./camconfig ./cmd/camconfig

# Local integration tests

.PHONY: test-integration
test-integration: test-integration-inmemory

.PHONY: test-integration-inmemory
test-integration-inmemory:
	@echo "=== Running integration tests (inmemory) ==="
	DB_URI=inmemory:// go run ./cmd/camconfig &> /tmp/camconfig-test.log & echo $$! > $(CAMCONFIG_PID); \
	for i in $$(seq 1 30); do curl -s http://localhost:8081/v1/health > /dev/null 2>&1 && break; sleep 1; done; \
	CAMCONFIG_ENDPOINT=localhost:8081 $(GINKGO); \
	TEST_EXIT=$$?; \
	kill `cat $(CAMCONFIG_PID)` 2>/dev/null || true; \
	rm -f $(CAMCONFIG_PID); \
	exit $$TEST_EXIT

.PHONY: test-integration-filestore
test-integration-filestore:
	@echo "=== Running integration tests (filestore) ==="
	rm -f ./data/camconfig.json; \
	DB_URI=file://./data/camconfig.json go run ./cmd/camconfig &> /tmp/camconfig-test.log & echo $$! > $(CAMCONFIG_PID); \
	for i in $$(seq 1 30); do curl -s http://localhost:8081/v1/health > /dev/null 2>&1 && break; sleep 1; done; \
	CAMCONFIG_ENDPOINT=localhost:8081 $(GINKGO); \
	TEST_EXIT=$$?; \
	kill `cat $(CAMCONFIG_PID)` 2>/dev/null || true; \
	rm -f $(CAMCONFIG_PID); \
	rm -f ./data/camconfig.json; \
	exit $$TEST_EXIT

.PHONY: test-integration-postgres
test-integration-postgres:
	@echo "=== Running integration tests (postgres) ==="
	@docker rm -f $(POSTGRES_CONTAINER) 2>/dev/null || true; \
	docker run -d --name $(POSTGRES_CONTAINER) -e POSTGRES_PASSWORD=t00r -e POSTGRES_DB=eventsourcing -p $(POSTGRES_PORT):5432 postgres:latest; \
	echo "Waiting for postgres to be ready..."; \
	for i in $$(seq 1 30); do \
		docker exec $(POSTGRES_CONTAINER) pg_isready -U postgres > /dev/null 2>&1 && break; \
		sleep 1; \
	done; \
	DB_URI="host=localhost port=$(POSTGRES_PORT) user=postgres dbname=eventsourcing password=t00r sslmode=disable" go run ./cmd/camconfig &> /tmp/camconfig-test.log & echo $$! > $(CAMCONFIG_PID); \
	for i in $$(seq 1 30); do curl -s http://localhost:8081/v1/health > /dev/null 2>&1 && break; sleep 1; done; \
	CAMCONFIG_ENDPOINT=localhost:8081 $(GINKGO); \
	TEST_EXIT=$$?; \
	kill `cat $(CAMCONFIG_PID)` 2>/dev/null || true; \
	rm -f $(CAMCONFIG_PID); \
	docker stop $(POSTGRES_CONTAINER) && docker rm $(POSTGRES_CONTAINER); \
	exit $$TEST_EXIT

# Docker CI integration tests

.PHONY: test-docker-inmemory
test-docker-inmemory:
	docker compose -f $(COMPOSE_FILE) --profile test-inmemory up --build --abort-on-container-exit; \
	docker compose -f $(COMPOSE_FILE) --profile test-inmemory down

.PHONY: test-docker-filestore
test-docker-filestore:
	rm -f ./data/camconfig.json; \
	docker compose -f $(COMPOSE_FILE) --profile test-filestore up --build --abort-on-container-exit; \
	docker compose -f $(COMPOSE_FILE) --profile test-filestore down

.PHONY: test-docker-postgres
test-docker-postgres:
	docker compose -f $(COMPOSE_FILE) --profile test-postgres up --build --abort-on-container-exit; \
	docker compose -f $(COMPOSE_FILE) --profile test-postgres down