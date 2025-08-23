package order

import (
	"context"
	"database/sql"

	orderdb "micro-holtye/internal/service/order/db"

	"github.com/google/uuid"
)

type Store struct {
	queries *orderdb.Queries
	db      *sql.DB
}

func NewStore(database *sql.DB) *Store {
	return &Store{
		queries: orderdb.New(database),
		db:      database,
	}
}

func (s *Store) CreateOrder(ctx context.Context, params orderdb.CreateOrderParams) (*orderdb.Order, error) {
	return s.queries.CreateOrder(ctx, params)
}

func (s *Store) GetOrder(ctx context.Context, id any) (*orderdb.Order, error) {
	orderID, ok := id.(uuid.UUID)
	if !ok {
		return nil, sql.ErrNoRows
	}
	return s.queries.GetOrder(ctx, orderID)
}

func (s *Store) UpdateOrderStatus(ctx context.Context, params orderdb.UpdateOrderStatusParams) (*orderdb.Order, error) {
	return s.queries.UpdateOrderStatus(ctx, params)
}

func (s *Store) ListOrdersByUser(ctx context.Context, userID any, limit, offset int32) ([]*orderdb.Order, error) {
	userUUID, ok := userID.(uuid.UUID)
	if !ok {
		return nil, sql.ErrNoRows
	}
	return s.queries.ListOrdersByUser(ctx, orderdb.ListOrdersByUserParams{
		UserID: userUUID,
		Limit:  limit,
		Offset: offset,
	})
}

func (s *Store) CancelOrder(ctx context.Context, id any) error {
	orderID, ok := id.(uuid.UUID)
	if !ok {
		return sql.ErrNoRows
	}
	return s.queries.CancelOrder(ctx, orderID)
}

func (s *Store) CreateOrderItem(ctx context.Context, params orderdb.CreateOrderItemParams) (*orderdb.OrderItem, error) {
	return s.queries.CreateOrderItem(ctx, params)
}

func (s *Store) GetOrderItems(ctx context.Context, orderID any) ([]*orderdb.OrderItem, error) {
	orderUUID, ok := orderID.(uuid.UUID)
	if !ok {
		return nil, sql.ErrNoRows
	}
	return s.queries.GetOrderItems(ctx, orderUUID)
}

func (s *Store) WithTx(ctx context.Context, fn func(*Store) error) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	txStore := &Store{
		queries: s.queries.WithTx(tx),
		db:      s.db,
	}

	if err := fn(txStore); err != nil {
		if rbErr := tx.Rollback(); rbErr != nil {
			return rbErr
		}
		return err
	}

	return tx.Commit()
}
