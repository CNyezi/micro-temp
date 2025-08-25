package logger

import (
	"context"
	"fmt"
	"strings"
	"time"

	"connectrpc.com/connect"
	"go.uber.org/zap"
)

// ConnectLoggingInterceptor Connect RPC 日志拦截器
type ConnectLoggingInterceptor struct {
	logger Logger
	config MiddlewareConfig
}

// 确保 ConnectLoggingInterceptor 实现 connect.Interceptor 接口
var _ connect.Interceptor = (*ConnectLoggingInterceptor)(nil)

// MiddlewareConfig 中间件配置
type MiddlewareConfig struct {
	LogRequests     bool     // 记录请求
	LogResponses    bool     // 记录响应
	LogHeaders      bool     // 记录头部信息
	SensitiveFields []string // 敏感字段列表
	MaxBodySize     int      // 最大请求体大小（字节）
	SlowThreshold   int      // 慢请求阈值（毫秒）
}

// NewConnectLoggingInterceptor 创建 Connect 日志拦截器
func NewConnectLoggingInterceptor(logger Logger, config MiddlewareConfig) *ConnectLoggingInterceptor {
	// 设置默认值
	if config.MaxBodySize <= 0 {
		config.MaxBodySize = 4096 // 4KB
	}
	if config.SlowThreshold <= 0 {
		config.SlowThreshold = 1000 // 1秒
	}

	return &ConnectLoggingInterceptor{
		logger: logger,
		config: config,
	}
}

// WrapUnary 包装一元 RPC 调用
func (i *ConnectLoggingInterceptor) WrapUnary(next connect.UnaryFunc) connect.UnaryFunc {
	return func(ctx context.Context, req connect.AnyRequest) (connect.AnyResponse, error) {
		startTime := time.Now()
		
		// 提取基础信息
		procedure := req.Spec().Procedure
		baseFields := []zap.Field{
			zap.String("method", "unary"),
			zap.String("procedure", procedure),
			zap.Time("start_time", startTime),
		}

		// 添加请求信息
		if i.config.LogRequests {
			reqFields := i.extractRequestFields(req)
			baseFields = append(baseFields, reqFields...)
		}

		// 记录请求开始
		i.logger.InfoContext(ctx, "RPC request started", baseFields...)

		// 执行实际调用
		resp, err := next(ctx, req)
		
		// 计算耗时
		duration := time.Since(startTime)
		
		// 构建响应日志字段
		responseFields := []zap.Field{
			zap.String("procedure", procedure),
			zap.Duration("duration", duration),
			zap.Int64("duration_ms", duration.Milliseconds()),
		}

		// 添加响应信息
		if i.config.LogResponses && resp != nil {
			respFields := i.extractResponseFields(resp)
			responseFields = append(responseFields, respFields...)
		}

		// 记录结果
		if err != nil {
			// 错误情况
			errorFields := append(responseFields, 
				zap.Error(err),
				zap.String("status", "error"),
			)
			
			// 提取 Connect 错误信息
			if connectErr, ok := err.(*connect.Error); ok {
				errorFields = append(errorFields,
					zap.String("error_code", connectErr.Code().String()),
					zap.String("error_message", connectErr.Message()),
				)
			}

			i.logger.ErrorContext(ctx, "RPC request failed", errorFields...)
		} else {
			// 成功情况
			successFields := append(responseFields, zap.String("status", "success"))
			
			// 判断是否为慢请求
			logLevel := "info"
			logMsg := "RPC request completed"
			if duration.Milliseconds() > int64(i.config.SlowThreshold) {
				logLevel = "warn"
				logMsg = "RPC request completed (slow)"
				successFields = append(successFields, zap.Bool("slow_request", true))
			}

			if logLevel == "warn" {
				i.logger.WarnContext(ctx, logMsg, successFields...)
			} else {
				i.logger.InfoContext(ctx, logMsg, successFields...)
			}
		}

		return resp, err
	}
}

