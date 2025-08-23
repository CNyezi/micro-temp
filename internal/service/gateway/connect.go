package gateway

import (
	"context"

	gatewayv1 "micro-holtye/gen/gateway/v1"
	"micro-holtye/gen/gateway/v1/gatewayv1connect"

	"connectrpc.com/connect"
)

type ConnectHandler struct {
	service *Service
}

func NewConnectHandler(service *Service) gatewayv1connect.GatewayServiceHandler {
	return &ConnectHandler{
		service: service,
	}
}

func (h *ConnectHandler) GetUserWithOrders(
	ctx context.Context,
	req *connect.Request[gatewayv1.GetUserWithOrdersRequest],
) (*connect.Response[gatewayv1.GetUserWithOrdersResponse], error) {
	if req.Msg.UserId == "" {
		return nil, connect.NewError(connect.CodeInvalidArgument, nil)
	}

	resp, err := h.service.GetUserWithOrders(ctx, req.Msg.UserId)
	if err != nil {
		return nil, err
	}

	return connect.NewResponse(resp), nil
}
