package postgresql

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	domainerrors "github.com/rwrrioe/sso/internal/domain/errors"
	"github.com/rwrrioe/sso/internal/domain/models"
)

func (s *Storage) SaveCode(
	ctx context.Context,
	uid, codeHash string,
	expiresAt time.Time,
) error {
	const op = "storage.SaveCode"

	var id uuid.UUID

	err := s.db.QueryRow(ctx,
		"INSERT INTO verification_codes(user_id, code_hash, expires_at) VALUES ($1, $2, $3) RETURNING id",
		uid, codeHash, expiresAt).Scan(&id)

	if err != nil {
		return fmt.Errorf("%s:%w", op, err)
	}

	return nil
}

func (s *Storage) Code(
	ctx context.Context,
	codeHash string) (*models.ResetCode, error) {
	const op = "storage.Code"
	var resetCode models.ResetCode

	err := s.db.QueryRow(ctx,
		`SELECT code, user_id, expires_at, used 
		FROM codes
		WHERE code = $1 AND user_id=$2`,
		codeHash).Scan(
		&resetCode.CodeHash,
		&resetCode.ExpiresAt,
		&resetCode.Used,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("%s:%w", op, domainerrors.ErrInvalidCode)
		}

		return nil, fmt.Errorf("%s:%w", op, err)
	}

	return &resetCode, nil
}

func (s *Storage) MarkUsed(ctx context.Context, codeHash string) error {
	const op = "storage.MarkUsed"

	res, err := s.db.Exec(ctx,
		`UPDATE verification_codes 
			SET used = TRUE
			WHERE code_hash=$1 
			`, codeHash)

	if res.RowsAffected() == 0 {
		return fmt.Errorf("%s:%w", op, domainerrors.ErrInvalidCode)
	}

	if err != nil {
		return fmt.Errorf("%s:%w", op, err)
	}

	return nil
}