// WrapStreamingClient 包装流式客户端调用
func (i *ConnectLoggingInterceptor) WrapStreamingClient(next connect.StreamingClientFunc) connect.StreamingClientFunc {
	return func(ctx context.Context, spec connect.Spec) connect.StreamingClientConn {
		startTime := time.Now()
		
		baseFields := []zap.Field{
			zap.String("method", "streaming_client"),
			zap.String("procedure", spec.Procedure),
			zap.Time("start_time", startTime),
		}

		i.logger.InfoContext(ctx, "Streaming client started", baseFields...)
		
		conn := next(ctx, spec)
		
		// 包装连接以记录流式操作
		return &wrappedStreamingClientConn{
			StreamingClientConn: conn,
			logger:              i.logger,
			procedure:           spec.Procedure,
			startTime:           startTime,
		}
	}
}

// WrapStreamingHandler 包装流式处理器
func (i *ConnectLoggingInterceptor) WrapStreamingHandler(next connect.StreamingHandlerFunc) connect.StreamingHandlerFunc {
	return func(ctx context.Context, conn connect.StreamingHandlerConn) error {
		startTime := time.Now()
		
		baseFields := []zap.Field{
			zap.String("method", "streaming_handler"),
			zap.String("procedure", conn.Spec().Procedure),
			zap.Time("start_time", startTime),
		}

		i.logger.InfoContext(ctx, "Streaming handler started", baseFields...)
		
		err := next(ctx, conn)
		
		duration := time.Since(startTime)
		
		responseFields := []zap.Field{
			zap.String("procedure", conn.Spec().Procedure),
			zap.Duration("duration", duration),
		}

		if err != nil {
			responseFields = append(responseFields, zap.Error(err), zap.String("status", "error"))
			i.logger.ErrorContext(ctx, "Streaming handler failed", responseFields...)
		} else {
			responseFields = append(responseFields, zap.String("status", "success"))
			i.logger.InfoContext(ctx, "Streaming handler completed", responseFields...)
		}

		return err
	}
}

// extractRequestFields 提取请求字段
func (i *ConnectLoggingInterceptor) extractRequestFields(req connect.AnyRequest) []zap.Field {
	var fields []zap.Field

	// 添加头部信息
	if i.config.LogHeaders {
		headers := make(map[string]string)
		for key, values := range req.Header() {
			// 检查是否为敏感头部
			if !i.isSensitiveField(key) {
				headers[key] = strings.Join(values, ",")
			} else {
				headers[key] = "[REDACTED]"
			}
		}
		if len(headers) > 0 {
			fields = append(fields, zap.Any("request_headers", headers))
		}
	}

	// 添加请求大小（如果有消息体）
	if req.Any() != nil {
		// 这里简化处理，实际应该根据具体的消息类型来计算大小
		fields = append(fields, zap.String("request_type", fmt.Sprintf("%T", req.Any())))
	}

	return fields
}

// extractResponseFields 提取响应字段
func (i *ConnectLoggingInterceptor) extractResponseFields(resp connect.AnyResponse) []zap.Field {
	var fields []zap.Field

	// 添加头部信息
	if i.config.LogHeaders {
		headers := make(map[string]string)
		for key, values := range resp.Header() {
			if !i.isSensitiveField(key) {
				headers[key] = strings.Join(values, ",")
			} else {
				headers[key] = "[REDACTED]"
			}
		}
		if len(headers) > 0 {
			fields = append(fields, zap.Any("response_headers", headers))
		}
	}

	// 添加响应类型
	if resp.Any() != nil {
		fields = append(fields, zap.String("response_type", fmt.Sprintf("%T", resp.Any())))
	}

	return fields
}

// isSensitiveField 检查字段是否敏感
func (i *ConnectLoggingInterceptor) isSensitiveField(fieldName string) bool {
	lowerField := strings.ToLower(fieldName)
	
	// 检查配置的敏感字段
	for _, sensitive := range i.config.SensitiveFields {
		if strings.ToLower(sensitive) == lowerField {
			return true
		}
	}
	
	// 检查常见的敏感头部
	commonSensitive := []string{
		"authorization", "cookie", "x-api-key", "x-auth-token",
		"authentication", "x-access-token", "bearer", "password",
	}
	
	for _, sensitive := range commonSensitive {
		if lowerField == sensitive || strings.Contains(lowerField, sensitive) {
			return true
		}
	}
	
	return false
}

