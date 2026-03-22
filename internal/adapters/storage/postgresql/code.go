package postgresql

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/rwrrioe/sso/internal/domain/models"
	"github.com/rwrrioe/sso/internal/usecase/auth"
)

func (s *Storage) SaveCode(
	ctx context.Context,
	code, uid string,
	expiresAt time.Time,
) (string, error) {
	const op = "storage.SaveCode"

	var id uuid.UUID

	err := s.db.QueryRow(ctx,
		"INSERT INTO verification_codes(user_id, code, expires_at) VALUES ($1, $2, $3) RETURNING id",
		uid, code, expiresAt).Scan(&id)

	if err != nil {
		return "", fmt.Errorf("%s:%w", op, err)
	}

	return id.String(), nil
}

func (s *Storage) Code(
	ctx context.Context,
	code string) (*models.ResetCode, error) {
	const op = "storage.Code"
	var resetCode models.ResetCode

	err := s.db.QueryRow(ctx,
		`SELECT code, user_id, expires_at, used 
		FROM codes
		WHERE code = $1 AND user_id=$2`,
		code).Scan(
		&resetCode.Code,
		&resetCode.ExpiresAt,
		&resetCode.Used,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("%s:%w", op, auth.ErrInvalidCode)
		}

		return nil, fmt.Errorf("%s:%w", op, err)
	}

	return &resetCode, nil
}

func (s *Storage) MarkUsed(ctx context.Context, code string) error {
	const op = "storage.MarkUsed"

	res, err := s.db.Exec(ctx,
		`UPDATE verification_codes 
			SET used = TRUE
			WHERE code=$1 
			`, code)

	if res.RowsAffected() == 0 {
		return fmt.Errorf("%s:%w", op, auth.ErrInvalidCode)
	}

	if err != nil {
		return fmt.Errorf("%s:%w", op, err)
	}

	return nil
}
