package main

import (
	"context"
	"flag"
	"fmt"
	"net/http"
	"os"

	"micro-holtye/gen/order/v1/orderv1connect"
	"micro-holtye/internal/pkg/config"
	"micro-holtye/internal/pkg/database"
	"micro-holtye/internal/pkg/observability"
	"micro-holtye/internal/service/order"

	"connectrpc.com/connect"
	"go.uber.org/zap"
)

// loggingInterceptor implements connect.Interceptor
type loggingInterceptor struct {
	logger *zap.Logger
}

func (i *loggingInterceptor) WrapUnary(next connect.UnaryFunc) connect.UnaryFunc {
	return func(ctx context.Context, req connect.AnyRequest) (connect.AnyResponse, error) {
		i.logger.Info("Request",
			zap.String("procedure", req.Spec().Procedure),
			zap.String("protocol", req.Peer().Protocol),
		)
		return next(ctx, req)
	}
}

func (i *loggingInterceptor) WrapStreamingClient(next connect.StreamingClientFunc) connect.StreamingClientFunc {
	return next
}

func (i *loggingInterceptor) WrapStreamingHandler(next connect.StreamingHandlerFunc) connect.StreamingHandlerFunc {
	return next
}

func main() {
	var configPath string
	flag.StringVar(&configPath, "config", "configs/order-service.yaml", "path to config file")
	flag.Parse()

	cfg, err := config.Load(configPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to load config: %v\n", err)
		os.Exit(1)
	}

	logger, err := observability.NewLogger(cfg.Log.Level, cfg.Log.Format)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to initialize logger: %v\n", err)
		os.Exit(1)
	}
	defer logger.Sync()

	db, err := database.NewConnection(cfg.Database.DSN())
	if err != nil {
		logger.Fatal("Failed to connect to database", zap.Error(err))
	}
	defer db.Close()

	redisClient, err := database.NewRedisClient(
		cfg.Redis.Host,
		cfg.Redis.Port,
		cfg.Redis.Password,
		cfg.Redis.DB,
	)
	if err != nil {
		logger.Fatal("Failed to connect to Redis", zap.Error(err))
	}
	defer redisClient.Close()

	store := order.NewStore(db.DB)
	service := order.NewService(store)
	handler := order.NewConnectHandler(service)

	mux := http.NewServeMux()

	interceptor := &loggingInterceptor{logger: logger}
	interceptors := connect.WithInterceptors(interceptor)

	path, orderHandler := orderv1connect.NewOrderServiceHandler(handler, interceptors)
	mux.Handle(path, orderHandler)

	addr := fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port)
	logger.Info("Starting order service", zap.String("address", addr))

	if err := http.ListenAndServe(addr, mux); err != nil {
		logger.Fatal("Failed to start server", zap.Error(err))
	}
}
