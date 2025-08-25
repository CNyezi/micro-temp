# 结构指导文档

## 项目组织

### 根目录结构
```
micro-temp/
├── api/                    # API 定义和生成代码
│   └── proto/              # Protocol Buffer 定义文件
│       ├── user/
│       │   └── v1/
│       │       └── user.proto
│       └── order/
│           └── v1/
│               └── order.proto
├── cmd/                    # 服务可执行文件入口
│   ├── user-service/
│   │   └── main.go
│   └── order-service/
│       └── main.go
├── configs/                # 配置文件
│   ├── user-service.yaml
│   └── order-service.yaml
├── gen/                    # 自动生成的代码（不手动修改）
│   ├── proto/              # Protocol Buffer 生成的 Go 代码
│   └── sqlc/               # sqlc 生成的数据库代码
├── internal/               # 内部包，不对外暴露
│   ├── pkg/                # 共享内部组件
│   │   ├── config/         # 配置管理
│   │   ├── database/       # 数据库连接池
│   │   ├── logger/         # 结构化日志
│   │   ├── auth/           # JWT 认证中间件
│   │   ├── cache/          # Redis 缓存封装
│   │   └── errors/         # 统一错误处理
│   └── service/            # 各微服务实现
│       ├── user/           # 用户服务
│       │   ├── connect.go  # Connect RPC 处理器
│       │   ├── service.go  # 业务逻辑实现
│       │   ├── store.go    # 数据访问层
│       │   ├── queries/    # SQL 查询定义
│       │   │   ├── users.sql
│       │   │   └── sessions.sql
│       │   └── db/         # sqlc 生成的查询代码
│       └── order/          # 订单服务
│           ├── connect.go
│           ├── service.go
│           ├── store.go
│           ├── queries/
│           │   └── orders.sql
│           └── db/
├── migrations/             # 数据库迁移文件
│   ├── 20240101000001_create_users_table.up.sql
│   ├── 20240101000001_create_users_table.down.sql
│   ├── 20240101000002_create_orders_table.up.sql
│   └── 20240101000002_create_orders_table.down.sql
├── scripts/                # 构建和部署脚本
│   ├── build.sh
│   ├── test.sh
│   └── deploy.sh
├── .spec-workflow/         # 规格工作流文档
│   ├── steering/           # 指导文档
│   └── specs/              # 功能规格文档
├── docker-compose.yml      # 本地开发环境
├── Dockerfile.user         # 用户服务容器化
├── Dockerfile.order        # 订单服务容器化
├── Makefile               # 开发命令集合
├── buf.yaml               # Buf 配置文件
├── buf.gen.yaml           # 代码生成配置
├── sqlc.yaml              # sqlc 配置文件
├── go.mod                 # Go 模块定义
├── go.sum                 # Go 依赖校验和
├── README.md              # 项目说明
└── CLAUDE.md              # Claude Code 指导文档
```

### 命名约定

#### 目录命名
- **小写 + 连字符**：`user-service`、`order-service`
- **复数形式**：`queries/`、`migrations/`、`configs/`
- **功能分组**：按功能域划分目录结构

#### 文件命名
```
服务实现文件：
├── connect.go      # Connect RPC 处理器实现
├── service.go      # 业务逻辑服务层
├── store.go        # 数据访问存储层
└── types.go        # 服务特定类型定义

配置文件：
├── user-service.yaml
└── order-service.yaml

SQL 文件：
├── users.sql       # 用户相关查询
├── orders.sql      # 订单相关查询
└── sessions.sql    # 会话相关查询

迁移文件：
├── 20240101000001_create_users_table.up.sql
└── 20240101000001_create_users_table.down.sql

测试文件：
├── service_test.go
├── store_test.go
└── integration_test.go
```

#### 包命名
```go
// 包命名使用单数形式，小写，无下划线
package userservice    // ✅ 正确
package user_service   // ❌ 错误：包含下划线
package UserService    // ❌ 错误：大写字母

// 导入路径反映目录结构
import (
    "github.com/project/internal/pkg/config"
    "github.com/project/internal/service/user" 
    "github.com/project/gen/proto/user/v1"
)
```

## 开发工作流程

### Git 分支策略

#### 分支模型
```
main                    # 主分支，始终保持可部署状态
├── develop             # 开发分支，集成最新功能
├── feature/user-auth   # 功能分支：用户认证
├── feature/order-mgmt  # 功能分支：订单管理
├── hotfix/critical-bug # 热修复分支：紧急问题
└── release/v1.0.0      # 发布分支：版本准备
```

#### 分支命名规范
- **功能分支**：`feature/描述性名称`
  - `feature/user-authentication`
  - `feature/order-payment-integration`
  
- **修复分支**：`fix/描述性名称`
  - `fix/user-login-validation`
  - `fix/order-status-update`
  
- **热修复分支**：`hotfix/描述性名称`
  - `hotfix/security-vulnerability`
  - `hotfix/database-connection-leak`

#### 提交信息规范
```
格式：<类型>(<范围>): <描述>

类型：
- feat: 新功能
- fix: 问题修复
- docs: 文档更新
- style: 代码格式（不影响功能）
- refactor: 代码重构
- perf: 性能优化
- test: 测试相关
- chore: 构建过程或辅助工具变动

示例：
feat(user): add JWT authentication middleware
fix(order): resolve order status update race condition
docs(api): update user service API documentation
refactor(store): optimize database query performance
```

### 代码审查流程

