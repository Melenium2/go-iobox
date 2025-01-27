.PHONY: test
test:
	go test -fullpath -race -cover -bench=. ./...

lint-local:
	golangci-lint run

infra-start:
	cd example && docker compose -p outbox up -d

infra-stop:
	cd example && docker compose -p outbox stop
