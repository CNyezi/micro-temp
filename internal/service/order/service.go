package order

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	orderdb "micro-holtye/internal/service/order/db"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type Service struct {
	store *Store
}

func NewService(store *Store) *Service {
	return &Service{
		store: store,
	}
}

type OrderItemInput struct {
	ProductID   string
	ProductName string
	Quantity    int32
	Price       float64
}

func (s *Service) CreateOrder(ctx context.Context, userID string, items []OrderItemInput) (*orderdb.Order, []*orderdb.OrderItem, error) {
	if len(items) == 0 {
		return nil, nil, errors.New("order must have at least one item")
	}

	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return nil, nil, fmt.Errorf("invalid user ID: %w", err)
	}

	totalAmount := decimal.Zero
	for _, item := range items {
		price := decimal.NewFromFloat(item.Price)
		quantity := decimal.NewFromInt32(item.Quantity)
		totalAmount = totalAmount.Add(price.Mul(quantity))
	}

	var order *orderdb.Order
	var orderItems []*orderdb.OrderItem

	err = s.store.WithTx(ctx, func(txStore *Store) error {
		order, err = txStore.CreateOrder(ctx, orderdb.CreateOrderParams{
			UserID:      userUUID,
			TotalAmount: totalAmount.String(),
			Status:      "pending",
		})
		if err != nil {
			return fmt.Errorf("failed to create order: %w", err)
		}

		for _, item := range items {
			orderItem, err := txStore.CreateOrderItem(ctx, orderdb.CreateOrderItemParams{
				OrderID:     order.ID,
				ProductID:   item.ProductID,
				ProductName: item.ProductName,
				Quantity:    item.Quantity,
				Price:       decimal.NewFromFloat(item.Price).String(),
			})
			if err != nil {
				return fmt.Errorf("failed to create order item: %w", err)
			}
			orderItems = append(orderItems, orderItem)
		}

		return nil
	})

	if err != nil {
		return nil, nil, err
	}

	return order, orderItems, nil
}

func (s *Service) GetOrder(ctx context.Context, id string) (*orderdb.Order, error) {
	orderUUID, err := uuid.Parse(id)
	if err != nil {
		return nil, fmt.Errorf("invalid order ID: %w", err)
	}

	order, err := s.store.GetOrder(ctx, orderUUID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.New("order not found")
		}
		return nil, fmt.Errorf("failed to get order: %w", err)
	}
	return order, nil
}

func (s *Service) GetOrderWithItems(ctx context.Context, id string) (*orderdb.Order, []*orderdb.OrderItem, error) {
	orderUUID, err := uuid.Parse(id)
	if err != nil {
		return nil, nil, fmt.Errorf("invalid order ID: %w", err)
	}

	order, err := s.store.GetOrder(ctx, orderUUID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil, errors.New("order not found")
		}
		return nil, nil, fmt.Errorf("failed to get order: %w", err)
	}

	items, err := s.store.GetOrderItems(ctx, orderUUID)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get order items: %w", err)
	}

	return order, items, nil
}

func (s *Service) UpdateOrderStatus(ctx context.Context, id string, status string) (*orderdb.Order, error) {
	orderUUID, err := uuid.Parse(id)
	if err != nil {
		return nil, fmt.Errorf("invalid order ID: %w", err)
	}

	order, err := s.store.UpdateOrderStatus(ctx, orderdb.UpdateOrderStatusParams{
		ID:     orderUUID,
		Status: status,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to update order status: %w", err)
	}

	return order, nil
}

func (s *Service) ListOrdersByUser(ctx context.Context, userID string, pageSize int32, offset int32) ([]*orderdb.Order, error) {
	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return nil, fmt.Errorf("invalid user ID: %w", err)
	}

	orders, err := s.store.ListOrdersByUser(ctx, userUUID, pageSize, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to list orders: %w", err)
	}
	return orders, nil
}

func (s *Service) CancelOrder(ctx context.Context, id string) error {
	orderUUID, err := uuid.Parse(id)
	if err != nil {
		return fmt.Errorf("invalid order ID: %w", err)
	}

	if err := s.store.CancelOrder(ctx, orderUUID); err != nil {
		return fmt.Errorf("failed to cancel order: %w", err)
	}
	return nil
}
