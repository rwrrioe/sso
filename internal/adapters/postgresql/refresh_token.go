package postgresql

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	domainerrors "github.com/rwrrioe/sso/internal/domain/errors"
	"github.com/rwrrioe/sso/internal/domain/models"
)

func (s *Storage) SaveRefreshToken(
	ctx context.Context,
	token, uid string,
	appID int,
	expiresAt time.Duration,
) (string, error) {
	const op = "postgresql.SaveRefreshToken"

	var savedToken string
	err := s.db.QueryRow(ctx,
		`INSERT INTO refresh_tokens(token, user_id, email, app_id, used, expires_at)
		 SELECT $1, u.id, u.email, $2, false, $3
		 FROM users u WHERE u.id = $4
		 RETURNING token`,
		token, appID, time.Now().Add(expiresAt), uid,
	).Scan(&savedToken)
	if err != nil {
		return "", fmt.Errorf("%s: %w", op, err)
	}

	return savedToken, nil
}

func (s *Storage) RefreshToken(
	ctx context.Context,
	token string,
) (*models.RefreshToken, error) {
	const op = "postgresql.RefreshToken"

	var t models.RefreshToken
	err := s.db.QueryRow(ctx,
		`SELECT token, user_id, email, app_id, expires_at
		 FROM refresh_tokens WHERE token = $1`,
		token,
	).Scan(&t.Token, &t.UserID, &t.Email, &t.AppID, &t.ExpirestAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("%s: %w", op, domainerrors.ErrInvalidToken)
		}
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return &t, nil
}

func (s *Storage) MarkUsed(ctx context.Context, token string) error {
	const op = "postgresql.MarkUsed"

	val, err := s.db.Exec(ctx,
		`UPDATE refresh_tokens SET used = true WHERE token = $1`,
		token,
	)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	if val.RowsAffected() == 0 {
		return fmt.Errorf("%s: %w", op, domainerrors.ErrInvalidToken)
	}

	return nil
}

func (s *Storage) DeleteToken(ctx context.Context, token string) error {
	const op = "postgresql.DeleteToken"

	val, err := s.db.Exec(ctx,
		`DELETE FROM refresh_tokens WHERE token = $1`,
		token,
	)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	if val.RowsAffected() == 0 {
		return fmt.Errorf("%s: %w", op, domainerrors.ErrInvalidToken)
	}

	return nil
}
