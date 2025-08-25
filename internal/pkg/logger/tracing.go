package logger

import (
	"context"
	"strings"

	"go.uber.org/zap"
	"go.opentelemetry.io/otel/trace"
)

// TracingExtractor 追踪信息提取器
type TracingExtractor struct {
	enabled bool
}

// NewTracingExtractor 创建追踪提取器
func NewTracingExtractor(enabled bool) *TracingExtractor {
	return &TracingExtractor{enabled: enabled}
}

// ExtractTraceFields 从上下文中提取追踪字段
func (t *TracingExtractor) ExtractTraceFields(ctx context.Context) []zap.Field {
	if !t.enabled {
		return nil
	}

	var fields []zap.Field

	// 提取 OpenTelemetry 追踪信息
	span := trace.SpanFromContext(ctx)
	if span != nil && span.SpanContext().IsValid() {
		spanContext := span.SpanContext()
		
		// 添加 trace ID
		if spanContext.HasTraceID() {
			fields = append(fields, zap.String("trace_id", spanContext.TraceID().String()))
		}
		
		// 添加 span ID
		if spanContext.HasSpanID() {
			fields = append(fields, zap.String("span_id", spanContext.SpanID().String()))
		}
		
		// 添加追踪标志
		if spanContext.IsSampled() {
			fields = append(fields, zap.Bool("trace_sampled", true))
		}
	}

	return fields
}

// ExtractAllContextFields 提取所有上下文字段（包括追踪和业务字段）
func (t *TracingExtractor) ExtractAllContextFields(ctx context.Context) []zap.Field {
	var fields []zap.Field

	// 添加追踪字段
	traceFields := t.ExtractTraceFields(ctx)
	fields = append(fields, traceFields...)

	// 添加业务字段
	businessFields := extractBusinessContextFields(ctx)
	fields = append(fields, businessFields...)

	return fields
}

// extractBusinessContextFields 从上下文中提取业务字段
func extractBusinessContextFields(ctx context.Context) []zap.Field {
	var fields []zap.Field
	
	// 提取请求 ID
	if requestID := getContextValue(ctx, "request_id", "requestID", "x-request-id"); requestID != "" {
		fields = append(fields, zap.String("request_id", requestID))
	}
	
	// 提取用户 ID
	if userID := getContextValue(ctx, "user_id", "userID", "user"); userID != "" {
		fields = append(fields, zap.String("user_id", userID))
	}
	
	// 提取会话 ID
	if sessionID := getContextValue(ctx, "session_id", "sessionID", "session"); sessionID != "" {
		fields = append(fields, zap.String("session_id", sessionID))
	}
	
	// 提取租户 ID
	if tenantID := getContextValue(ctx, "tenant_id", "tenantID", "tenant"); tenantID != "" {
		fields = append(fields, zap.String("tenant_id", tenantID))
	}
	
	// 提取客户端 IP
	if clientIP := getContextValue(ctx, "client_ip", "clientIP", "ip", "remote_addr"); clientIP != "" {
		fields = append(fields, zap.String("client_ip", clientIP))
	}
	
	// 提取用户代理
	if userAgent := getContextValue(ctx, "user_agent", "userAgent", "User-Agent"); userAgent != "" {
		fields = append(fields, zap.String("user_agent", userAgent))
	}

	return fields
}

// getContextValue 从上下文中获取值，支持多个键名
func getContextValue(ctx context.Context, keys ...string) string {
	for _, key := range keys {
		if val := ctx.Value(key); val != nil {
			if str, ok := val.(string); ok && str != "" {
				return str
			}
		}
	}
	return ""
}

// WithTraceContext 为日志器添加追踪上下文
func WithTraceContext(ctx context.Context, logger Logger) Logger {
	extractor := NewTracingExtractor(true)
	fields := extractor.ExtractAllContextFields(ctx)
	
	if len(fields) == 0 {
		return logger
	}
	
	return logger.WithFields(fields...)
}

// CreateTracedLogger 创建带有追踪支持的日志器
func CreateTracedLogger(config LoggerConfig) (Logger, error) {
	// 使用现有的 CreateLoggerWithOutputs 或 NewLogger
	var logger Logger
	var err error

	// 如果配置了多输出，使用多输出创建器
	if hasMultipleOutputs(config) {
		logger, err = CreateLoggerWithOutputs(config)
	} else {
		logger, err = NewLogger(config)
	}

	if err != nil {
		return nil, err
	}

	// 包装为支持追踪的日志器
	return &tracedLogger{
		logger:    logger,
		extractor: NewTracingExtractor(config.Tracing.Enabled),
	}, nil
}

// hasMultipleOutputs 检查是否配置了多种输出
func hasMultipleOutputs(config LoggerConfig) bool {
	outputCount := 0
	
	if config.Output.Console.Enabled || config.Format == "console" || config.Format == "" {
		outputCount++
	}
	
	if config.Output.File.Enabled && config.Output.File.Path != "" {
		outputCount++
	}
	
	if config.Output.Remote.Enabled && config.Output.Remote.Endpoint != "" {
		outputCount++
	}
	
	return outputCount > 1
}

// tracedLogger 支持追踪的日志器包装器
type tracedLogger struct {
	logger    Logger
	extractor *TracingExtractor
}

// Debug 实现 Logger 接口
func (t *tracedLogger) Debug(msg string, fields ...zap.Field) {
	t.logger.Debug(msg, fields...)
}

