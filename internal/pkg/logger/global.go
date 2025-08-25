package logger

import (
	"context"
	"sync"

	"go.uber.org/zap"
)

// 全局日志器管理
var (
	globalLogger     Logger
	globalLoggerLock sync.RWMutex
	globalInitOnce   sync.Once
)

// GetGlobalLogger 获取全局日志器
func GetGlobalLogger() Logger {
	globalLoggerLock.RLock()
	logger := globalLogger
	globalLoggerLock.RUnlock()
	
	if logger != nil {
		return logger
	}
	
	// 如果没有设置全局日志器，初始化默认的
	globalInitOnce.Do(initDefaultGlobalLogger)
	
	globalLoggerLock.RLock()
	defer globalLoggerLock.RUnlock()
	return globalLogger
}

// SetGlobalLogger 设置全局日志器
func SetGlobalLogger(logger Logger) {
	globalLoggerLock.Lock()
	defer globalLoggerLock.Unlock()
	
	// 同步旧的日志器
	if globalLogger != nil {
		_ = globalLogger.Sync()
	}
	
	globalLogger = logger
}

// ReplaceGlobalLogger 原子性替换全局日志器
func ReplaceGlobalLogger(logger Logger) func() {
	globalLoggerLock.Lock()
	defer globalLoggerLock.Unlock()
	
	oldLogger := globalLogger
	globalLogger = logger
	
	// 返回恢复函数
	return func() {
		SetGlobalLogger(oldLogger)
	}
}

// initDefaultGlobalLogger 初始化默认的全局日志器
func initDefaultGlobalLogger() {
	logger, err := NewLoggerFromEnv()
	if err != nil {
		// 如果创建失败，使用最基础的日志器
		logger = &fallbackLogger{}
	}
	
	globalLoggerLock.Lock()
	globalLogger = logger
	globalLoggerLock.Unlock()
}

// 全局便捷日志函数

// Debug 全局调试日志
func Debug(msg string, fields ...zap.Field) {
	GetGlobalLogger().Debug(msg, fields...)
}

// Info 全局信息日志
func Info(msg string, fields ...zap.Field) {
	GetGlobalLogger().Info(msg, fields...)
}

// Warn 全局警告日志
func Warn(msg string, fields ...zap.Field) {
	GetGlobalLogger().Warn(msg, fields...)
}

// Error 全局错误日志
func Error(msg string, fields ...zap.Field) {
	GetGlobalLogger().Error(msg, fields...)
}

// Fatal 全局致命错误日志
func Fatal(msg string, fields ...zap.Field) {
	GetGlobalLogger().Fatal(msg, fields...)
}

// WithFields 创建带字段的全局日志器
func WithFields(fields ...zap.Field) Logger {
	return GetGlobalLogger().WithFields(fields...)
}

// WithService 创建带服务名的全局日志器
func WithService(service string) Logger {
	return GetGlobalLogger().WithService(service)
}

// Sync 同步全局日志器
func Sync() error {
	return GetGlobalLogger().Sync()
}

// 全局日志器配置管理

// GlobalLoggerConfig 全局日志器配置
type GlobalLoggerConfig struct {
	config LoggerConfig
	mutex  sync.RWMutex
}

var globalConfig = &GlobalLoggerConfig{
	config: DefaultLoggerConfig(),
}

// SetGlobalLoggerConfig 设置全局日志器配置
func SetGlobalLoggerConfig(config LoggerConfig) error {
	globalConfig.mutex.Lock()
	defer globalConfig.mutex.Unlock()
	
	logger, err := CreateLogger(config)
	if err != nil {
		return err
	}
	
	globalConfig.config = config
	SetGlobalLogger(logger)
	
	return nil
}

// GetGlobalLoggerConfig 获取全局日志器配置
func GetGlobalLoggerConfig() LoggerConfig {
	globalConfig.mutex.RLock()
	defer globalConfig.mutex.RUnlock()
	return globalConfig.config
}

