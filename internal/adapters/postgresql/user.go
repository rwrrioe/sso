package postgresql

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	domainerrors "github.com/rwrrioe/sso/internal/domain/errors"
	"github.com/rwrrioe/sso/internal/domain/models"
)

func (s *Storage) SaveUser(
	ctx context.Context,
	email string,
	passHash []byte) (uuid.UUID, error) {
	const op = "postgresql.SaveUser"

	var id uuid.UUID
	err := s.db.QueryRow(ctx,
		"INSERT INTO users(email, pass_hash) VALUES ($1, $2) RETURNING id",
		email, passHash).Scan(&id)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return uuid.Nil, fmt.Errorf("%s:%w", op, domainerrors.ErrUserExists)
		}
		return uuid.Nil, fmt.Errorf("%s:%w", op, err)
	}

	return id, nil
}

func (s *Storage) User(ctx context.Context, email string) (*models.User, error) {
	const op = "postgresql.User"
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
			return nil, fmt.Errorf("%s:%w", op, domainerrors.ErrUserNotFound)
		}

		return nil, fmt.Errorf("%s:%w", op, err)
	}

	return &user, nil
}

func (s *Storage) SetNewPassword(
	ctx context.Context,
	email string,
	passHash []byte) error {

	const op = "postgresql.SetNewPassword"

	val, err := s.db.Exec(ctx,
		"UPDATE TABLE users WHERE email=$1 SET pass_hash=$2",
		email, passHash)
	if err != nil {
		return fmt.Errorf("%s:%w", op, err)

	}

	if val.RowsAffected() == 0 {
		return fmt.Errorf("%s:%w", op, domainerrors.ErrUserNotFound)
	}

	return nil
}
