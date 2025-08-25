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

# Template initialization commands
.PHONY: template/init
template/init:
	@echo "🚀 初始化新的微服务项目..."
	@read -p "请输入项目名称 (例如: my-awesome-service): " PROJECT_NAME && \
	if [ -z "$$PROJECT_NAME" ]; then \
		echo "❌ 项目名称不能为空"; \
		exit 1; \
	fi && \
	echo "📝 将项目名从 'micro-holtye' 替换为 '$$PROJECT_NAME'..." && \
	find . -name "*.go" -not -path "./gen/*" -exec sed -i "s/micro-holtye/$$PROJECT_NAME/g" {} \; && \
	find . -name "*.proto" -exec sed -i "s/micro-holtye/$$PROJECT_NAME/g" {} \; && \
	find . -name "go.mod" -exec sed -i "s/micro-holtye/$$PROJECT_NAME/g" {} \; && \
	echo "🎉 项目初始化完成！新项目名: $$PROJECT_NAME" && \
	echo "📋 下一步请运行: make dev/setup"

.PHONY: service/new
service/new:
	@read -p "请输入服务名称 (例如: product): " SERVICE_NAME && \
	if [ -z "$$SERVICE_NAME" ]; then \
		echo "❌ 服务名称不能为空"; \
		exit 1; \
	fi && \
	./scripts/new-service.sh "$$SERVICE_NAME"
