.PHONY: stop run test lint build

run:
	go run cmd/server/main.go

test:
	go test ./... -v -race -cover

lint:
	golangci-lint run ./...

build:
	go build -o bin/server cmd/server/main.go

stop:
	@kill $$(lsof -t -i:8082) 2>/dev/null || true