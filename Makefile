APP=eth2usd

.PHONY: build run lint tidy clean

build:
	go build -o bin/$(APP) ./cmd/eth2usd

run:
	go run ./cmd/eth2usd --format=text --account $$ACCOUNT --chainlink-registry $$REGISTRY --rpc-url $$RPC

lint:
	golangci-lint run

tidy:
	go mod tidy

clean:
	rm -rf bin
