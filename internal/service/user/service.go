package user

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	userdb "micro-holtye/internal/service/user/db"

	"github.com/google/uuid"

	"golang.org/x/crypto/bcrypt"
)

type Service struct {
	store *Store
}

func NewService(store *Store) *Service {
	return &Service{
		store: store,
	}
}

func (s *Service) CreateUser(ctx context.Context, email, username, fullName, password string) (*userdb.User, error) {
	existingUser, _ := s.store.GetUserByEmail(ctx, email)
	if existingUser != nil {
		return nil, errors.New("user with this email already exists")
	}

	existingUser, _ = s.store.GetUserByUsername(ctx, username)
	if existingUser != nil {
		return nil, errors.New("user with this username already exists")
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	user, err := s.store.CreateUser(ctx, userdb.CreateUserParams{
		Email:        email,
		Username:     username,
		FullName:     sql.NullString{String: fullName, Valid: fullName != ""},
		PasswordHash: string(hashedPassword),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	return user, nil
}

func (s *Service) GetUser(ctx context.Context, id string) (*userdb.User, error) {
	user, err := s.store.GetUser(ctx, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.New("user not found")
		}
		return nil, fmt.Errorf("failed to get user: %w", err)
	}
	return user, nil
}

func (s *Service) UpdateUser(ctx context.Context, id string, email, username, fullName *string) (*userdb.User, error) {
	userID, err := uuid.Parse(id)
	if err != nil {
		return nil, fmt.Errorf("invalid user ID: %w", err)
	}

	params := userdb.UpdateUserParams{
		ID: userID,
	}

	if email != nil {
		params.Email = *email
	}
	if username != nil {
		params.Username = *username
	}
	if fullName != nil {
		params.FullName = sql.NullString{String: *fullName, Valid: true}
	}

	user, err := s.store.UpdateUser(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("failed to update user: %w", err)
	}

	return user, nil
}

func (s *Service) DeleteUser(ctx context.Context, id string) error {
	if err := s.store.DeleteUser(ctx, id); err != nil {
		return fmt.Errorf("failed to delete user: %w", err)
	}
	return nil
}

func (s *Service) ListUsers(ctx context.Context, pageSize int32, offset int32) ([]*userdb.User, error) {
	users, err := s.store.ListUsers(ctx, pageSize, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to list users: %w", err)
	}
	return users, nil
}
