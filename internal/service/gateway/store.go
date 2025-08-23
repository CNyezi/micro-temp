package gateway

import (
	"context"
	"net/http"

	orderv1 "micro-holtye/gen/order/v1"
	"micro-holtye/gen/order/v1/orderv1connect"
	userv1 "micro-holtye/gen/user/v1"
	"micro-holtye/gen/user/v1/userv1connect"

	"connectrpc.com/connect"
)

type Store struct {
	userClient  userv1connect.UserServiceClient
	orderClient orderv1connect.OrderServiceClient
}

func NewStore(userServiceURL, orderServiceURL string) *Store {
	return &Store{
		userClient: userv1connect.NewUserServiceClient(
			http.DefaultClient,
			userServiceURL,
		),
		orderClient: orderv1connect.NewOrderServiceClient(
			http.DefaultClient,
			orderServiceURL,
		),
	}
}

func (s *Store) GetUser(ctx context.Context, userID string) (*userv1.User, error) {
	req := connect.NewRequest(&userv1.GetUserRequest{
		Id: userID,
	})

	resp, err := s.userClient.GetUser(ctx, req)
	if err != nil {
		return nil, err
	}

	return resp.Msg.User, nil
}

func (s *Store) ListUserOrders(ctx context.Context, userID string, limit int32) ([]*orderv1.Order, error) {
	req := connect.NewRequest(&orderv1.ListOrdersRequest{
		UserId:   userID,
		PageSize: limit,
	})

	resp, err := s.orderClient.ListOrders(ctx, req)
	if err != nil {
		return nil, err
	}

	return resp.Msg.Orders, nil
}
