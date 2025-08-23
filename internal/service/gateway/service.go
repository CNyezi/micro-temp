package gateway

import (
	"context"
	"fmt"
	"log"

	gatewayv1 "micro-holtye/gen/gateway/v1"
	orderv1 "micro-holtye/gen/order/v1"
	userv1 "micro-holtye/gen/user/v1"

	"connectrpc.com/connect"
	"golang.org/x/sync/errgroup"
)

type Service struct {
	store *Store
}

func NewService(store *Store) *Service {
	return &Service{
		store: store,
	}
}

func (s *Service) GetUserWithOrders(ctx context.Context, userID string) (*gatewayv1.GetUserWithOrdersResponse, error) {
	// 声明需要从并发任务中获取的变量
	var user *userv1.User
	var orders []*orderv1.Order

	// 创建一个 errgroup，它会绑定到传入的 context
	g, gCtx := errgroup.WithContext(ctx)

	// 并发获取用户信息
	g.Go(func() error {
		var err error
		user, err = s.store.GetUser(gCtx, userID)
		if err != nil {
			// 返回错误，errgroup 会处理取消其他 goroutine
			return connect.NewError(connect.CodeNotFound, fmt.Errorf("user not found: %w", err))
		}
		return nil
	})

	// 并发获取用户订单
	g.Go(func() error {
		var err error
		orders, err = s.store.ListUserOrders(gCtx, userID, 10)
		if err != nil {
			// 获取订单失败不是致命错误，记录日志但不返回错误
			// 这样即使订单服务不可用，用户仍能获取基本信息
			log.Printf("WARN: failed to fetch orders for user %s: %v", userID, err)
			// 返回 nil 表示这个任务"成功"（优雅降级）
			return nil
		}
		return nil
	})

	// 等待所有 goroutine 完成
	if err := g.Wait(); err != nil {
		// 如果获取用户信息失败，这是致命错误
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

	return &gatewayv1.GetUserWithOrdersResponse{
		User:        userInfo,
		Orders:      orderInfos,
		TotalOrders: int32(len(orderInfos)),
	}, nil
}
