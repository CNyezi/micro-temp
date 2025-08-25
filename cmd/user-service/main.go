package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"micro-holtye/gen/user/v1/userv1connect"
	"micro-holtye/internal/pkg/config"
	"micro-holtye/internal/pkg/database"
	"micro-holtye/internal/pkg/logger"
	"micro-holtye/internal/service/user"

	"connectrpc.com/connect"
	"go.uber.org/zap"
)

func main() {
	var configPath string
	flag.StringVar(&configPath, "config", "configs/user-service.yaml", "path to config file")
	flag.Parse()

	cfg, err := config.Load(configPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to load config: %v\n", err)
		os.Exit(1)
	}

	// 初始化统一日志组件，启用追踪功能
	loggerConfig := logger.LoggerConfig{
		Level:       logger.InfoLevel,
		Format:      "json",
		ServiceName: "user-service",
		Version:     os.Getenv("SERVICE_VERSION"),
		Environment: getEnvironment(),
		Output: logger.OutputConfig{
			Console: logger.ConsoleOutputConfig{Enabled: true},
			File: logger.FileOutputConfig{
				Enabled:    true,
				Path:       "logs/user-service.log",
				MaxSize:    100, // 100MB
				MaxBackups: 3,
				MaxAge:     7, // 7天
				Compress:   true,
			},
		},
		Tracing: logger.TracingConfig{
			Enabled: true, // 启用追踪功能
		},
	}

	// 创建带追踪功能的日志器
	appLogger, err := logger.CreateTracedLogger(loggerConfig)
	if err != nil {
		log.Fatalf("Failed to create logger: %v", err)
	}

	// 设置为全局日志器
	logger.SetGlobalLogger(appLogger)

	// 使用新的日志器记录启动信息
	appLogger.Info("Initializing user service",
		zap.String("service", "user-service"),
		zap.String("version", os.Getenv("SERVICE_VERSION")),
		zap.String("environment", getEnvironment()),
	)

	db, err := database.NewConnection(cfg.Database.DSN())
	if err != nil {
		appLogger.Fatal("Failed to connect to database", zap.Error(err))
	}
	defer db.Close()

	redisClient, err := database.NewRedisClient(
		cfg.Redis.Host,
		cfg.Redis.Port,
		cfg.Redis.Password,
		cfg.Redis.DB,
	)
	if err != nil {
		appLogger.Fatal("Failed to connect to Redis", zap.Error(err))
	}
	defer redisClient.Close()

	store := user.NewStore(db.DB)
	service := user.NewService(store, appLogger)
	handler := user.NewConnectHandler(service)

	// 创建日志中间件
	middlewareConfig := logger.MiddlewareConfig{
		LogRequests:   true,
		LogResponses:  true,
		LogHeaders:    false, // 减少日志量
		SlowThreshold: 1000,  // 1秒慢请求阈值
		SensitiveFields: []string{
			"authorization", "cookie", "x-api-key", 
			"password", "token", "secret",
		},
	}
	
	loggingInterceptor := logger.NewConnectLoggingInterceptor(appLogger, middlewareConfig)

	// 创建带中间件的处理器
	mux := http.NewServeMux()
	path, userHandler := userv1connect.NewUserServiceHandler(
		handler,
		connect.WithInterceptors(loggingInterceptor),
	)
	mux.Handle(path, userHandler)

	// 添加健康检查端点
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		appLogger.InfoContext(ctx, "Health check requested",
			logger.RequestID(r.Header.Get("X-Request-ID")),
			logger.Component("health-check"),
		)
		
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"healthy","service":"user-service"}`))
	})

	// 创建HTTP服务器
	addr := fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port)
	server := &http.Server{
		Addr:    addr,
		Handler: mux,
	}

	// 启动服务器
	go func() {
		appLogger.Info("Starting user service", 
			zap.String("address", addr),
		)
		
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			appLogger.Fatal("Failed to start server", zap.Error(err))
		}
	}()

	// 优雅关闭
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	appLogger.Info("Shutting down user service...")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		appLogger.Error("Server forced to shutdown", zap.Error(err))
	}

	// 同步日志
	if err := appLogger.Sync(); err != nil {
		log.Printf("Failed to sync logger: %v", err)
	}

	appLogger.Info("User service stopped")
}

// getEnvironment 获取运行环境
func getEnvironment() string {
	env := os.Getenv("ENVIRONMENT")
	if env == "" {
		env = "development"
	}
	return env
}
