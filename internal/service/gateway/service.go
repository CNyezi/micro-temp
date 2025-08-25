package gateway

import (
	"context"
	"fmt"

	gatewayv1 "micro-holtye/gen/gateway/v1"
	"micro-holtye/internal/pkg/logger"
	orderv1 "micro-holtye/gen/order/v1"
	userv1 "micro-holtye/gen/user/v1"

	"connectrpc.com/connect"
	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"
)

type Service struct {
	store  *Store
	logger logger.Logger
}

func NewService(store *Store, logger logger.Logger) *Service {
	return &Service{
		store:  store,
		logger: logger,
	}
}

func (s *Service) GetUserWithOrders(ctx context.Context, userID string) (*gatewayv1.GetUserWithOrdersResponse, error) {
	// 使用带追踪的日志记录请求开始
	s.logger.InfoContext(ctx, "GetUserWithOrders request started",
		logger.UserID(userID),
		logger.Operation("GetUserWithOrders"),
		logger.Component("gateway-service"),
	)

	// 声明需要从并发任务中获取的变量
	var user *userv1.User
	var orders []*orderv1.Order

	// 创建一个 errgroup，它会绑定到传入的 context
	g, gCtx := errgroup.WithContext(ctx)

	// 并发获取用户信息
	g.Go(func() error {
		s.logger.DebugContext(gCtx, "Fetching user information",
			logger.UserID(userID),
			logger.Component("user-service-client"),
		)
		
		var err error
		user, err = s.store.GetUser(gCtx, userID)
		if err != nil {
			s.logger.ErrorContext(gCtx, "Failed to fetch user information",
				logger.UserID(userID),
				logger.ErrorCode("USER_NOT_FOUND"),
				zap.Error(err),
			)
			// 返回错误，errgroup 会处理取消其他 goroutine
			return connect.NewError(connect.CodeNotFound, fmt.Errorf("user not found: %w", err))
		}
		
		s.logger.DebugContext(gCtx, "Successfully fetched user information",
			logger.UserID(userID),
			zap.String("username", user.Username),
		)
		return nil
	})

	// 并发获取用户订单
	g.Go(func() error {
		s.logger.DebugContext(gCtx, "Fetching user orders",
			logger.UserID(userID),
			logger.Component("order-service-client"),
		)
		
		var err error
		orders, err = s.store.ListUserOrders(gCtx, userID, 10)
		if err != nil {
			// 获取订单失败不是致命错误，记录日志但不返回错误
			// 这样即使订单服务不可用，用户仍能获取基本信息
			s.logger.WarnContext(gCtx, "Failed to fetch user orders, using graceful degradation",
				logger.UserID(userID),
				logger.ErrorCode("ORDERS_UNAVAILABLE"),
				zap.Error(err),
			)
			// 返回 nil 表示这个任务"成功"（优雅降级）
			return nil
		}
		
		s.logger.DebugContext(gCtx, "Successfully fetched user orders",
			logger.UserID(userID),
			zap.Int("order_count", len(orders)),
		)
		return nil
	})

	// 等待所有 goroutine 完成
	if err := g.Wait(); err != nil {
		// 如果获取用户信息失败，这是致命错误
		s.logger.ErrorContext(ctx, "GetUserWithOrders request failed",
			logger.UserID(userID),
			logger.ErrorCode("REQUEST_FAILED"),
			zap.Error(err),
		)
		return nil, err
	}

	// 构建响应 - 将内部服务数据转换为 Gateway API 格式
	userInfo := &gatewayv1.UserInfo{
		Id:        user.Id,
		Email:     user.Email,
		Username:  user.Username,
		FullName:  user.FullName,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
	}

	// 转换订单信息
	var orderInfos []*gatewayv1.OrderInfo
	for _, order := range orders {
		var items []*gatewayv1.OrderItem
		for _, item := range order.Items {
			items = append(items, &gatewayv1.OrderItem{
				ProductId:   item.ProductId,
				ProductName: item.ProductName,
				Quantity:    item.Quantity,
				Price:       item.Price,
			})
		}

		orderInfo := &gatewayv1.OrderInfo{
			Id:          order.Id,
			UserId:      order.UserId,
			Items:       items,
			TotalAmount: order.TotalAmount,
			Status:      order.Status.String(),
			CreatedAt:   order.CreatedAt,
			UpdatedAt:   order.UpdatedAt,
		}
		orderInfos = append(orderInfos, orderInfo)
	}

	response := &gatewayv1.GetUserWithOrdersResponse{
		User:        userInfo,
		Orders:      orderInfos,
		TotalOrders: int32(len(orderInfos)),
	}

	// 记录请求成功完成
	s.logger.InfoContext(ctx, "GetUserWithOrders request completed successfully",
		logger.UserID(userID),
		logger.Operation("GetUserWithOrders"),
		zap.String("username", user.Username),
		zap.Int("total_orders", len(orderInfos)),
		logger.StatusCode(200),
	)

	return response, nil
}
