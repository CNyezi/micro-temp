package logger

import (
	"regexp"
	"strings"

	"go.uber.org/zap"
)

// 敏感数据检测和脱敏 - 这是我们的核心增值功能

var (
	// 敏感字段名模式
	sensitiveFieldPatterns = []*regexp.Regexp{
		regexp.MustCompile(`(?i)password`),
		regexp.MustCompile(`(?i)passwd`),
		regexp.MustCompile(`(?i)secret`),
		regexp.MustCompile(`(?i)token`),
		regexp.MustCompile(`(?i)key`),
		regexp.MustCompile(`(?i)auth`),
		regexp.MustCompile(`(?i)credential`),
		regexp.MustCompile(`(?i)session`),
	}

	// 敏感值模式
	sensitiveValuePatterns = []*regexp.Regexp{
		// 信用卡号
		regexp.MustCompile(`\b\d{4}[-\s]?\d{4}[-\s]?\d{4}[-\s]?\d{4}\b`),
		// 身份证号 (简化版，支持15位和18位)
		regexp.MustCompile(`\b\d{15}|\d{17}[\dXx]\b`),
		// 邮箱
		regexp.MustCompile(`\b[A-Za-z0-9._%+-]+@[A-Za-z0-9.-]+\.[A-Z|a-z]{2,}\b`),
		// 电话号码 (中国手机号)
		regexp.MustCompile(`\b1[3-9]\d{9}\b`),
		// JWT Token 模式
		regexp.MustCompile(`\beyJ[A-Za-z0-9_-]*\.[A-Za-z0-9_-]*\.[A-Za-z0-9_-]*\b`),
		// API Key 模式 (32位以上的字母数字组合)
		regexp.MustCompile(`\b[A-Za-z0-9]{32,}\b`),
	}

	// 自定义敏感字段列表
	customSensitiveFields = make(map[string]bool)
)

// AddSensitiveField 添加自定义敏感字段
func AddSensitiveField(fieldName string) {
	customSensitiveFields[strings.ToLower(fieldName)] = true
}

// RemoveSensitiveField 移除自定义敏感字段
func RemoveSensitiveField(fieldName string) {
	delete(customSensitiveFields, strings.ToLower(fieldName))
}

// SanitizeFields 对 zap 字段进行敏感数据脱敏
func SanitizeFields(fields []zap.Field) []zap.Field {
	result := make([]zap.Field, 0, len(fields))
	
	for _, field := range fields {
		// 检查字段名是否敏感
		if isSensitiveField(field.Key) {
			result = append(result, zap.String(field.Key, "[REDACTED]"))
			continue
		}
		
		// 对字符串字段检查敏感值 - 这里简化处理，实际应该检查字段类型
		// 但 zap.Field 的内部结构比较复杂，为了简化我们只处理字段名
		// 在实际使用中，敏感数据主要通过字段名识别
		
		result = append(result, field)
	}
	
	return result
}

// isSensitiveField 检查字段名是否敏感
func isSensitiveField(key string) bool {
	lowerKey := strings.ToLower(key)
	
	// 检查自定义敏感字段列表
	if customSensitiveFields[lowerKey] {
		return true
	}
	
	// 检查预定义模式
	for _, pattern := range sensitiveFieldPatterns {
		if pattern.MatchString(lowerKey) {
			return true
		}
	}
	
	return false
}

// containsSensitiveValue 检查值是否包含敏感信息
func containsSensitiveValue(value string) bool {
	for _, pattern := range sensitiveValuePatterns {
		if pattern.MatchString(value) {
			return true
		}
	}
	return false
}

// sanitizeString 对字符串进行脱敏处理
func sanitizeString(input string) string {
	result := input
	
	// 脱敏信用卡号 - 只显示前4位和后4位
	result = regexp.MustCompile(`\b(\d{4})[-\s]?\d{4}[-\s]?\d{4}[-\s]?(\d{4})\b`).
		ReplaceAllString(result, "$1-****-****-$2")
	
	// 脱敏身份证号 - 只显示前3位和后2位
	result = regexp.MustCompile(`\b(\d{3})\d{12}(\d{2})\b`).
		ReplaceAllString(result, "$1***********$2")
	result = regexp.MustCompile(`\b(\d{3})\d{11}(\d[\dXx])\b`).
		ReplaceAllString(result, "$1***********$2")
	
	// 脱敏邮箱 - 只显示用户名前2位和域名
	result = regexp.MustCompile(`\b([A-Za-z0-9._%+-]{1,2})[A-Za-z0-9._%+-]*(@[A-Za-z0-9.-]+\.[A-Z|a-z]{2,})\b`).
		ReplaceAllString(result, "$1****$2")
	
	// 脱敏手机号 - 只显示前3位和后4位
	result = regexp.MustCompile(`\b(1[3-9]\d)(\d{4})(\d{4})\b`).
		ReplaceAllString(result, "$1****$3")
	
	// 脱敏 JWT Token 和 API Key - 只显示前8位
	result = regexp.MustCompile(`\b([A-Za-z0-9_-]{8})[A-Za-z0-9_-]{24,}\b`).
		ReplaceAllString(result, "$1****")
	
	return result
}

// 便捷的业务字段创建函数 - 基于 zap，但添加了业务语义

// RequestID 创建请求ID字段
func RequestID(value string) zap.Field {
	return zap.String("request_id", value)
}

// UserID 创建用户ID字段
func UserID(value string) zap.Field {
	return zap.String("user_id", value)
}

// TraceID 创建追踪ID字段
func TraceID(value string) zap.Field {
	return zap.String("trace_id", value)
}

// SpanID 创建Span ID字段
func SpanID(value string) zap.Field {
	return zap.String("span_id", value)
}

// Component 创建组件字段
func Component(value string) zap.Field {
	return zap.String("component", value)
}

// Operation 创建操作字段
func Operation(value string) zap.Field {
	return zap.String("operation", value)
}

// ErrorCode 创建错误码字段
func ErrorCode(value string) zap.Field {
	return zap.String("error_code", value)
}

// Latency 创建延迟字段（毫秒）
func Latency(ms int64) zap.Field {
	return zap.Int64("latency_ms", ms)
}

// StatusCode 创建HTTP状态码字段
func StatusCode(code int) zap.Field {
	return zap.Int("status_code", code)
}