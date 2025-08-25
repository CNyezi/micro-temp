package user

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	userdb "micro-holtye/internal/service/user/db"
	"micro-holtye/internal/pkg/logger"

	"github.com/google/uuid"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
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

func (s *Service) CreateUser(ctx context.Context, email, username, fullName, password string) (*userdb.User, error) {
	s.logger.InfoContext(ctx, "CreateUser request started",
		zap.String("email", email),
		zap.String("username", username),
		logger.Operation("CreateUser"),
		logger.Component("user-service"),
	)

	existingUser, _ := s.store.GetUserByEmail(ctx, email)
	if existingUser != nil {
		s.logger.WarnContext(ctx, "User creation failed: email already exists",
			zap.String("email", email),
			logger.ErrorCode("EMAIL_EXISTS"),
		)
		return nil, errors.New("user with this email already exists")
	}

	existingUser, _ = s.store.GetUserByUsername(ctx, username)
	if existingUser != nil {
		s.logger.WarnContext(ctx, "User creation failed: username already exists",
			zap.String("username", username),
			logger.ErrorCode("USERNAME_EXISTS"),
		)
		return nil, errors.New("user with this username already exists")
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		s.logger.ErrorContext(ctx, "Failed to hash password",
			logger.ErrorCode("HASH_FAILURE"),
			zap.Error(err),
		)
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	user, err := s.store.CreateUser(ctx, userdb.CreateUserParams{
		Email:        email,
		Username:     username,
		FullName:     sql.NullString{String: fullName, Valid: fullName != ""},
		PasswordHash: string(hashedPassword),
	})
	if err != nil {
		s.logger.ErrorContext(ctx, "Failed to create user in database",
			zap.String("email", email),
			zap.String("username", username),
			logger.ErrorCode("DB_CREATE_FAILURE"),
			zap.Error(err),
		)
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	s.logger.InfoContext(ctx, "User created successfully",
		logger.UserID(user.ID.String()),
		zap.String("username", user.Username),
		logger.StatusCode(201),
	)

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
