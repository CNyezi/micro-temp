package logger

import (
	"context"
	"fmt"
	"os"
	"runtime"
	"strings"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// zapLogger 实现 Logger 接口的 zap 包装器
type zapLogger struct {
	zap        *zap.Logger
	sugar      *zap.SugaredLogger
	level      Level
	service    string
	baseFields []zap.Field
}

// NewLogger 创建新的日志器实例
func NewLogger(config LoggerConfig) (Logger, error) {
	zapConfig := buildZapConfig(config)
	
	zapInst, err := zapConfig.Build()
	if err != nil {
		return nil, fmt.Errorf("failed to build zap logger: %w", err)
	}

	// 添加服务信息作为基础字段
	var baseFields []zap.Field
	if config.ServiceName != "" {
		baseFields = append(baseFields, zap.String("service", config.ServiceName))
	}
	if config.Version != "" {
		baseFields = append(baseFields, zap.String("version", config.Version))
	}
	if config.Environment != "" {
		baseFields = append(baseFields, zap.String("environment", config.Environment))
	}

	logger := &zapLogger{
		zap:        zapInst.With(baseFields...),
		sugar:      zapInst.Sugar(),
		level:      config.Level,
		service:    config.ServiceName,
		baseFields: baseFields,
	}

	return logger, nil
}

// buildZapConfig 构建 zap 配置
func buildZapConfig(config LoggerConfig) zap.Config {
	var zapConfig zap.Config

	if config.Format == "json" {
		zapConfig = zap.NewProductionConfig()
	} else {
		zapConfig = zap.NewDevelopmentConfig()
		zapConfig.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	}

	zapConfig.Level = zap.NewAtomicLevelAt(levelToZapLevel(config.Level))
	zapConfig.OutputPaths = []string{"stdout"}
	zapConfig.ErrorOutputPaths = []string{"stderr"}

	return zapConfig
}

// levelToZapLevel 转换日志级别
func levelToZapLevel(level Level) zapcore.Level {
	switch level {
	case DebugLevel:
		return zapcore.DebugLevel
	case InfoLevel:
		return zapcore.InfoLevel
	case WarnLevel:
		return zapcore.WarnLevel
	case ErrorLevel:
		return zapcore.ErrorLevel
	case FatalLevel:
		return zapcore.FatalLevel
	default:
		return zapcore.InfoLevel
	}
}

// Debug 记录调试日志
func (l *zapLogger) Debug(msg string, fields ...zap.Field) {
	l.zap.Debug(msg, l.enhanceFields(fields...)...)
}

// Info 记录信息日志
func (l *zapLogger) Info(msg string, fields ...zap.Field) {
	l.zap.Info(msg, l.enhanceFields(fields...)...)
}

// Warn 记录警告日志
func (l *zapLogger) Warn(msg string, fields ...zap.Field) {
	l.zap.Warn(msg, l.enhanceFields(fields...)...)
}

// Error 记录错误日志
func (l *zapLogger) Error(msg string, fields ...zap.Field) {
	// 自动添加堆栈信息到错误日志
	fieldsWithStack := append(fields, zap.String("stack_trace", getStackTrace()))
	l.zap.Error(msg, l.enhanceFields(fieldsWithStack...)...)
}

// Fatal 记录致命错误日志
func (l *zapLogger) Fatal(msg string, fields ...zap.Field) {
	// 自动添加堆栈信息到致命错误日志
	fieldsWithStack := append(fields, zap.String("stack_trace", getStackTrace()))
	l.zap.Fatal(msg, l.enhanceFields(fieldsWithStack...)...)
}

// DebugContext 记录带上下文的调试日志
func (l *zapLogger) DebugContext(ctx context.Context, msg string, fields ...zap.Field) {
	contextFields := extractContextFields(ctx)
	allFields := append(contextFields, fields...)
	l.Debug(msg, allFields...)
}

// InfoContext 记录带上下文的信息日志
func (l *zapLogger) InfoContext(ctx context.Context, msg string, fields ...zap.Field) {
	contextFields := extractContextFields(ctx)
	allFields := append(contextFields, fields...)
	l.Info(msg, allFields...)
}

// WarnContext 记录带上下文的警告日志
func (l *zapLogger) WarnContext(ctx context.Context, msg string, fields ...zap.Field) {
	contextFields := extractContextFields(ctx)
	allFields := append(contextFields, fields...)
	l.Warn(msg, allFields...)
}

// ErrorContext 记录带上下文的错误日志
func (l *zapLogger) ErrorContext(ctx context.Context, msg string, fields ...zap.Field) {
	contextFields := extractContextFields(ctx)
	allFields := append(contextFields, fields...)
	l.Error(msg, allFields...)
}

// WithFields 创建带有额外字段的日志器
func (l *zapLogger) WithFields(fields ...zap.Field) Logger {
	return &zapLogger{
		zap:        l.zap.With(fields...),
		sugar:      l.sugar,
		level:      l.level,
		service:    l.service,
		baseFields: l.baseFields,
	}
}

// WithContext 创建带有上下文的日志器
func (l *zapLogger) WithContext(ctx context.Context) Logger {
	contextFields := extractContextFields(ctx)
	return l.WithFields(contextFields...)
}

// WithService 创建带有服务名的日志器
func (l *zapLogger) WithService(service string) Logger {
	return &zapLogger{
		zap:        l.zap.With(zap.String("service", service)),
		sugar:      l.sugar,
		level:      l.level,
		service:    service,
		baseFields: l.baseFields,
	}
}

// IfDebug 返回条件调试日志器
func (l *zapLogger) IfDebug() ConditionalLogger {
	if l.level <= DebugLevel {
		return &conditionalLogger{logger: l, level: DebugLevel}
	}
	return &noopConditionalLogger{}
}

// IfInfo 返回条件信息日志器
func (l *zapLogger) IfInfo() ConditionalLogger {
	if l.level <= InfoLevel {
		return &conditionalLogger{logger: l, level: InfoLevel}
	}
	return &noopConditionalLogger{}
}

// SetLevel 设置日志级别
func (l *zapLogger) SetLevel(level Level) {
	l.level = level
	l.zap = l.zap.WithOptions(zap.IncreaseLevel(levelToZapLevel(level)))
	l.sugar = l.zap.Sugar()
}

// Sync 同步日志输出
func (l *zapLogger) Sync() error {
	return l.zap.Sync()
}

// enhanceFields 增强字段（添加敏感数据处理等）
func (l *zapLogger) enhanceFields(fields ...zap.Field) []zap.Field {
	// 这里可以添加敏感数据脱敏等逻辑
	// 目前直接返回，保持高性能
	return fields
}

// conditionalLogger 条件日志器实现
type conditionalLogger struct {
	logger Logger
	level  Level
}

func (c *conditionalLogger) Log(msg string, fields ...zap.Field) {
	switch c.level {
	case DebugLevel:
		c.logger.Debug(msg, fields...)
	case InfoLevel:
		c.logger.Info(msg, fields...)
	}
}

func (c *conditionalLogger) Logf(format string, args ...any) {
	msg := fmt.Sprintf(format, args...)
	c.Log(msg)
}

// noopConditionalLogger 空操作条件日志器
type noopConditionalLogger struct{}

func (n *noopConditionalLogger) Log(msg string, fields ...zap.Field) {}
func (n *noopConditionalLogger) Logf(format string, args ...any) {}

// extractContextFields 从上下文中提取字段 - 这是我们的核心增值功能
func extractContextFields(ctx context.Context) []zap.Field {
	var fields []zap.Field
	
	// 提取请求 ID
	if requestID := getRequestIDFromContext(ctx); requestID != "" {
		fields = append(fields, zap.String("request_id", requestID))
	}
	
	// 提取用户 ID
	if userID := getUserIDFromContext(ctx); userID != "" {
		fields = append(fields, zap.String("user_id", userID))
	}
	
	// TODO: 添加分布式追踪支持
	// if traceID := getTraceIDFromContext(ctx); traceID != "" {
	//     fields = append(fields, zap.String("trace_id", traceID))
	// }
	// if spanID := getSpanIDFromContext(ctx); spanID != "" {
	//     fields = append(fields, zap.String("span_id", spanID))
	// }
	
	return fields
}

// getRequestIDFromContext 从上下文获取请求ID
func getRequestIDFromContext(ctx context.Context) string {
	if val := ctx.Value("request_id"); val != nil {
		if requestID, ok := val.(string); ok {
			return requestID
		}
	}
	return ""
}

// getUserIDFromContext 从上下文获取用户ID
func getUserIDFromContext(ctx context.Context) string {
	if val := ctx.Value("user_id"); val != nil {
		if userID, ok := val.(string); ok {
			return userID
		}
	}
	return ""
}

// getStackTrace 获取堆栈跟踪
func getStackTrace() string {
	const depth = 32
	var pcs [depth]uintptr
	n := runtime.Callers(3, pcs[:])
	
	var sb strings.Builder
	frames := runtime.CallersFrames(pcs[:n])
	
	for {
		frame, more := frames.Next()
		if !more {
			break
		}
		
		// 跳过系统库的堆栈信息
		if strings.Contains(frame.File, "runtime/") {
			continue
		}
		
		sb.WriteString(fmt.Sprintf("%s:%d %s\n", frame.File, frame.Line, frame.Function))
	}
	
	return sb.String()
}

// NewLoggerFromEnv 从环境变量创建日志器 (兼容现有API)
func NewLoggerFromEnv() (Logger, error) {
	level := os.Getenv("LOG_LEVEL")
	if level == "" {
		level = "info"
	}

	format := os.Getenv("LOG_FORMAT")
	if format == "" {
		format = "console"
	}

	serviceName := os.Getenv("SERVICE_NAME")
	if serviceName == "" {
		serviceName = "unknown-service"
	}

	config := LoggerConfig{
		Level:       ParseLevel(level),
		Format:      format,
		ServiceName: serviceName,
		Version:     os.Getenv("SERVICE_VERSION"),
		Environment: os.Getenv("ENVIRONMENT"),
	}

	return NewLogger(config)
}