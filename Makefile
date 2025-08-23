.PHONY: proto
proto:
	buf generate api

.PHONY: deps
deps:
	go mod download
	go mod tidy

.PHONY: build
build: proto
	go build -o bin/user-service ./cmd/user-service
	go build -o bin/order-service ./cmd/order-service
	go build -o bin/gateway-service ./cmd/gateway-service

.PHONY: run/user-service
run/user-service:
	go run ./cmd/user-service

.PHONY: run/order-service
run/order-service:
	go run ./cmd/order-service

.PHONY: run/gateway-service
run/gateway-service:
	go run ./cmd/gateway-service

.PHONY: test
test:
	go test -v ./...

.PHONY: lint
lint:
	golangci-lint run

.PHONY: migrate/up
migrate/up:
	migrate -path migrations -database "$${DATABASE_URL}" up

.PHONY: migrate/down
migrate/down:
	migrate -path migrations -database "$${DATABASE_URL}" down

.PHONY: migrate/create
migrate/create:
	migrate create -ext sql -dir migrations -seq $(name)

.PHONY: sqlc
sqlc:
	sqlc generate

.PHONY: clean
clean:
	rm -rf bin/ gen/

.PHONY: install-tools
install-tools:
	go install github.com/bufbuild/buf/cmd/buf@latest
	go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest
	go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest

.PHONY: dev/setup
dev/setup: install-tools deps proto sqlc
