# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## 项目概述

这是一个基于 Connect 框架的 Go 微服务项目，使用 Protocol Buffers 定义 API，sqlc 生成数据库代码。

## 常用开发命令

### 环境初始化
```bash
# 安装必要工具
make install-tools

# 启动 PostgreSQL 和 Redis（使用 Docker）
make docker/up

# 设置数据库环境变量
export DATABASE_URL="postgres://pigeon:pigeon123@localhost:5432/pigeon_db?sslmode=disable"

# 运行数据库迁移
make migrate/up

# 完整的开发环境设置
make dev/setup
```

### 代码生成
```bash
# 生成 Protocol Buffer 和 Connect 代码
make proto

# 生成 sqlc 数据库查询代码
make sqlc
```

### 构建和运行
```bash
# 构建所有服务
make build

# 运行单个服务
make run/user-service
make run/order-service

# 运行测试
make test

# 代码检查
make lint
```

### 数据库迁移
```bash
# 创建新的迁移文件
make migrate/create name=add_new_table

# 应用迁移
make migrate/up

# 回滚迁移
make migrate/down
```

## 架构模式

### 服务分层架构
每个微服务遵循以下分层：

1. **Connect Handler** (`internal/service/{service}/connect.go`)
   - 实现 Connect RPC 接口
   - 处理请求/响应转换
   - 错误码映射

2. **Service Layer** (`internal/service/{service}/service.go`)
   - 业务逻辑实现
   - 事务管理
   - 业务验证

3. **Store Layer** (`internal/service/{service}/store.go`)
   - 数据访问抽象
   - 缓存集成（Redis）
   - 数据库连接管理

4. **Database Layer** (`internal/service/{service}/db/`)
   - sqlc 生成的代码
   - SQL 查询定义（`queries/` 目录）

### 添加新功能的工作流

1. **定义 API**：在 `api/proto/{service}/v1/` 编辑 `.proto` 文件
2. **生成代码**：运行 `make proto` 生成 Connect 代码
3. **定义 SQL 查询**：在 `internal/service/{service}/queries/` 添加 SQL
4. **生成数据库代码**：运行 `make sqlc`
5. **实现业务逻辑**：
   - 在 Store 层添加数据访问方法
   - 在 Service 层实现业务逻辑
   - 在 Connect Handler 实现 RPC 方法
6. **测试**：运行 `make test`
7. **代码检查**：运行 `make lint`

## 项目结构说明

- `api/proto/` - Protocol Buffer 定义
- `gen/` - 自动生成的代码（不要手动修改）
- `internal/service/` - 各微服务实现
- `internal/pkg/` - 共享组件（配置、数据库、日志等）
- `cmd/` - 服务入口点
- `migrations/` - 数据库迁移文件
- `configs/` - 服务配置文件

## 重要文档链接

- Connect 框架 (Go): https://connectrpc.com/docs/go/getting-started
- Buf 配置文档:
  - buf.yaml: https://buf.build/docs/configuration/v2/buf-yaml/
  - buf.gen.yaml: https://buf.build/docs/configuration/v2/buf-gen-yaml/
- sqlc 文档: https://docs.sqlc.dev/en/latest/tutorials/getting-started-postgresql.html

## 开发注意事项

- 所有生成的代码位于 `gen/` 目录，请勿手动修改
- 数据库查询使用 sqlc，SQL 文件位于 `internal/service/{service}/queries/`
- 每个服务都有独立的配置文件在 `configs/` 目录
- 使用 `make docker/up` 启动的服务会在后台运行，使用 `make docker/down` 停止
- 密码哈希使用 bcrypt，UUID 使用 github.com/google/uuid