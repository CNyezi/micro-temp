# 统一日志组件实现计划

## 任务概述
本实现计划将基于现有的 `internal/pkg/observability/logger.go` 进行扩展和改进，创建一个统一的、功能完整的日志组件。计划按照渐进式的方式实施，确保每个阶段都可以独立测试和部署，最大化代码复用并保持向后兼容性。

## 任务列表

### 阶段 1：核心接口和配置扩展

- [x] 1. 扩展配置结构支持新的日志功能
  - 文件：internal/pkg/config/config.go
  - 扩展现有的 LogConfig 结构，添加输出、追踪、中间件和性能配置
  - 保持与现有配置的向后兼容性
  - _需求: 1.4, 1.5_
  - _复用: internal/pkg/config/config.go_

- [x] 2. 定义统一日志接口和数据结构
  - 文件：internal/pkg/logger/types.go
  - 定义 Logger 接口、Field 结构、ConditionalLogger 等核心类型
  - 定义标准化日志字段结构 StandardFields
  - _需求: 1.6_
  - _复用: 无（新建）_

- [x] 3. 实现核心日志器扩展
  - 文件：internal/pkg/logger/logger.go
  - 扩展现有的 zap 日志器实现 Logger 接口
  - 添加 WithFields、WithContext、WithService 等结构化日志方法
  - 实现条件日志记录功能（IfDebug、IfInfo）
  - _需求: 1.1, 1.6_
  - _复用: internal/pkg/observability/logger.go_

- [x] 4. 创建字段工具函数
  - 文件：internal/pkg/logger/fields.go
  - 实现常用字段的便捷创建函数（String、Int、Duration、Error 等）
  - 添加敏感数据检测和脱敏功能
  - _需求: 1.3_
  - _复用: 无（新建）_

### 阶段 2：格式化器和输出管理

- [x] 5. 实现可配置的格式化器
  - 文件：internal/pkg/logger/formatter.go
  - 创建 JSON 和控制台格式化器
  - 支持自定义字段顺序和格式选项
  - 集成敏感数据脱敏功能
  - _需求: 1.1, 1.3_
  - _复用: internal/pkg/observability/logger.go 中的格式化逻辑_

- [x] 6. 实现多输出目标管理
  - 文件：internal/pkg/logger/output.go
  - 支持控制台、文件、远程等多种输出目标
  - 实现输出目标的故障转移和恢复机制
  - 添加日志轮转和文件管理功能
  - _需求: 1.4_
  - _复用: internal/pkg/observability/logger.go 中的输出配置_

### 阶段 3：分布式追踪集成

- [x] 7. 实现追踪上下文提取器
  - 文件：internal/pkg/logger/tracing.go
  - 集成 OpenTelemetry，自动提取 trace ID 和 span ID
  - 实现追踪上下文的传播和注入
  - 添加服务信息和版本标识
  - _需求: 1.2_
  - _复用: 无（新建，可能需要添加 OpenTelemetry 依赖）_

- [x] 8. 扩展日志器支持追踪上下文
  - 文件：internal/pkg/logger/logger.go（继续）
  - 在 WithContext 方法中自动提取追踪信息
  - 为所有日志方法添加追踪字段支持
  - 实现追踪相关的便捷方法
  - _需求: 1.2_
  - _复用: internal/pkg/logger/logger.go_

### 阶段 4：Connect RPC 中间件统一

- [x] 9. 实现统一的 Connect 日志中间件
  - 文件：internal/pkg/logger/middleware.go
  - 创建 ConnectInterceptor 结构和配置
  - 实现请求/响应日志记录功能
  - 添加敏感字段脱敏和请求体大小限制
  - _需求: 1.3_
  - _复用: cmd/user-service/main.go 中的 loggingInterceptor 模式_

- [x] 10. 为中间件添加配置选项
  - 文件：internal/pkg/logger/middleware.go（继续）
  - 实现 InterceptorOption 模式
  - 支持选择性日志记录（请求、响应、头部信息）
  - 添加性能影响最小化的优化
  - _需求: 1.3_
  - _复用: internal/pkg/logger/middleware.go_

### 阶段 5：工厂函数和便捷 API

- [x] 11. 实现日志器工厂函数
  - 文件：internal/pkg/logger/factory.go
  - 创建 NewLogger、NewLoggerFromConfig 等工厂函数
  - 支持从环境变量、配置文件、代码配置创建日志器
  - 保持与现有 observability.NewLogger 的兼容性
  - _需求: 1.4, 1.6_
  - _复用: internal/pkg/observability/logger.go_

