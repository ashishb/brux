BINARY_NAME := "brux"

format:
	go fmt ./...
	go mod tidy
	go fix ./...
	golangci-lint run --concurrency=1 --fix

lint:
	go mod tidy
	go vet ./...
	golangci-lint run

test:
	go test ./... -v

build:
	go build -v -o bin/${BINARY_NAME} ./cmd/brux
