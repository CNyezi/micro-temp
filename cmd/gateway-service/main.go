package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"micro-holtye/gen/gateway/v1/gatewayv1connect"
	"micro-holtye/internal/pkg/logger"
	"micro-holtye/internal/service/gateway"

	"connectrpc.com/connect"
	"go.uber.org/zap"
)

func main() {
	// 初始化统一日志组件，启用追踪功能
	loggerConfig := logger.LoggerConfig{
		Level:       logger.InfoLevel,
		Format:      "json",
		ServiceName: "gateway-service",
		Version:     os.Getenv("SERVICE_VERSION"),
		Environment: getEnvironment(),
		Output: logger.OutputConfig{
			Console: logger.ConsoleOutputConfig{Enabled: true},
			File: logger.FileOutputConfig{
				Enabled:    true,
				Path:       "logs/gateway-service.log",
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
	appLogger.Info("Initializing gateway service",
		zap.String("service", "gateway-service"),
		zap.String("version", os.Getenv("SERVICE_VERSION")),
		zap.String("environment", getEnvironment()),
	)

	userServiceURL := os.Getenv("USER_SERVICE_URL")
	if userServiceURL == "" {
		userServiceURL = "http://localhost:8080"
	}

	orderServiceURL := os.Getenv("ORDER_SERVICE_URL")
	if orderServiceURL == "" {
		orderServiceURL = "http://localhost:8081"
	}

	serverAddress := os.Getenv("SERVER_ADDRESS")
	if serverAddress == "" {
		serverAddress = ":8082"
	}

	appLogger.Info("Service configuration",
		zap.String("user_service_url", userServiceURL),
		zap.String("order_service_url", orderServiceURL),
		zap.String("server_address", serverAddress),
	)

	// 创建服务组件
	store := gateway.NewStore(userServiceURL, orderServiceURL)
	service := gateway.NewService(store, appLogger) // 传入日志器
	handler := gateway.NewConnectHandler(service, appLogger)

	// 创建日志中间件
	middlewareConfig := logger.MiddlewareConfig{
		LogRequests:   true,
		LogResponses:  true,
		LogHeaders:    false, // 网关通常不记录头部信息以减少日志量
		SlowThreshold: 2000,  // 网关的慢请求阈值设为2秒
		SensitiveFields: []string{
			"authorization", "cookie", "x-api-key", 
			"password", "token", "secret",
		},
	}
	
	loggingInterceptor := logger.NewConnectLoggingInterceptor(appLogger, middlewareConfig)

	// 创建带中间件的处理器
	mux := http.NewServeMux()
	path, h := gatewayv1connect.NewGatewayServiceHandler(
		handler,
		connect.WithInterceptors(loggingInterceptor),
	)
	mux.Handle(path, h)

	// 添加健康检查端点
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		appLogger.InfoContext(ctx, "Health check requested",
			logger.RequestID(r.Header.Get("X-Request-ID")),
			logger.Component("health-check"),
		)
		
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"healthy","service":"gateway-service"}`))
	})

	// 创建HTTP服务器
	server := &http.Server{
		Addr:    serverAddress,
		Handler: mux,
	}

	// 启动服务器
	go func() {
		appLogger.Info("Starting gateway service", 
			zap.String("address", serverAddress),
		)
		
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			appLogger.Fatal("Failed to start server", zap.Error(err))
		}
	}()

	// 优雅关闭
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	appLogger.Info("Shutting down gateway service...")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		appLogger.Error("Server forced to shutdown", zap.Error(err))
	}

	// 同步日志
	if err := appLogger.Sync(); err != nil {
		log.Printf("Failed to sync logger: %v", err)
	}

	appLogger.Info("Gateway service stopped")
}

// getEnvironment 获取运行环境
func getEnvironment() string {
	env := os.Getenv("ENVIRONMENT")
	if env == "" {
		env = "development"
	}
	return env
}
