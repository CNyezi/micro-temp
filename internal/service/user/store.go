package user

import (
	"context"
	"database/sql"

	userdb "micro-holtye/internal/service/user/db"

	"github.com/google/uuid"
)

type Store struct {
	queries userdb.Querier
	db      *sql.DB
}

func NewStore(database *sql.DB) *Store {
	return &Store{
		queries: userdb.New(database),
		db:      database,
	}
}

func (s *Store) CreateUser(ctx context.Context, params userdb.CreateUserParams) (*userdb.User, error) {
	return s.queries.CreateUser(ctx, params)
}

func (s *Store) GetUser(ctx context.Context, id any) (*userdb.User, error) {
	userID, ok := id.(uuid.UUID)
	if !ok {
		return nil, sql.ErrNoRows
	}
	return s.queries.GetUser(ctx, userID)
}

func (s *Store) GetUserByEmail(ctx context.Context, email string) (*userdb.User, error) {
	return s.queries.GetUserByEmail(ctx, email)
}

func (s *Store) GetUserByUsername(ctx context.Context, username string) (*userdb.User, error) {
	return s.queries.GetUserByUsername(ctx, username)
}

func (s *Store) UpdateUser(ctx context.Context, params userdb.UpdateUserParams) (*userdb.User, error) {
	return s.queries.UpdateUser(ctx, params)
}

func (s *Store) DeleteUser(ctx context.Context, id any) error {
	userID, ok := id.(uuid.UUID)
	if !ok {
		return sql.ErrNoRows
	}
	return s.queries.DeleteUser(ctx, userID)
}

func (s *Store) ListUsers(ctx context.Context, limit, offset int32) ([]*userdb.User, error) {
	return s.queries.ListUsers(ctx, userdb.ListUsersParams{
		Limit:  limit,
		Offset: offset,
	})
}
