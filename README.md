# micro-holtye

基于 Connect 框架的 Go 微服务项目模板。

## 项目结构

```
.
├── api/                    # API 契约定义 (proto 文件)
├── cmd/                    # 服务入口
├── configs/                # 配置文件
├── gen/                    # 自动生成的代码 (不要手动修改)
├── internal/               # 内部代码
│   ├── pkg/               # 共享基础设施
│   └── service/           # 服务实现
├── migrations/            # 数据库迁移文件
└── Makefile              # 自动化命令
```

## 快速开始

### 1. 安装依赖工具

```bash
make install-tools
```

### 3. 生成代码

```bash
make proto
make sqlc
```

### 4. 运行数据库迁移

```bash
export DATABASE_URL="postgres://pigeon:pigeon123@localhost:5432/pigeon_db?sslmode=disable"
make migrate/up
```

### 5. 启动服务

用户服务:
```bash
make run/user-service
```

订单服务:
```bash
make run/order-service
```

## 开发流程

1. **定义 API**: 在 `api/proto/` 目录下编辑 `.proto` 文件
2. **生成代码**: 运行 `make proto` 生成 Connect 代码
3. **实现服务**: 在 `internal/service/` 目录下实现业务逻辑
4. **测试**: 运行 `make test`

## 技术栈

- **框架**: Connect (gRPC-compatible RPC framework)
- **配置**: Viper
- **数据库**: PostgreSQL + sqlc
- **缓存**: Redis
- **日志**: Zap
- **工具链**: Buf

## 常用命令

- `make proto` - 生成 protobuf 代码
- `make sqlc` - 生成数据库查询代码
- `make build` - 构建所有服务
- `make test` - 运行测试
- `make lint` - 运行代码检查
- `make migrate/create name=<name>` - 创建新的数据库迁移