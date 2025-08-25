package logger

import (
	"fmt"
	"os"

	"go.uber.org/zap"
)

// LoggerFactory 日志器工厂
type LoggerFactory struct {
	defaultConfig LoggerConfig
}

// NewLoggerFactory 创建日志器工厂
func NewLoggerFactory() *LoggerFactory {
	return &LoggerFactory{
		defaultConfig: DefaultLoggerConfig(),
	}
}

// DefaultLoggerConfig 默认日志器配置
func DefaultLoggerConfig() LoggerConfig {
	return LoggerConfig{
		Level:       InfoLevel,
		Format:      "console",
		ServiceName: "unknown-service",
		Version:     "1.0.0",
		Environment: "development",
		Output: OutputConfig{
			Console: ConsoleOutputConfig{Enabled: true},
			File:    FileOutputConfig{Enabled: false},
			Remote:  RemoteOutputConfig{Enabled: false},
		},
		Tracing: TracingConfig{Enabled: false},
	}
}

// CreateLogger 创建日志器（统一入口）
func (f *LoggerFactory) CreateLogger(config LoggerConfig) (Logger, error) {
	// 合并默认配置
	mergedConfig := f.mergeWithDefaults(config)
	
	// 根据配置选择创建方式
	if mergedConfig.Tracing.Enabled {
		// 启用追踪的日志器
		return CreateTracedLogger(mergedConfig)
	}
	
	if hasMultipleOutputs(mergedConfig) {
		// 多输出日志器
		return CreateLoggerWithOutputs(mergedConfig)
	}
	
	// 基础日志器
	return NewLogger(mergedConfig)
}

// CreateSimpleLogger 创建简单日志器
func (f *LoggerFactory) CreateSimpleLogger(level Level, format string) (Logger, error) {
	config := LoggerConfig{
		Level:  level,
		Format: format,
	}
	return f.CreateLogger(config)
}

// CreateServiceLogger 创建服务日志器（带服务信息）
func (f *LoggerFactory) CreateServiceLogger(serviceName, version, environment string) (Logger, error) {
	config := LoggerConfig{
		ServiceName: serviceName,
		Version:     version,
		Environment: environment,
	}
	return f.CreateLogger(config)
}

// CreateFileLogger 创建文件日志器
func (f *LoggerFactory) CreateFileLogger(filePath string) (Logger, error) {
	config := LoggerConfig{
		Output: OutputConfig{
			Console: ConsoleOutputConfig{Enabled: false},
			File: FileOutputConfig{
				Enabled: true,
				Path:    filePath,
			},
		},
	}
	return f.CreateLogger(config)
}

// CreateTracedServiceLogger 创建带追踪的服务日志器
func (f *LoggerFactory) CreateTracedServiceLogger(serviceName, version, environment string) (Logger, error) {
	config := LoggerConfig{
		ServiceName: serviceName,
		Version:     version,
		Environment: environment,
		Tracing: TracingConfig{
			Enabled: true,
		},
	}
	return f.CreateLogger(config)
}

// mergeWithDefaults 合并默认配置
func (f *LoggerFactory) mergeWithDefaults(config LoggerConfig) LoggerConfig {
	result := f.defaultConfig
	
	// 覆盖非零值
	if config.Level != 0 {
		result.Level = config.Level
	}
	if config.Format != "" {
		result.Format = config.Format
	}
	if config.ServiceName != "" {
		result.ServiceName = config.ServiceName
	}
	if config.Version != "" {
		result.Version = config.Version
	}
	if config.Environment != "" {
		result.Environment = config.Environment
	}
	
	// 合并输出配置
	if config.Output.Console.Enabled {
		result.Output.Console = config.Output.Console
	}
	if config.Output.File.Enabled {
		result.Output.File = config.Output.File
	}
	if config.Output.Remote.Enabled {
		result.Output.Remote = config.Output.Remote
	}
	
	// 合并追踪配置
	if config.Tracing.Enabled {
		result.Tracing = config.Tracing
	}
	
	return result
}

// 全局工厂实例
var defaultFactory = NewLoggerFactory()

// 便捷的全局函数

// CreateLogger 创建日志器（全局便捷函数）
func CreateLogger(config LoggerConfig) (Logger, error) {
	return defaultFactory.CreateLogger(config)
}

// CreateSimpleLogger 创建简单日志器（全局便捷函数）
func CreateSimpleLogger(level Level, format string) (Logger, error) {
	return defaultFactory.CreateSimpleLogger(level, format)
}

// CreateServiceLogger 创建服务日志器（全局便捷函数）
func CreateServiceLogger(serviceName, version, environment string) (Logger, error) {
	return defaultFactory.CreateServiceLogger(serviceName, version, environment)
}

// CreateFileLogger 创建文件日志器（全局便捷函数）
func CreateFileLogger(filePath string) (Logger, error) {
	return defaultFactory.CreateFileLogger(filePath)
}

// CreateTracedServiceLogger 创建带追踪的服务日志器（全局便捷函数）
func CreateTracedServiceLogger(serviceName, version, environment string) (Logger, error) {
	return defaultFactory.CreateTracedServiceLogger(serviceName, version, environment)
}

// 兼容性函数 - 保持与现有 observability.NewLogger 的兼容性