// wrappedStreamingClientConn 包装的流式客户端连接
type wrappedStreamingClientConn struct {
	connect.StreamingClientConn
	logger    Logger
	procedure string
	startTime time.Time
}

func (w *wrappedStreamingClientConn) CloseRequest() error {
	err := w.StreamingClientConn.CloseRequest()
	if err != nil {
		w.logger.Error("Failed to close streaming request",
			zap.String("procedure", w.procedure),
			zap.Error(err),
		)
	}
	return err
}

func (w *wrappedStreamingClientConn) CloseResponse() error {
	err := w.StreamingClientConn.CloseResponse()
	duration := time.Since(w.startTime)
	
	fields := []zap.Field{
		zap.String("procedure", w.procedure),
		zap.Duration("total_duration", duration),
	}
	
	if err != nil {
		fields = append(fields, zap.Error(err), zap.String("status", "error"))
		w.logger.Error("Streaming client connection closed with error", fields...)
	} else {
		fields = append(fields, zap.String("status", "success"))
		w.logger.Info("Streaming client connection closed", fields...)
	}
	
	return err
}

// InterceptorOption 拦截器配置选项
type InterceptorOption func(*MiddlewareConfig)

// WithRequestLogging 启用请求日志记录
func WithRequestLogging(enabled bool) InterceptorOption {
	return func(config *MiddlewareConfig) {
		config.LogRequests = enabled
	}
}

// WithResponseLogging 启用响应日志记录
func WithResponseLogging(enabled bool) InterceptorOption {
	return func(config *MiddlewareConfig) {
		config.LogResponses = enabled
	}
}

// WithHeaderLogging 启用头部日志记录
func WithHeaderLogging(enabled bool) InterceptorOption {
	return func(config *MiddlewareConfig) {
		config.LogHeaders = enabled
	}
}

// WithSensitiveFields 设置敏感字段
func WithSensitiveFields(fields []string) InterceptorOption {
	return func(config *MiddlewareConfig) {
		config.SensitiveFields = fields
	}
}

// WithMaxBodySize 设置最大请求体大小
func WithMaxBodySize(size int) InterceptorOption {
	return func(config *MiddlewareConfig) {
		config.MaxBodySize = size
	}
}

// WithSlowThreshold 设置慢请求阈值
func WithSlowThreshold(ms int) InterceptorOption {
	return func(config *MiddlewareConfig) {
		config.SlowThreshold = ms
	}
}

// NewConnectLoggingInterceptorWithOptions 使用选项创建拦截器
func NewConnectLoggingInterceptorWithOptions(logger Logger, options ...InterceptorOption) *ConnectLoggingInterceptor {
	config := MiddlewareConfig{
		LogRequests:   true,
		LogResponses:  true,
		LogHeaders:    false, // 默认不记录头部以保护隐私
		MaxBodySize:   4096,
		SlowThreshold: 1000,
	}

	for _, option := range options {
		option(&config)
	}

	return NewConnectLoggingInterceptor(logger, config)
}

// DefaultMiddlewareConfig 默认中间件配置
func DefaultMiddlewareConfig() MiddlewareConfig {
	return MiddlewareConfig{
		LogRequests:   true,
		LogResponses:  true,
		LogHeaders:    false,
		MaxBodySize:   4096,  // 4KB
		SlowThreshold: 1000,  // 1秒
		SensitiveFields: []string{
			"password", "token", "key", "secret", "auth",
		},
	}
}

// CreateInterceptors 创建 Connect 拦截器选项
func CreateInterceptors(logger Logger, config MiddlewareConfig) []connect.Option {
	interceptor := NewConnectLoggingInterceptor(logger, config)
	
	return []connect.Option{
		connect.WithInterceptors(interceptor),
	}
}

// 注意：WrapUnary、WrapStreamingClient、WrapStreamingHandler 方法已经在上面定义
// 它们构成了 connect.Interceptor 接口的实现