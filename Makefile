.PHONY: run-memory run-memory-d run-postgres run-postgres-d stop test

run-memory:
	docker compose --profile memory up --build

run-memory-d:
	docker compose --profile memory up --build -d

run-postgres:
	docker compose --profile postgres up --build

run-postgres-d:
	docker compose --profile postgres up --build -d

stop:
	docker compose --profile memory --profile postgres down

test:
	go test ./...