// NewLoggerCompat 兼容现有的 NewLogger API
func NewLoggerCompat(level, format string) (Logger, error) {
	config := LoggerConfig{
		Level:  ParseLevel(level),
		Format: format,
	}
	return CreateLogger(config)
}

// NewLoggerFromEnvCompat 兼容现有的 NewLoggerFromEnv API  
func NewLoggerFromEnvCompat() (Logger, error) {
	level := os.Getenv("LOG_LEVEL")
	if level == "" {
		level = "info"
	}

	format := os.Getenv("LOG_FORMAT")
	if format == "" {
		format = "console"
	}

	serviceName := os.Getenv("SERVICE_NAME")
	version := os.Getenv("SERVICE_VERSION")
	environment := os.Getenv("ENVIRONMENT")

	config := LoggerConfig{
		Level:       ParseLevel(level),
		Format:      format,
		ServiceName: serviceName,
		Version:     version,
		Environment: environment,
		Tracing: TracingConfig{
			Enabled: os.Getenv("TRACING_ENABLED") == "true",
		},
	}

	return CreateLogger(config)
}

// CreateMiddleware 创建中间件（全局便捷函数）
func CreateMiddleware(logger Logger, options ...InterceptorOption) *ConnectLoggingInterceptor {
	return NewConnectLoggingInterceptorWithOptions(logger, options...)
}

// CreateMiddlewareWithDefaults 使用默认配置创建中间件
func CreateMiddlewareWithDefaults(logger Logger) *ConnectLoggingInterceptor {
	return NewConnectLoggingInterceptor(logger, DefaultMiddlewareConfig())
}

// 预设配置

// ProductionLoggerConfig 生产环境日志配置
func ProductionLoggerConfig(serviceName string) LoggerConfig {
	return LoggerConfig{
		Level:       InfoLevel,
		Format:      "json",
		ServiceName: serviceName,
		Version:     os.Getenv("SERVICE_VERSION"),
		Environment: "production",
		Output: OutputConfig{
			Console: ConsoleOutputConfig{Enabled: true},
			File: FileOutputConfig{
				Enabled:    true,
				Path:       fmt.Sprintf("logs/%s.log", serviceName),
				MaxSize:    100, // 100MB
				MaxBackups: 3,
				MaxAge:     7, // 7天
				Compress:   true,
			},
		},
		Tracing: TracingConfig{Enabled: true},
	}
}

// DevelopmentLoggerConfig 开发环境日志配置
func DevelopmentLoggerConfig(serviceName string) LoggerConfig {
	return LoggerConfig{
		Level:       DebugLevel,
		Format:      "console",
		ServiceName: serviceName,
		Version:     "dev",
		Environment: "development",
		Output: OutputConfig{
			Console: ConsoleOutputConfig{Enabled: true},
		},
		Tracing: TracingConfig{Enabled: false},
	}
}

// TestLoggerConfig 测试环境日志配置
func TestLoggerConfig(serviceName string) LoggerConfig {
	return LoggerConfig{
		Level:       WarnLevel, // 测试时只记录警告和错误
		Format:      "json",
		ServiceName: serviceName,
		Version:     "test",
		Environment: "test",
		Output: OutputConfig{
			Console: ConsoleOutputConfig{Enabled: true},
		},
		Tracing: TracingConfig{Enabled: false},
	}
}

// QuickStart 快速开始函数

// MustCreateLogger 创建日志器，失败时 panic
func MustCreateLogger(config LoggerConfig) Logger {
	logger, err := CreateLogger(config)
	if err != nil {
		panic(fmt.Sprintf("Failed to create logger: %v", err))
	}
	return logger
}

// MustCreateServiceLogger 创建服务日志器，失败时 panic
func MustCreateServiceLogger(serviceName, environment string) Logger {
	var config LoggerConfig
	
	switch environment {
	case "production", "prod":
		config = ProductionLoggerConfig(serviceName)
	case "development", "dev":
		config = DevelopmentLoggerConfig(serviceName)
	case "test", "testing":
		config = TestLoggerConfig(serviceName)
	default:
		config = DevelopmentLoggerConfig(serviceName)
		config.Environment = environment
	}
	
	return MustCreateLogger(config)
}

// QuickLogger 快速创建日志器（用于简单场景）
func QuickLogger() Logger {
	return MustCreateServiceLogger("quick-service", "development")
}

// 构建信息

// BuildInfo 构建信息
type BuildInfo struct {
	Service     string
	Version     string
	Environment string
	BuildTime   string
	GitCommit   string
}

// CreateLoggerWithBuildInfo 使用构建信息创建日志器
func CreateLoggerWithBuildInfo(buildInfo BuildInfo) (Logger, error) {
	config := LoggerConfig{
		ServiceName: buildInfo.Service,
		Version:     buildInfo.Version,
		Environment: buildInfo.Environment,
	}
	
	logger, err := CreateLogger(config)
	if err != nil {
		return nil, err
	}
	
	// 添加构建信息字段
	buildFields := []zap.Field{
		zap.String("build_time", buildInfo.BuildTime),
		zap.String("git_commit", buildInfo.GitCommit),
	}
	
	return logger.WithFields(buildFields...), nil
}