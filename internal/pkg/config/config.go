package config

import (
	"fmt"
	"strings"

	"github.com/spf13/viper"
)

type Config struct {
	Server   ServerConfig   `mapstructure:"server"`
	Database DatabaseConfig `mapstructure:"database"`
	Redis    RedisConfig    `mapstructure:"redis"`
	Log      LogConfig      `mapstructure:"log"`
}

type ServerConfig struct {
	Host string `mapstructure:"host"`
	Port int    `mapstructure:"port"`
}

type DatabaseConfig struct {
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	User     string `mapstructure:"user"`
	Password string `mapstructure:"password"`
	Database string `mapstructure:"database"`
	SSLMode  string `mapstructure:"ssl_mode"`
}

type RedisConfig struct {
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	Password string `mapstructure:"password"`
	DB       int    `mapstructure:"db"`
}

type LogConfig struct {
	// 基础配置 - 保持向后兼容性
	Level  string `mapstructure:"level"`
	Format string `mapstructure:"format"`
	
	// 新增配置
	Output      OutputConfig      `mapstructure:"output"`
	Tracing     TracingConfig     `mapstructure:"tracing"`
	Middleware  MiddlewareConfig  `mapstructure:"middleware"`
	Performance PerformanceConfig `mapstructure:"performance"`
}

// OutputConfig 输出配置
type OutputConfig struct {
	Console ConsoleConfig `mapstructure:"console"`
	File    FileConfig    `mapstructure:"file"`
	Remote  RemoteConfig  `mapstructure:"remote"`
}

// ConsoleConfig 控制台输出配置
type ConsoleConfig struct {
	Enabled    bool `mapstructure:"enabled"`
	Colorized  bool `mapstructure:"colorized"`
	TimeFormat string `mapstructure:"time_format"`
}

// FileConfig 文件输出配置
type FileConfig struct {
	Enabled    bool   `mapstructure:"enabled"`
	Path       string `mapstructure:"path"`
	MaxSize    int    `mapstructure:"max_size_mb"`
	MaxBackups int    `mapstructure:"max_backups"`
	MaxAge     int    `mapstructure:"max_age_days"`
	Compress   bool   `mapstructure:"compress"`
}

// RemoteConfig 远程输出配置
type RemoteConfig struct {
	Enabled   bool   `mapstructure:"enabled"`
	Endpoint  string `mapstructure:"endpoint"`
	Protocol  string `mapstructure:"protocol"` // http, grpc, tcp
	BatchSize int    `mapstructure:"batch_size"`
	Timeout   int    `mapstructure:"timeout_ms"`
	TLS       bool   `mapstructure:"tls"`
}

// TracingConfig 追踪配置
type TracingConfig struct {
	Enabled     bool   `mapstructure:"enabled"`
	ServiceName string `mapstructure:"service_name"`
	Version     string `mapstructure:"version"`
	Environment string `mapstructure:"environment"`
}

// MiddlewareConfig 中间件配置
type MiddlewareConfig struct {
	LogRequests     bool     `mapstructure:"log_requests"`
	LogResponses    bool     `mapstructure:"log_responses"`
	LogHeaders      bool     `mapstructure:"log_headers"`
	SensitiveFields []string `mapstructure:"sensitive_fields"`
	MaxBodySize     int      `mapstructure:"max_body_size"`
}

// PerformanceConfig 性能配置
type PerformanceConfig struct {
	AsyncWrite    bool `mapstructure:"async_write"`
	BufferSize    int  `mapstructure:"buffer_size"`
	FlushInterval int  `mapstructure:"flush_interval_ms"`
}

func Load(configPath string) (*Config, error) {
	v := viper.New()

	v.SetConfigFile(configPath)
	v.SetConfigType("yaml")

	v.SetEnvPrefix("PIGEON")
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()

	if err := v.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var config Config
	if err := v.Unmarshal(&config); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	return &config, nil
}

func (c *DatabaseConfig) DSN() string {
	return fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		c.Host, c.Port, c.User, c.Password, c.Database, c.SSLMode)
}