#### 审查清单
**功能性审查**
- [ ] 功能实现是否符合需求规格
- [ ] 错误处理是否完整和恰当
- [ ] 边界条件和异常情况是否考虑周全
- [ ] 单元测试覆盖率是否足够

**代码质量审查**
- [ ] 代码风格是否符合项目规范
- [ ] 变量和函数命名是否清晰易懂
- [ ] 代码复杂度是否合理
- [ ] 是否存在代码重复

**安全性审查**
- [ ] 输入验证是否充分
- [ ] 敏感数据是否正确处理
- [ ] 认证和授权逻辑是否正确
- [ ] SQL 注入和其他安全漏洞是否防范

**性能审查**
- [ ] 数据库查询是否优化
- [ ] 缓存策略是否合理
- [ ] 资源使用是否高效
- [ ] 并发安全是否保证

#### 审查流程步骤
1. **创建 Pull Request**
   ```bash
   # 创建功能分支
   git checkout -b feature/new-feature develop
   
   # 开发和提交
   git add .
   git commit -m "feat(service): implement new feature"
   
   # 推送并创建 PR
   git push origin feature/new-feature
   ```

2. **自动检查**
   - 运行 `make lint` 检查代码风格
   - 运行 `make test` 执行单元测试
   - 运行 `make build` 确保编译通过

3. **同行审查**
   - 至少一位同行开发者审查
   - 使用审查清单逐项检查
   - 通过 PR 评论提供反馈

4. **修改和重新审查**
   - 根据反馈修改代码
   - 重新推送更新
   - 确保所有问题解决

5. **合并到主分支**
   - 所有检查通过后合并
   - 删除功能分支
   - 更新本地开发分支

### 测试工作流程

#### 测试层级
```
测试金字塔：
    /\
   /E2E\     - 端到端测试：完整业务流程
  /____\
 /集成测试\   - 集成测试：服务间交互
/________\
/单元测试\ - 单元测试：函数和方法级别
/________\
```

#### 测试命令
```bash
# 运行所有测试
make test

# 运行单元测试
make test/unit

# 运行集成测试
make test/integration

# 运行端到端测试  
make test/e2e

# 生成测试覆盖率报告
make test/coverage
```

#### 测试文件组织
```
internal/service/user/
├── service.go          # 业务逻辑实现
├── service_test.go     # 单元测试
├── store.go            # 数据访问层
├── store_test.go       # 单元测试
├── integration_test.go # 集成测试
└── testdata/           # 测试数据文件
    ├── fixtures.sql
    └── mock_data.json
```

### 部署工作流程

#### 环境管理
```
开发环境 (Development)
├── 本地开发：docker-compose up
├── 功能测试：每个功能分支部署
└── 集成测试：develop 分支自动部署

预发布环境 (Staging)  
├── 发布候选：release 分支部署
├── 用户验收测试
└── 性能测试

生产环境 (Production)
├── 蓝绿部署：零停机更新
├── 金丝雀发布：渐进式发布
└── 监控告警：实时健康检查
```

#### 部署步骤
```bash
# 1. 构建镜像
make docker/build

# 2. 运行测试
make test/all

# 3. 推送镜像
make docker/push

# 4. 部署到环境
make deploy/staging    # 预发布环境
make deploy/production # 生产环境

# 5. 验证部署
make health/check
```

## 文档结构

### 文档组织
```
项目文档层次：
├── README.md              # 项目概述和快速开始
├── CLAUDE.md              # Claude Code 开发指导
├── .spec-workflow/        # 规格驱动开发文档
│   ├── steering/          # 指导文档
│   │   ├── product.md     # 产品指导
│   │   ├── tech.md        # 技术指导  
│   │   └── structure.md   # 结构指导
│   └── specs/             # 功能规格文档
│       ├── user-auth/     # 用户认证功能
│       └── order-mgmt/    # 订单管理功能
├── docs/                  # 详细技术文档
│   ├── api/               # API 文档
│   ├── deployment/        # 部署文档
│   ├── architecture/      # 架构文档
│   └── troubleshooting/   # 问题排查指南
└── internal/service/*/    # 代码级文档
    └── README.md          # 服务特定文档
```

### 文档维护
- **及时更新**：代码变更时同步更新相关文档
- **版本控制**：文档与代码一起进行版本管理
- **审查机制**：文档更新也需要经过审查
- **定期检查**：定期检查文档的准确性和完整性

## 团队协作规范

### 沟通指南
- **日常沟通**：通过 IM 工具进行日常交流
- **技术讨论**：使用 GitHub Issues 和 PR 评论
- **架构决策**：通过 ADR（架构决策记录）文档化
- **知识分享**：定期技术分享会和代码评审

### 会议结构
- **每日站会**：进度同步和问题发现（15分钟）
- **周计划会**：工作规划和优先级确定（30分钟）  
- **双周回顾**：工作总结和流程改进（60分钟）
- **技术分享**：技术学习和经验分享（30分钟）

### 决策流程
1. **问题识别**：明确需要决策的技术问题
2. **方案研究**：收集和分析可能的解决方案
3. **团队讨论**：组织技术讨论会评估方案
4. **决策记录**：使用 ADR 文档化决策过程和结果
5. **执行跟踪**：跟踪决策执行效果和可能的调整

### 知识管理
- **代码注释**：关键业务逻辑必须有清晰注释
- **ADR 记录**：重要技术决策必须文档化
- **技术债务**：使用 TODO 注释标记技术债务
- **最佳实践**：持续更新团队最佳实践文档

这个结构指导文档为团队提供了清晰的项目组织方式、开发流程和协作规范，确保项目的长期可维护性和团队效率。