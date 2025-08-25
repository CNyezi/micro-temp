# 技术指导文档

## 架构概述

### 系统架构图
```
┌─────────────────────────────────────────────────────────────┐
│                        客户端应用                            │
└─────────────────────┬───────────────────────────────────────┘
                      │ Connect RPC over HTTP/2
┌─────────────────────┼───────────────────────────────────────┐
│                   API 网关                                   │
├─────────────────────┼───────────────────────────────────────┤
│              服务发现 & 负载均衡                             │
└─────────────────────┬───────────────────────────────────────┘
                      │
       ┌──────────────┼──────────────┐
       │              │              │
┌─────────────┐ ┌─────────────┐ ┌─────────────┐
│ User Service│ │Order Service│ │ ... Service │
│             │ │             │ │             │
│ Connect RPC │ │ Connect RPC │ │ Connect RPC │
│ Handler     │ │ Handler     │ │ Handler     │
│             │ │             │ │             │
│ Business    │ │ Business    │ │ Business    │
│ Logic Layer │ │ Logic Layer │ │ Logic Layer │
│             │ │             │ │             │
│ Data Access │ │ Data Access │ │ Data Access │
│ Layer (sqlc)│ │ Layer (sqlc)│ │ Layer (sqlc)│
└─────┬───────┘ └─────┬───────┘ └─────┬───────┘
      │               │               │
      └───────┬───────┼───────┬───────┘
              │       │       │
        ┌─────────────┴───────────────┐
        │        PostgreSQL            │
        │      (Primary Database)      │
        └─────────────────────────────┘
              │               │
        ┌─────────────┐ ┌─────────────┐
        │    Redis    │ │  监控系统    │
        │   (Cache)   │ │(Prometheus) │
        └─────────────┘ └─────────────┘
```

### 技术栈选择

#### 后端框架
- **Connect for Go**: 基于 HTTP/2 的高性能 RPC 框架
  - 类型安全的 gRPC 兼容 API
  - 原生支持 HTTP/JSON 和二进制协议
  - 优秀的流式处理能力
  
#### API 定义
- **Protocol Buffers (proto3)**: 
  - 强类型 API 契约
  - 自动代码生成
  - 版本兼容性管理
  - 跨语言支持

#### 数据库技术
- **PostgreSQL**: 主要关系型数据库
  - ACID 事务支持
  - 丰富的数据类型
  - 强大的查询优化器
  
- **sqlc**: 类型安全的 SQL 代码生成器
  - 编译时 SQL 验证
  - 零运行时反射开销
  - 原生 PostgreSQL 特性支持

- **Redis**: 内存缓存和会话存储
  - 高性能键值存储
  - 丰富的数据结构
  - 持久化选项

## 开发标准

### 代码规范

#### Go 代码标准
```go
// 包命名：简短、小写、无下划线
package userservice

// 接口命名：以 -er 结尾
type UserStorer interface {
    GetUser(ctx context.Context, id uuid.UUID) (*User, error)
    CreateUser(ctx context.Context, user *User) error
}

// 错误处理：使用 fmt.Errorf 包装错误
func (s *Service) GetUser(ctx context.Context, id uuid.UUID) (*User, error) {
    user, err := s.store.GetUser(ctx, id)
    if err != nil {
        return nil, fmt.Errorf("failed to get user %s: %w", id, err)
    }
    return user, nil
}

// 结构体标签：使用 json, db 标签
type User struct {
    ID        uuid.UUID `json:"id" db:"id"`
    Email     string    `json:"email" db:"email"`
    CreatedAt time.Time `json:"created_at" db:"created_at"`
}
```

#### 目录结构标准
```
cmd/                    # 服务入口点
├── user-service/
└── order-service/

internal/               # 内部包，不对外暴露
├── pkg/               # 共享内部包
│   ├── config/        # 配置管理
│   ├── database/      # 数据库连接
│   ├── logger/        # 结构化日志
│   └── auth/          # 认证中间件
└── service/           # 各微服务实现
    ├── user/          # 用户服务
    │   ├── connect.go # Connect RPC 处理器
    │   ├── service.go # 业务逻辑层
    │   ├── store.go   # 数据访问层
    │   ├── queries/   # SQL 查询文件
    │   └── db/        # sqlc 生成的代码
    └── order/         # 订单服务
```

#### 测试规范
```go
// 测试文件命名：*_test.go
// 测试函数命名：TestFunctionName
func TestUserService_CreateUser(t *testing.T) {
    tests := []struct {
        name    string
        input   *CreateUserRequest
        want    *User
        wantErr bool
    }{
        {
            name: "valid user creation",
            input: &CreateUserRequest{
                Email:    "test@example.com",
                Password: "password123",
            },
            want: &User{
                Email: "test@example.com",
            },
            wantErr: false,
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // 测试实现
        })
    }
}
```

### 安全标准

#### 认证授权
- **JWT 令牌**：用于无状态认证
- **RBAC 权限模型**：基于角色的访问控制
- **API 密钥管理**：服务间认证

#### 数据安全
```go
// 密码哈希：使用 bcrypt
import "golang.org/x/crypto/bcrypt"

func HashPassword(password string) (string, error) {
    bytes, err := bcrypt.GenerateFromPassword([]byte(password), 14)
    return string(bytes), err
}

// 敏感数据脱敏：日志中隐藏敏感信息
type User struct {
    ID       uuid.UUID `json:"id"`
    Email    string    `json:"email"`
    Password string    `json:"-"` // 不序列化到 JSON
}
```

