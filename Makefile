.PHONY: test
test:
	go test -fullpath -race -cover -bench=. ./...

lint-local:
	golangci-lint run

infra-start:
	docker compose -p outbox up -d

infra-stop:
	docker compose -p outbox stop
