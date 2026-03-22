package postgresql

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/rwrrioe/sso/internal/domain/models"
	"github.com/rwrrioe/sso/internal/usecase/auth"
)

func (s *Storage) SaveUser(
	ctx context.Context,
	email string,
	passHash []byte) (uuid.UUID, error) {
	const op = "storage.postgresql.SaveUser"

	var id uuid.UUID
	err := s.db.QueryRow(ctx,
		"INSERT INTO users(email, pass_hash) VALUES ($1, $2) RETURNING id",
		email, passHash).Scan(&id)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return uuid.Nil, fmt.Errorf("%s:%w", op, auth.ErrUserExists)
		}
		return uuid.Nil, fmt.Errorf("%s:%w", op, err)
	}

	return id, nil
}

func (s *Storage) User(ctx context.Context, email string) (*models.User, error) {
	const op = "storage.postgresql.User"
	var user models.User

	err := s.db.QueryRow(ctx,
		"SELECT id, email, pass_hash FROM users WHERE email=$1",
		email).Scan(
		&user.ID,
		&user.Email,
		&user.PassHash,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("%s:%w", op, auth.ErrUserNotFound)
		}

		return nil, fmt.Errorf("%s:%w", op, err)
	}

	return &user, nil
}
