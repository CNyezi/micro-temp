package gateway

import (
	"context"
	"fmt"

	gatewayv1 "micro-holtye/gen/gateway/v1"
	"micro-holtye/gen/gateway/v1/gatewayv1connect"
	"micro-holtye/internal/pkg/logger"

	"connectrpc.com/connect"
	"go.uber.org/zap"
)

type ConnectHandler struct {
	service *Service
	logger  logger.Logger
}

func NewConnectHandler(service *Service, logger logger.Logger) gatewayv1connect.GatewayServiceHandler {
	return &ConnectHandler{
		service: service,
		logger:  logger,
	}
}

func (h *ConnectHandler) GetUserWithOrders(
	ctx context.Context,
	req *connect.Request[gatewayv1.GetUserWithOrdersRequest],
) (*connect.Response[gatewayv1.GetUserWithOrdersResponse], error) {
	// 添加请求 ID 到上下文（用于追踪）
	requestID := req.Header().Get("X-Request-ID")
	if requestID != "" {
		ctx = context.WithValue(ctx, "request_id", requestID)
	}

	// 参数验证
	if req.Msg.UserId == "" {
		h.logger.WarnContext(ctx, "Invalid request: missing user ID",
			logger.Component("connect-handler"),
			logger.Operation("GetUserWithOrders"),
			logger.ErrorCode("INVALID_ARGUMENT"),
		)
		return nil, connect.NewError(connect.CodeInvalidArgument, 
			fmt.Errorf("user_id is required"))
	}

	h.logger.InfoContext(ctx, "Processing GetUserWithOrders request",
		logger.UserID(req.Msg.UserId),
		logger.RequestID(requestID),
		logger.Component("connect-handler"),
	)

	resp, err := h.service.GetUserWithOrders(ctx, req.Msg.UserId)
	if err != nil {
		h.logger.ErrorContext(ctx, "GetUserWithOrders request failed in handler",
			logger.UserID(req.Msg.UserId),
			logger.RequestID(requestID),
			zap.Error(err),
		)
		return nil, err
	}

	h.logger.InfoContext(ctx, "GetUserWithOrders request completed in handler",
		logger.UserID(req.Msg.UserId),
		logger.RequestID(requestID),
		zap.Int("response_orders_count", int(resp.TotalOrders)),
	)

	return connect.NewResponse(resp), nil
}
