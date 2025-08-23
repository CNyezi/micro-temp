package user

import (
	"context"
	"errors"

	userv1 "micro-holtye/gen/user/v1"
	"micro-holtye/gen/user/v1/userv1connect"

	"connectrpc.com/connect"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type ConnectHandler struct {
	userv1connect.UnimplementedUserServiceHandler
	service *Service
}

func NewConnectHandler(service *Service) *ConnectHandler {
	return &ConnectHandler{
		service: service,
	}
}

func (h *ConnectHandler) CreateUser(
	ctx context.Context,
	req *connect.Request[userv1.CreateUserRequest],
) (*connect.Response[userv1.CreateUserResponse], error) {
	msg := req.Msg

	user, err := h.service.CreateUser(ctx, msg.Email, msg.Username, msg.FullName, msg.Password)
	if err != nil {
		return nil, connect.NewError(connect.CodeAlreadyExists, err)
	}

	return connect.NewResponse(&userv1.CreateUserResponse{
		User: &userv1.User{
			Id:        user.ID.String(),
			Email:     user.Email,
			Username:  user.Username,
			FullName:  user.FullName.String,
			CreatedAt: timestamppb.New(user.CreatedAt),
			UpdatedAt: timestamppb.New(user.UpdatedAt),
		},
	}), nil
}

func (h *ConnectHandler) GetUser(
	ctx context.Context,
	req *connect.Request[userv1.GetUserRequest],
) (*connect.Response[userv1.GetUserResponse], error) {
	user, err := h.service.GetUser(ctx, req.Msg.Id)
	if err != nil {
		if err.Error() == "user not found" {
			return nil, connect.NewError(connect.CodeNotFound, errors.New("user not found"))
		}
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	return connect.NewResponse(&userv1.GetUserResponse{
		User: &userv1.User{
			Id:        user.ID.String(),
			Email:     user.Email,
			Username:  user.Username,
			FullName:  user.FullName.String,
			CreatedAt: timestamppb.New(user.CreatedAt),
			UpdatedAt: timestamppb.New(user.UpdatedAt),
		},
	}), nil
}

func (h *ConnectHandler) UpdateUser(
	ctx context.Context,
	req *connect.Request[userv1.UpdateUserRequest],
) (*connect.Response[userv1.UpdateUserResponse], error) {
	msg := req.Msg

	var email, username, fullName *string
	if msg.Email != nil {
		email = msg.Email
	}
	if msg.Username != nil {
		username = msg.Username
	}
	if msg.FullName != nil {
		fullName = msg.FullName
	}

	user, err := h.service.UpdateUser(ctx, msg.Id, email, username, fullName)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	return connect.NewResponse(&userv1.UpdateUserResponse{
		User: &userv1.User{
			Id:        user.ID.String(),
			Email:     user.Email,
			Username:  user.Username,
			FullName:  user.FullName.String,
			CreatedAt: timestamppb.New(user.CreatedAt),
			UpdatedAt: timestamppb.New(user.UpdatedAt),
		},
	}), nil
}

func (h *ConnectHandler) DeleteUser(
	ctx context.Context,
	req *connect.Request[userv1.DeleteUserRequest],
) (*connect.Response[userv1.DeleteUserResponse], error) {
	if err := h.service.DeleteUser(ctx, req.Msg.Id); err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	return connect.NewResponse(&userv1.DeleteUserResponse{
		Success: true,
	}), nil
}

func (h *ConnectHandler) ListUsers(
	ctx context.Context,
	req *connect.Request[userv1.ListUsersRequest],
) (*connect.Response[userv1.ListUsersResponse], error) {
	pageSize := req.Msg.PageSize
	if pageSize == 0 {
		pageSize = 10
	}

	users, err := h.service.ListUsers(ctx, pageSize, 0)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	var pbUsers []*userv1.User
	for _, user := range users {
		pbUsers = append(pbUsers, &userv1.User{
			Id:        user.ID.String(),
			Email:     user.Email,
			Username:  user.Username,
			FullName:  user.FullName.String,
			CreatedAt: timestamppb.New(user.CreatedAt),
			UpdatedAt: timestamppb.New(user.UpdatedAt),
		})
	}

	return connect.NewResponse(&userv1.ListUsersResponse{
		Users: pbUsers,
	}), nil
}
