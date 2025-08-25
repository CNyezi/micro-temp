package logger

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// 使用 zap 的现有格式化能力，只提供配置便捷函数

// CreateZapConfig 创建 zap 配置
func CreateZapConfig(level Level, format string) zap.Config {
	var config zap.Config

	if format == "json" {
		config = zap.NewProductionConfig()
		config.EncoderConfig = zap.NewProductionEncoderConfig()
		config.EncoderConfig.TimeKey = "timestamp"
		config.EncoderConfig.LevelKey = "level" 
		config.EncoderConfig.MessageKey = "message"
		config.EncoderConfig.CallerKey = "caller"
		config.EncoderConfig.StacktraceKey = "stacktrace"
		config.EncoderConfig.EncodeTime = zapcore.RFC3339TimeEncoder
		config.EncoderConfig.EncodeLevel = zapcore.LowercaseLevelEncoder
	} else {
		config = zap.NewDevelopmentConfig()
		config.EncoderConfig = zap.NewDevelopmentEncoderConfig()
		config.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
		config.EncoderConfig.EncodeTime = zapcore.TimeEncoderOfLayout("2006-01-02 15:04:05")
		config.EncoderConfig.EncodeCaller = zapcore.ShortCallerEncoder
	}

	config.Level = zap.NewAtomicLevelAt(levelToZapLevel(level))
	config.OutputPaths = []string{"stdout"}
	config.ErrorOutputPaths = []string{"stderr"}
	config.DisableCaller = false
	config.DisableStacktrace = false

	return config
}

// CreateZapEncoder 创建 zap 编码器
func CreateZapEncoder(format string) zapcore.Encoder {
	if format == "json" {
		encoderConfig := zapcore.EncoderConfig{
			TimeKey:        "timestamp",
			LevelKey:       "level",
			NameKey:        "logger",
			CallerKey:      "caller",
			MessageKey:     "message",
			StacktraceKey:  "stacktrace",
			LineEnding:     zapcore.DefaultLineEnding,
			EncodeLevel:    zapcore.LowercaseLevelEncoder,
			EncodeTime:     zapcore.RFC3339TimeEncoder,
			EncodeDuration: zapcore.StringDurationEncoder,
			EncodeCaller:   zapcore.ShortCallerEncoder,
		}
		return zapcore.NewJSONEncoder(encoderConfig)
	}
	
	encoderConfig := zapcore.EncoderConfig{
		TimeKey:        "T",
		LevelKey:       "L",
		NameKey:        "N",
		CallerKey:      "C",
		MessageKey:     "M",
		StacktraceKey:  "S",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.CapitalColorLevelEncoder,
		EncodeTime:     zapcore.TimeEncoderOfLayout("2006-01-02 15:04:05"),
		EncodeDuration: zapcore.StringDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	}
	return zapcore.NewConsoleEncoder(encoderConfig)
}

// CreateCore 创建 zapcore.Core
func CreateCore(encoder zapcore.Encoder, writer zapcore.WriteSyncer, level zapcore.Level) zapcore.Core {
	return zapcore.NewCore(encoder, writer, level)
}