func (t *tracedLogger) Info(msg string, fields ...zap.Field) {
	t.logger.Info(msg, fields...)
}

func (t *tracedLogger) Warn(msg string, fields ...zap.Field) {
	t.logger.Warn(msg, fields...)
}

func (t *tracedLogger) Error(msg string, fields ...zap.Field) {
	t.logger.Error(msg, fields...)
}

func (t *tracedLogger) Fatal(msg string, fields ...zap.Field) {
	t.logger.Fatal(msg, fields...)
}

// 上下文日志方法 - 这是核心功能，自动提取追踪信息
func (t *tracedLogger) DebugContext(ctx context.Context, msg string, fields ...zap.Field) {
	allFields := t.mergeWithContextFields(ctx, fields...)
	t.logger.Debug(msg, allFields...)
}

func (t *tracedLogger) InfoContext(ctx context.Context, msg string, fields ...zap.Field) {
	allFields := t.mergeWithContextFields(ctx, fields...)
	t.logger.Info(msg, allFields...)
}

func (t *tracedLogger) WarnContext(ctx context.Context, msg string, fields ...zap.Field) {
	allFields := t.mergeWithContextFields(ctx, fields...)
	t.logger.Warn(msg, allFields...)
}

func (t *tracedLogger) ErrorContext(ctx context.Context, msg string, fields ...zap.Field) {
	allFields := t.mergeWithContextFields(ctx, fields...)
	t.logger.Error(msg, allFields...)
}

// mergeWithContextFields 合并上下文字段和传入字段
func (t *tracedLogger) mergeWithContextFields(ctx context.Context, fields ...zap.Field) []zap.Field {
	contextFields := t.extractor.ExtractAllContextFields(ctx)
	
	// 如果没有上下文字段，直接返回原字段
	if len(contextFields) == 0 {
		return fields
	}
	
	// 合并字段，传入的字段优先级更高
	allFields := make([]zap.Field, 0, len(contextFields)+len(fields))
	allFields = append(allFields, contextFields...)
	allFields = append(allFields, fields...)
	
	return allFields
}

func (t *tracedLogger) WithFields(fields ...zap.Field) Logger {
	return &tracedLogger{
		logger:    t.logger.WithFields(fields...),
		extractor: t.extractor,
	}
}

func (t *tracedLogger) WithContext(ctx context.Context) Logger {
	// 提取上下文字段并创建新的日志器
	contextFields := t.extractor.ExtractAllContextFields(ctx)
	return &tracedLogger{
		logger:    t.logger.WithFields(contextFields...),
		extractor: t.extractor,
	}
}

func (t *tracedLogger) WithService(service string) Logger {
	return &tracedLogger{
		logger:    t.logger.WithService(service),
		extractor: t.extractor,
	}
}

func (t *tracedLogger) IfDebug() ConditionalLogger {
	return t.logger.IfDebug()
}

func (t *tracedLogger) IfInfo() ConditionalLogger {
	return t.logger.IfInfo()
}

func (t *tracedLogger) SetLevel(level Level) {
	t.logger.SetLevel(level)
}

func (t *tracedLogger) Sync() error {
	return t.logger.Sync()
}

// TraceableContext 可追踪的上下文接口
type TraceableContext interface {
	// WithTraceID 添加追踪ID
	WithTraceID(traceID string) context.Context
	
	// WithSpanID 添加SpanID  
	WithSpanID(spanID string) context.Context
	
	// WithRequestID 添加请求ID
	WithRequestID(requestID string) context.Context
	
	// WithUserID 添加用户ID
	WithUserID(userID string) context.Context
}

// contextWithValues 带值的上下文实现
type contextWithValues struct {
	context.Context
	values map[string]string
}

// NewTraceableContext 创建可追踪的上下文
func NewTraceableContext(parent context.Context) TraceableContext {
	return &contextWithValues{
		Context: parent,
		values:  make(map[string]string),
	}
}

func (c *contextWithValues) WithTraceID(traceID string) context.Context {
	return context.WithValue(c.Context, "trace_id", traceID)
}

func (c *contextWithValues) WithSpanID(spanID string) context.Context {
	return context.WithValue(c.Context, "span_id", spanID)
}

func (c *contextWithValues) WithRequestID(requestID string) context.Context {
	return context.WithValue(c.Context, "request_id", requestID)
}

func (c *contextWithValues) WithUserID(userID string) context.Context {
	return context.WithValue(c.Context, "user_id", userID)
}

// GenerateTraceID 生成追踪ID（简化版本，生产环境应使用更规范的实现）
func GenerateTraceID() string {
	// 这是一个简化实现，实际应该使用 OpenTelemetry 的 trace ID 生成器
	return generateRandomHex(32) // 32字符的十六进制字符串
}

// GenerateSpanID 生成SpanID
func GenerateSpanID() string {
	return generateRandomHex(16) // 16字符的十六进制字符串
}

// generateRandomHex 生成随机十六进制字符串
func generateRandomHex(length int) string {
	const charset = "0123456789abcdef"
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[len(charset)/2] // 简化实现，实际应该使用随机数
	}
	return string(b)
}

// IsTracingEnabled 检查追踪是否启用
func IsTracingEnabled(config LoggerConfig) bool {
	return config.Tracing.Enabled
}

// SanitizeTraceValue 清理追踪值（去除敏感信息）
func SanitizeTraceValue(value string) string {
	// 追踪ID通常不包含敏感信息，但可以在这里添加额外的清理逻辑
	return strings.TrimSpace(value)
}