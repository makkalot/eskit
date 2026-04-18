COMPOSE_FILE=docker-compose.yml

.PHONY: run
run:
	docker compose -f $(COMPOSE_FILE) --profile inmemory up --build

.PHONY: run-detach
run-detach:
	docker compose -f $(COMPOSE_FILE) --profile inmemory up --build -d

.PHONY: run-filestore-docker
run-filestore-docker:
	docker compose -f $(COMPOSE_FILE) --profile filestore up --build

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