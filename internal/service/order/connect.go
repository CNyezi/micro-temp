package order

import (
	"context"
	"errors"
	orderv1 "micro-holtye/gen/order/v1"
	"micro-holtye/gen/order/v1/orderv1connect"

	"connectrpc.com/connect"
	"github.com/shopspring/decimal"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type ConnectHandler struct {
	orderv1connect.UnimplementedOrderServiceHandler
	service *Service
}

func NewConnectHandler(service *Service) *ConnectHandler {
	return &ConnectHandler{
		service: service,
	}
}

func (h *ConnectHandler) CreateOrder(
	ctx context.Context,
	req *connect.Request[orderv1.CreateOrderRequest],
) (*connect.Response[orderv1.CreateOrderResponse], error) {
	msg := req.Msg

	var items []OrderItemInput
	for _, item := range msg.Items {
		items = append(items, OrderItemInput{
			ProductID:   item.ProductId,
			ProductName: item.ProductName,
			Quantity:    item.Quantity,
			Price:       item.Price,
		})
	}

	order, orderItems, err := h.service.CreateOrder(ctx, msg.UserId, items)
	if err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, err)
	}

	var pbItems []*orderv1.OrderItem
	for _, item := range orderItems {
		price, _ := decimal.NewFromString(item.Price)
		priceFloat, _ := price.Float64()
		pbItems = append(pbItems, &orderv1.OrderItem{
			ProductId:   item.ProductID,
			ProductName: item.ProductName,
			Quantity:    item.Quantity,
			Price:       priceFloat,
		})
	}

	totalAmount, _ := decimal.NewFromString(order.TotalAmount)
	totalFloat, _ := totalAmount.Float64()

	return connect.NewResponse(&orderv1.CreateOrderResponse{
		Order: &orderv1.Order{
			Id:          order.ID.String(),
			UserId:      order.UserID.String(),
			Items:       pbItems,
			TotalAmount: totalFloat,
			Status:      mapStatusToProto(order.Status),
			CreatedAt:   timestamppb.New(order.CreatedAt),
			UpdatedAt:   timestamppb.New(order.UpdatedAt),
		},
	}), nil
}

func (h *ConnectHandler) GetOrder(
	ctx context.Context,
	req *connect.Request[orderv1.GetOrderRequest],
) (*connect.Response[orderv1.GetOrderResponse], error) {
	order, items, err := h.service.GetOrderWithItems(ctx, req.Msg.Id)
	if err != nil {
		if err.Error() == "order not found" {
			return nil, connect.NewError(connect.CodeNotFound, errors.New("order not found"))
		}
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	var pbItems []*orderv1.OrderItem
	for _, item := range items {
		price, _ := decimal.NewFromString(item.Price)
		priceFloat, _ := price.Float64()
		pbItems = append(pbItems, &orderv1.OrderItem{
			ProductId:   item.ProductID,
			ProductName: item.ProductName,
			Quantity:    item.Quantity,
			Price:       priceFloat,
		})
	}

	totalAmount, _ := decimal.NewFromString(order.TotalAmount)
	totalFloat, _ := totalAmount.Float64()

	return connect.NewResponse(&orderv1.GetOrderResponse{
		Order: &orderv1.Order{
			Id:          order.ID.String(),
			UserId:      order.UserID.String(),
			Items:       pbItems,
			TotalAmount: totalFloat,
			Status:      mapStatusToProto(order.Status),
			CreatedAt:   timestamppb.New(order.CreatedAt),
			UpdatedAt:   timestamppb.New(order.UpdatedAt),
		},
	}), nil
}

func (h *ConnectHandler) UpdateOrderStatus(
	ctx context.Context,
	req *connect.Request[orderv1.UpdateOrderStatusRequest],
) (*connect.Response[orderv1.UpdateOrderStatusResponse], error) {
	status := mapStatusFromProto(req.Msg.Status)
	order, err := h.service.UpdateOrderStatus(ctx, req.Msg.Id, status)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	totalAmount, _ := decimal.NewFromString(order.TotalAmount)
	totalFloat, _ := totalAmount.Float64()

	return connect.NewResponse(&orderv1.UpdateOrderStatusResponse{
		Order: &orderv1.Order{
			Id:          order.ID.String(),
			UserId:      order.UserID.String(),
			TotalAmount: totalFloat,
			Status:      mapStatusToProto(order.Status),
			CreatedAt:   timestamppb.New(order.CreatedAt),
			UpdatedAt:   timestamppb.New(order.UpdatedAt),
		},
	}), nil
}

func (h *ConnectHandler) ListOrders(
	ctx context.Context,
	req *connect.Request[orderv1.ListOrdersRequest],
) (*connect.Response[orderv1.ListOrdersResponse], error) {
	pageSize := req.Msg.PageSize
	if pageSize == 0 {
		pageSize = 10
	}

	orders, err := h.service.ListOrdersByUser(ctx, req.Msg.UserId, pageSize, 0)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	var pbOrders []*orderv1.Order
	for _, order := range orders {
		totalAmount, _ := decimal.NewFromString(order.TotalAmount)
		totalFloat, _ := totalAmount.Float64()

		pbOrders = append(pbOrders, &orderv1.Order{
			Id:          order.ID.String(),
			UserId:      order.UserID.String(),
			TotalAmount: totalFloat,
			Status:      mapStatusToProto(order.Status),
			CreatedAt:   timestamppb.New(order.CreatedAt),
			UpdatedAt:   timestamppb.New(order.UpdatedAt),
		})
	}

	return connect.NewResponse(&orderv1.ListOrdersResponse{
		Orders: pbOrders,
	}), nil
}

func (h *ConnectHandler) CancelOrder(
	ctx context.Context,
	req *connect.Request[orderv1.CancelOrderRequest],
) (*connect.Response[orderv1.CancelOrderResponse], error) {
	if err := h.service.CancelOrder(ctx, req.Msg.Id); err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	return connect.NewResponse(&orderv1.CancelOrderResponse{
		Success: true,
	}), nil
}

func mapStatusToProto(status string) orderv1.OrderStatus {
	switch status {
	case "pending":
		return orderv1.OrderStatus_ORDER_STATUS_PENDING
	case "processing":
		return orderv1.OrderStatus_ORDER_STATUS_PROCESSING
	case "shipped":
		return orderv1.OrderStatus_ORDER_STATUS_SHIPPED
	case "delivered":
		return orderv1.OrderStatus_ORDER_STATUS_DELIVERED
	case "cancelled":
		return orderv1.OrderStatus_ORDER_STATUS_CANCELLED
	default:
		return orderv1.OrderStatus_ORDER_STATUS_UNSPECIFIED
	}
}

func mapStatusFromProto(status orderv1.OrderStatus) string {
	switch status {
	case orderv1.OrderStatus_ORDER_STATUS_PENDING:
		return "pending"
	case orderv1.OrderStatus_ORDER_STATUS_PROCESSING:
		return "processing"
	case orderv1.OrderStatus_ORDER_STATUS_SHIPPED:
		return "shipped"
	case orderv1.OrderStatus_ORDER_STATUS_DELIVERED:
		return "delivered"
	case orderv1.OrderStatus_ORDER_STATUS_CANCELLED:
		return "cancelled"
	default:
		return "pending"
	}
}