// UpdateGlobalLogLevel 更新全局日志级别
func UpdateGlobalLogLevel(level Level) {
	globalConfig.mutex.Lock()
	defer globalConfig.mutex.Unlock()
	
	globalConfig.config.Level = level
	GetGlobalLogger().SetLevel(level)
}

// GetGlobalLogLevel 获取全局日志级别
func GetGlobalLogLevel() Level {
	globalConfig.mutex.RLock()
	defer globalConfig.mutex.RUnlock()
	return globalConfig.config.Level
}

// 应用初始化辅助函数

// InitGlobalLogger 初始化全局日志器（应用启动时调用）
func InitGlobalLogger(serviceName, environment string) error {
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
	
	return SetGlobalLoggerConfig(config)
}

// InitGlobalLoggerFromEnv 从环境变量初始化全局日志器
func InitGlobalLoggerFromEnv() error {
	logger, err := NewLoggerFromEnvCompat()
	if err != nil {
		return err
	}
	
	SetGlobalLogger(logger)
	return nil
}

// MustInitGlobalLogger 初始化全局日志器，失败时 panic
func MustInitGlobalLogger(serviceName, environment string) {
	if err := InitGlobalLogger(serviceName, environment); err != nil {
		panic(err)
	}
}

// 优雅关闭支持

// Cleanup 清理全局日志器资源
func Cleanup() error {
	globalLoggerLock.RLock()
	logger := globalLogger
	globalLoggerLock.RUnlock()
	
	if logger != nil {
		return logger.Sync()
	}
	
	return nil
}

// RegisterCleanup 注册清理函数（可用于信号处理）
func RegisterCleanup(cleanup func()) {
	// 这里可以注册到应用的优雅关闭机制中
	// 具体实现取决于应用框架
}

// fallbackLogger 后备日志器（当主日志器创建失败时使用）
type fallbackLogger struct{}

func (f *fallbackLogger) Debug(msg string, fields ...zap.Field) {}
func (f *fallbackLogger) Info(msg string, fields ...zap.Field)  {}
func (f *fallbackLogger) Warn(msg string, fields ...zap.Field)  {}
func (f *fallbackLogger) Error(msg string, fields ...zap.Field) {}
func (f *fallbackLogger) Fatal(msg string, fields ...zap.Field) {}

func (f *fallbackLogger) DebugContext(ctx context.Context, msg string, fields ...zap.Field) {}
func (f *fallbackLogger) InfoContext(ctx context.Context, msg string, fields ...zap.Field)  {}
func (f *fallbackLogger) WarnContext(ctx context.Context, msg string, fields ...zap.Field)  {}
func (f *fallbackLogger) ErrorContext(ctx context.Context, msg string, fields ...zap.Field) {}

func (f *fallbackLogger) WithFields(fields ...zap.Field) Logger { return f }
func (f *fallbackLogger) WithContext(ctx context.Context) Logger { return f }
func (f *fallbackLogger) WithService(service string) Logger     { return f }

func (f *fallbackLogger) IfDebug() ConditionalLogger { return &noopConditionalLogger{} }
func (f *fallbackLogger) IfInfo() ConditionalLogger  { return &noopConditionalLogger{} }

func (f *fallbackLogger) SetLevel(level Level) {}
func (f *fallbackLogger) Sync() error          { return nil }

// 测试辅助函数

// ResetGlobalLogger 重置全局日志器（仅用于测试）
func ResetGlobalLogger() {
	globalLoggerLock.Lock()
	defer globalLoggerLock.Unlock()
	
	if globalLogger != nil {
		_ = globalLogger.Sync()
	}
	
	globalLogger = nil
	globalInitOnce = sync.Once{}
}

// SetTestGlobalLogger 设置测试用的全局日志器
func SetTestGlobalLogger(logger Logger) func() {
	oldLogger := GetGlobalLogger()
	SetGlobalLogger(logger)
	
	return func() {
		SetGlobalLogger(oldLogger)
	}
}