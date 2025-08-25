package logger

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

// OutputManager 管理多种输出目标
type OutputManager struct {
	cores []zapcore.Core
	mutex sync.RWMutex
}

// NewOutputManager 创建输出管理器
func NewOutputManager() *OutputManager {
	return &OutputManager{
		cores: make([]zapcore.Core, 0),
	}
}

// AddConsoleOutput 添加控制台输出
func (om *OutputManager) AddConsoleOutput(level zapcore.Level, encoder zapcore.Encoder) {
	om.mutex.Lock()
	defer om.mutex.Unlock()

	// 使用标准输出
	writer := zapcore.AddSync(os.Stdout)
	core := zapcore.NewCore(encoder, writer, level)
	om.cores = append(om.cores, core)
}

// AddFileOutput 添加文件输出（带轮转）
func (om *OutputManager) AddFileOutput(config InternalFileOutputConfig, level zapcore.Level, encoder zapcore.Encoder) error {
	om.mutex.Lock()
	defer om.mutex.Unlock()

	// 确保目录存在
	dir := filepath.Dir(config.Filename)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create log directory: %w", err)
	}

	// 使用 lumberjack 进行日志轮转
	lumberjackLogger := &lumberjack.Logger{
		Filename:   config.Filename,
		MaxSize:    config.MaxSizeMB,    // MB
		MaxBackups: config.MaxBackups,   // 保留的旧文件数量
		MaxAge:     config.MaxAgeDays,   // 保留的天数
		Compress:   config.Compress,     // 是否压缩
	}

	writer := zapcore.AddSync(lumberjackLogger)
	core := zapcore.NewCore(encoder, writer, level)
	om.cores = append(om.cores, core)

	return nil
}

// AddRemoteOutput 添加远程输出（预留接口）
func (om *OutputManager) AddRemoteOutput(config InternalRemoteOutputConfig, level zapcore.Level, encoder zapcore.Encoder) error {
	om.mutex.Lock()
	defer om.mutex.Unlock()

	// TODO: 实现远程日志输出
	// 这里可以集成如 syslog、fluentd、elasticsearch 等
	
	return fmt.Errorf("remote output not implemented yet")
}

// CreateTeeCore 创建组合的 Core
func (om *OutputManager) CreateTeeCore() zapcore.Core {
	om.mutex.RLock()
	defer om.mutex.RUnlock()

	if len(om.cores) == 0 {
		// 如果没有配置任何输出，默认使用控制台
		encoder := CreateZapEncoder("console")
		writer := zapcore.AddSync(os.Stdout)
		return zapcore.NewCore(encoder, writer, zapcore.InfoLevel)
	}

	if len(om.cores) == 1 {
		return om.cores[0]
	}

	return zapcore.NewTee(om.cores...)
}

// InternalFileOutputConfig 内部文件输出配置
type InternalFileOutputConfig struct {
	Filename   string
	MaxSizeMB  int
	MaxBackups int
	MaxAgeDays int
	Compress   bool
}

// InternalRemoteOutputConfig 内部远程输出配置
type InternalRemoteOutputConfig struct {
	Type     string // syslog, http, grpc, tcp
	Endpoint string
	Protocol string
	Timeout  int
	TLS      bool
	BatchSize int
}

// CreateLoggerWithOutputs 根据配置创建带有多输出的日志器
func CreateLoggerWithOutputs(config LoggerConfig) (Logger, error) {
	// 对于复杂的多输出配置，我们直接构建多输出的 Core
	
	// 构建多输出的 Core
	outputManager := NewOutputManager()
	level := levelToZapLevel(config.Level)

	// 添加控制台输出
	if shouldAddConsoleOutput(config) {
		encoder := CreateZapEncoder(config.Format)
		outputManager.AddConsoleOutput(level, encoder)
	}

	// 添加文件输出
	if shouldAddFileOutput(config) {
		encoder := CreateZapEncoder("json") // 文件通常使用JSON格式
		fileConfig := createFileOutputConfig(config)
		if err := outputManager.AddFileOutput(fileConfig, level, encoder); err != nil {
			return nil, fmt.Errorf("failed to add file output: %w", err)
		}
	}

	// 添加远程输出
	if shouldAddRemoteOutput(config) {
		encoder := CreateZapEncoder("json") // 远程输出通常使用JSON格式
		remoteConfig := createRemoteOutputConfig(config)
		if err := outputManager.AddRemoteOutput(remoteConfig, level, encoder); err != nil {
			// 远程输出失败不应该阻止日志器创建，只记录警告
			fmt.Fprintf(os.Stderr, "Warning: failed to add remote output: %v\n", err)
		}
	}

	// 创建组合的 Core
	core := outputManager.CreateTeeCore()
	
	// 创建 zap logger
	zapInst := zap.New(core, zap.AddCaller(), zap.AddStacktrace(zapcore.ErrorLevel))

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

	if len(baseFields) > 0 {
		zapInst = zapInst.With(baseFields...)
	}

	// 创建我们的日志器包装
	logger := &zapLoggerInternal{
		zap:        zapInst,
		sugar:      zapInst.Sugar(),
		level:      config.Level,
		service:    config.ServiceName,
		baseFields: baseFields,
	}

	return logger, nil
}

// zapLoggerInternal 内部使用的 zap 日志器包装器
// 这是为了避免与 logger.go 中的 zapLogger 类型冲突
type zapLoggerInternal struct {
	zap        *zap.Logger
	sugar      *zap.SugaredLogger
	level      Level
	service    string
	baseFields []zap.Field
}

// 实现 Logger 接口的所有方法
func (l *zapLoggerInternal) Debug(msg string, fields ...zap.Field) {
	l.zap.Debug(msg, fields...)
}