- [x] 12. 创建全局日志器管理
  - 文件：internal/pkg/logger/global.go
  - 提供全局日志器的设置和获取功能
  - 实现线程安全的日志器替换
  - 添加默认日志器的初始化
  - _需求: 1.6_
  - _复用: 无（新建）_

### 阶段 6：性能优化和监控

- [x] 13. 实现异步写入和缓冲机制
  - 文件：internal/pkg/logger/async.go
  - 创建异步日志写入器
  - 实现可配置的缓冲区和刷新策略
  - 添加背压处理和队列满载处理
  - _需求: 1.5_
  - _复用: 无（新建）_

- [x] 14. 添加日志组件自身的监控指标
  - 文件：internal/pkg/logger/metrics.go
  - 集成 Prometheus 指标收集
  - 监控日志写入速度、错误率、缓冲区使用情况
  - 提供健康检查接口
  - _需求: 1.5_
  - _复用: 可能复用现有的 Prometheus 集成（如果存在）_

### 阶段 7：现有服务集成和迁移

- [x] 15. 更新用户服务使用新的日志组件
  - 文件：cmd/user-service/main.go
  - 替换现有的 observability.NewLogger 调用
  - 使用统一的 Connect 中间件替换自定义 loggingInterceptor
  - 更新配置文件支持新的日志选项
  - _需求: 所有需求_
  - _复用: cmd/user-service/main.go, configs/user-service.yaml_

- [x] 16. 更新订单服务使用新的日志组件
  - 文件：cmd/order-service/main.go
  - 应用与用户服务相同的更新模式
  - 统一使用新的中间件和配置
  - _需求: 所有需求_
  - _复用: cmd/order-service/main.go, configs/order-service.yaml_

- [x] 17. 更新网关服务使用新的日志组件
  - 文件：cmd/gateway-service/main.go
  - 应用与其他服务相同的更新模式
  - 特别处理网关特定的日志需求（请求路由、负载均衡等）
  - _需求: 所有需求_
  - _复用: cmd/gateway-service/main.go, configs/gateway.yaml_

### 阶段 8：测试和文档

- [x] 18. 创建核心日志器单元测试
  - 文件：internal/pkg/logger/logger_test.go
  - 测试所有日志级别和方法
  - 测试结构化字段的正确处理
  - 测试追踪上下文的提取和使用
  - _需求: 所有需求_
  - _复用: 无（新建）_

- [x] 19. 创建中间件集成测试
  - 文件：internal/pkg/logger/middleware_test.go
  - 测试 Connect RPC 的日志记录功能
  - 测试敏感数据脱敏
  - 测试性能影响评估
  - _需求: 1.3_
  - _复用: 无（新建）_

- [x] 20. 创建端到端集成测试
  - 文件：test/integration/logger_integration_test.go
  - 测试完整的日志流程（从 API 请求到日志输出）
  - 测试多服务间的追踪上下文传播
  - 测试不同配置下的行为
  - _需求: 所有需求_
  - _复用: 可能复用现有的集成测试框架_

- [x] 21. 创建性能基准测试
  - 文件：internal/pkg/logger/benchmark_test.go
  - 基准测试日志记录性能
  - 对比新旧实现的性能差异
  - 测试并发场景下的性能表现
  - _需求: 1.5_
  - _复用: 无（新建）_

- [x] 22. 编写使用文档和示例
  - 文件：docs/unified-logger-guide.md
  - 详细的配置选项说明
  - 常见使用场景的代码示例
  - 迁移指南和最佳实践
  - _需求: 所有需求_
  - _复用: 无（新建）_

- [x] 23. 更新项目主文档
  - 文件：README.md, CLAUDE.md
  - 更新日志相关的使用说明
  - 添加新的日志配置示例
  - 更新开发指南中的日志部分
  - _需求: 所有需求_
  - _复用: README.md, CLAUDE.md_

### 阶段 9：向后兼容性和清理

- [x] 24. 创建向后兼容性适配器
  - 文件：internal/pkg/observability/compat.go
  - 保持现有 API 的兼容性
  - 提供从旧 API 到新 API 的迁移路径
  - 添加弃用警告和迁移建议
  - _需求: 所有需求_
  - _复用: internal/pkg/observability/logger.go_

- [x] 25. 清理和优化现有代码
  - 文件：internal/pkg/observability/logger.go
  - 标记旧实现为已弃用
  - 移除重复的代码和依赖
  - 优化包结构和导入路径
  - _需求: 所有需求_
  - _复用: internal/pkg/observability/logger.go_