#### 输入验证
```go
// 使用 validator 库进行输入验证
import "github.com/go-playground/validator/v10"

type CreateUserRequest struct {
    Email    string `validate:"required,email" json:"email"`
    Password string `validate:"required,min=8" json:"password"`
}

func (r *CreateUserRequest) Validate() error {
    validate := validator.New()
    return validate.Struct(r)
}
```

## 性能标准

### 响应时间要求
- **API 响应时间**：P95 < 50ms, P99 < 100ms
- **数据库查询**：P95 < 10ms, P99 < 50ms
- **缓存访问**：P95 < 1ms, P99 < 5ms

### 并发处理
```go
// 使用 context 进行超时控制
ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
defer cancel()

// 使用 sync.WaitGroup 处理并发任务
var wg sync.WaitGroup
for _, item := range items {
    wg.Add(1)
    go func(item Item) {
        defer wg.Done()
        processItem(ctx, item)
    }(item)
}
wg.Wait()
```

### 缓存策略
```go
// Redis 缓存模式
func (s *Store) GetUserWithCache(ctx context.Context, id uuid.UUID) (*User, error) {
    // 1. 尝试从缓存获取
    cacheKey := fmt.Sprintf("user:%s", id)
    cached, err := s.redis.Get(ctx, cacheKey).Result()
    if err == nil {
        var user User
        json.Unmarshal([]byte(cached), &user)
        return &user, nil
    }
    
    // 2. 从数据库获取
    user, err := s.queries.GetUser(ctx, id)
    if err != nil {
        return nil, err
    }
    
    // 3. 写入缓存
    userJSON, _ := json.Marshal(user)
    s.redis.Set(ctx, cacheKey, userJSON, time.Hour)
    
    return user, nil
}
```

## 集成模式

### 服务间通信
```proto
// user/v1/user.proto - Protocol Buffer 定义
syntax = "proto3";
package user.v1;

service UserService {
  rpc CreateUser(CreateUserRequest) returns (CreateUserResponse);
  rpc GetUser(GetUserRequest) returns (GetUserResponse);
  rpc UpdateUser(UpdateUserRequest) returns (UpdateUserResponse);
}

message CreateUserRequest {
  string email = 1;
  string password = 2;
}
```

### 数据库集成
```sql
-- internal/service/user/queries/users.sql
-- name: GetUser :one
SELECT id, email, password_hash, created_at, updated_at
FROM users
WHERE id = $1;

-- name: CreateUser :one
INSERT INTO users (id, email, password_hash, created_at, updated_at)
VALUES ($1, $2, $3, $4, $5)
RETURNING *;

-- name: UpdateUser :exec
UPDATE users
SET email = $2, updated_at = $3
WHERE id = $1;
```

### 监控集成
```go
// 使用 Prometheus 指标收集
import (
    "github.com/prometheus/client_golang/prometheus"
    "github.com/prometheus/client_golang/prometheus/promauto"
)

var (
    httpRequestsTotal = promauto.NewCounterVec(
        prometheus.CounterOpts{
            Name: "http_requests_total",
            Help: "Total number of HTTP requests",
        },
        []string{"method", "endpoint", "status"},
    )
    
    requestDuration = promauto.NewHistogramVec(
        prometheus.HistogramOpts{
            Name: "http_request_duration_seconds",
            Help: "HTTP request duration in seconds",
        },
        []string{"method", "endpoint"},
    )
)
```

## 部署和运维

### 容器化标准
```dockerfile
# Dockerfile 模板
FROM golang:1.21-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o service ./cmd/service

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /app/service .
EXPOSE 8080
CMD ["./service"]
```

### 健康检查
```go
// 健康检查端点
func (s *Server) HealthCheck(ctx context.Context, req *connect.Request[emptypb.Empty]) (*connect.Response[*v1.HealthResponse], error) {
    // 检查数据库连接
    if err := s.db.PingContext(ctx); err != nil {
        return nil, connect.NewError(connect.CodeUnavailable, fmt.Errorf("database unhealthy: %w", err))
    }
    
    // 检查 Redis 连接
    if err := s.redis.Ping(ctx).Err(); err != nil {
        return nil, connect.NewError(connect.CodeUnavailable, fmt.Errorf("redis unhealthy: %w", err))
    }
    
    return connect.NewResponse(&v1.HealthResponse{
        Status: "healthy",
    }), nil
}
```

### 配置管理
```go
// 环境配置结构
type Config struct {
    Server struct {
        Host string `env:"SERVER_HOST" envDefault:"localhost"`
        Port int    `env:"SERVER_PORT" envDefault:"8080"`
    }
    Database struct {
        URL string `env:"DATABASE_URL" envDefault:"postgres://localhost/app"`
    }
    Redis struct {
        URL string `env:"REDIS_URL" envDefault:"redis://localhost:6379"`
    }
    Auth struct {
        JWTSecret string `env:"JWT_SECRET" envDefault:"secret"`
    }
}
```

这个技术指导文档为整个微服务平台的技术实现提供了详细的标准和最佳实践，确保代码质量、性能和可维护性。