func (l *zapLoggerInternal) Info(msg string, fields ...zap.Field) {
	l.zap.Info(msg, fields...)
}

func (l *zapLoggerInternal) Warn(msg string, fields ...zap.Field) {
	l.zap.Warn(msg, fields...)
}

func (l *zapLoggerInternal) Error(msg string, fields ...zap.Field) {
	l.zap.Error(msg, fields...)
}

func (l *zapLoggerInternal) Fatal(msg string, fields ...zap.Field) {
	l.zap.Fatal(msg, fields...)
}

func (l *zapLoggerInternal) DebugContext(ctx context.Context, msg string, fields ...zap.Field) {
	// 简化实现，直接调用 Debug
	l.Debug(msg, fields...)
}

func (l *zapLoggerInternal) InfoContext(ctx context.Context, msg string, fields ...zap.Field) {
	l.Info(msg, fields...)
}

func (l *zapLoggerInternal) WarnContext(ctx context.Context, msg string, fields ...zap.Field) {
	l.Warn(msg, fields...)
}

func (l *zapLoggerInternal) ErrorContext(ctx context.Context, msg string, fields ...zap.Field) {
	l.Error(msg, fields...)
}

func (l *zapLoggerInternal) WithFields(fields ...zap.Field) Logger {
	return &zapLoggerInternal{
		zap:        l.zap.With(fields...),
		sugar:      l.sugar,
		level:      l.level,
		service:    l.service,
		baseFields: l.baseFields,
	}
}

func (l *zapLoggerInternal) WithContext(ctx context.Context) Logger {
	return l // 简化实现
}

func (l *zapLoggerInternal) WithService(service string) Logger {
	return &zapLoggerInternal{
		zap:        l.zap.With(zap.String("service", service)),
		sugar:      l.sugar,
		level:      l.level,
		service:    service,
		baseFields: l.baseFields,
	}
}

func (l *zapLoggerInternal) IfDebug() ConditionalLogger {
	if l.level <= DebugLevel {
		return &conditionalLoggerInternal{logger: l, level: DebugLevel}
	}
	return &noopConditionalLoggerInternal{}
}

func (l *zapLoggerInternal) IfInfo() ConditionalLogger {
	if l.level <= InfoLevel {
		return &conditionalLoggerInternal{logger: l, level: InfoLevel}
	}
	return &noopConditionalLoggerInternal{}
}

func (l *zapLoggerInternal) SetLevel(level Level) {
	l.level = level
}

func (l *zapLoggerInternal) Sync() error {
	return l.zap.Sync()
}

// 内部条件日志器
type conditionalLoggerInternal struct {
	logger Logger
	level  Level
}

func (c *conditionalLoggerInternal) Log(msg string, fields ...zap.Field) {
	switch c.level {
	case DebugLevel:
		c.logger.Debug(msg, fields...)
	case InfoLevel:
		c.logger.Info(msg, fields...)
	}
}

func (c *conditionalLoggerInternal) Logf(format string, args ...any) {
	msg := fmt.Sprintf(format, args...)
	c.Log(msg)
}

// 内部空操作条件日志器
type noopConditionalLoggerInternal struct{}

func (n *noopConditionalLoggerInternal) Log(msg string, fields ...zap.Field) {}
func (n *noopConditionalLoggerInternal) Logf(format string, args ...any)    {}

// shouldAddConsoleOutput 判断是否应该添加控制台输出
func shouldAddConsoleOutput(config LoggerConfig) bool {
	// 默认情况下，开发环境和格式为console时启用控制台输出
	if config.Format == "console" || config.Format == "" {
		return true
	}
	
	// 检查显式配置
	return config.Output.Console.Enabled
}

// shouldAddFileOutput 判断是否应该添加文件输出
func shouldAddFileOutput(config LoggerConfig) bool {
	return config.Output.File.Enabled && config.Output.File.Path != ""
}

// shouldAddRemoteOutput 判断是否应该添加远程输出
func shouldAddRemoteOutput(config LoggerConfig) bool {
	return config.Output.Remote.Enabled && config.Output.Remote.Endpoint != ""
}

// createFileOutputConfig 创建文件输出配置
func createFileOutputConfig(config LoggerConfig) InternalFileOutputConfig {
	fileConfig := config.Output.File
	
	return InternalFileOutputConfig{
		Filename:   fileConfig.Path,
		MaxSizeMB:  fileConfig.MaxSize,
		MaxBackups: fileConfig.MaxBackups,
		MaxAgeDays: fileConfig.MaxAge,
		Compress:   fileConfig.Compress,
	}
}

// createRemoteOutputConfig 创建远程输出配置
func createRemoteOutputConfig(config LoggerConfig) InternalRemoteOutputConfig {
	remoteConfig := config.Output.Remote
	
	return InternalRemoteOutputConfig{
		Type:      remoteConfig.Protocol,
		Endpoint:  remoteConfig.Endpoint,
		Protocol:  remoteConfig.Protocol,
		Timeout:   remoteConfig.Timeout,
		TLS:       remoteConfig.TLS,
		BatchSize: remoteConfig.BatchSize,
	}
}

// DefaultInternalFileOutputConfig 默认文件输出配置
func DefaultInternalFileOutputConfig() InternalFileOutputConfig {
	return InternalFileOutputConfig{
		Filename:   "logs/app.log",
		MaxSizeMB:  100,
		MaxBackups: 3,
		MaxAgeDays: 28,
		Compress:   true,
	}
}

// DefaultInternalRemoteOutputConfig 默认远程输出配置
func DefaultInternalRemoteOutputConfig() InternalRemoteOutputConfig {
	return InternalRemoteOutputConfig{
		Type:      "http",
		Protocol:  "http",
		Timeout:   5000, // 5秒
		TLS:       false,
		BatchSize: 100,
	}
}