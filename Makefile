run:
	@go run cmd/radio/main.go

build:
	@go build -o radio cmd/radio/main.go

lint:
	@golangci-lint run

test:
	@go test -cover ./...
