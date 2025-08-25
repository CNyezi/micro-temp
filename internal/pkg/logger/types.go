package logger

import (
	"context"
	
	"go.uber.org/zap"
)

// Logger 定义统一的日志记录接口
type Logger interface {
	// 基础日志方法 - 直接使用 zap.Field
	Debug(msg string, fields ...zap.Field)
	Info(msg string, fields ...zap.Field)
	Warn(msg string, fields ...zap.Field)
	Error(msg string, fields ...zap.Field)
	Fatal(msg string, fields ...zap.Field)

	// 上下文日志方法
	DebugContext(ctx context.Context, msg string, fields ...zap.Field)
	InfoContext(ctx context.Context, msg string, fields ...zap.Field)
	WarnContext(ctx context.Context, msg string, fields ...zap.Field)
	ErrorContext(ctx context.Context, msg string, fields ...zap.Field)

	// 结构化日志方法
	WithFields(fields ...zap.Field) Logger
	WithContext(ctx context.Context) Logger
	WithService(service string) Logger

	// 条件日志方法
	IfDebug() ConditionalLogger
	IfInfo() ConditionalLogger

	// 配置和管理
	SetLevel(level Level)
	Sync() error
}

// ConditionalLogger 条件日志记录接口
type ConditionalLogger interface {
	Log(msg string, fields ...zap.Field)
	Logf(format string, args ...any)
}

// Level 日志级别
type Level int8

const (
	DebugLevel Level = iota - 1
	InfoLevel
	WarnLevel
	ErrorLevel
	FatalLevel
)

// String 返回日志级别的字符串表示
func (l Level) String() string {
	switch l {
	case DebugLevel:
		return "debug"
	case InfoLevel:
		return "info"
	case WarnLevel:
		return "warn"
	case ErrorLevel:
		return "error"
	case FatalLevel:
		return "fatal"
	default:
		return "unknown"
	}
}

// ParseLevel 解析字符串为日志级别
func ParseLevel(s string) Level {
	switch s {
	case "debug":
		return DebugLevel
	case "info":
		return InfoLevel
	case "warn":
		return WarnLevel
	case "error":
		return ErrorLevel
	case "fatal":
		return FatalLevel
	default:
		return InfoLevel
	}
}


// ConnectInterceptor Connect RPC 日志中间件
type ConnectInterceptor struct {
	logger Logger
	config InterceptorConfig
}

// InterceptorConfig 中间件配置
type InterceptorConfig struct {
	LogRequests     bool
	LogResponses    bool
	LogHeaders      bool
	SensitiveFields []string
	MaxBodySize     int
}


// OutputConfig 输出配置
type OutputConfig struct {
	Console ConsoleOutputConfig
	File    FileOutputConfig  
	Remote  RemoteOutputConfig
}

// ConsoleOutputConfig 控制台输出配置
type ConsoleOutputConfig struct {
	Enabled bool
}

// FileOutputConfig 文件输出配置  
type FileOutputConfig struct {
	Enabled    bool
	Path       string
	MaxSize    int
	MaxBackups int
	MaxAge     int
	Compress   bool
}

// RemoteOutputConfig 远程输出配置
type RemoteOutputConfig struct {
	Enabled   bool
	Endpoint  string
	Protocol  string
	BatchSize int
	Timeout   int
	TLS       bool
}

// TracingConfig 追踪配置
type TracingConfig struct {
	Enabled bool
}

// LoggerConfig 日志器配置
type LoggerConfig struct {
	Level       Level
	Format      string
	ServiceName string
	Version     string
	Environment string
	Output      OutputConfig
	Tracing     TracingConfig
}