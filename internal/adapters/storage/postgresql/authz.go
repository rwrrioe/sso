package postgresql

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/rwrrioe/sso/internal/domain/models"
	"github.com/rwrrioe/sso/internal/usecase/auth"
)

func (s *Storage) IsAdmin(ctx context.Context, userID uuid.UUID) (bool, error) {
	const op = "storage.postgresql.IsAdmin"
	var roleID int64

	err := s.db.QueryRow(ctx, "SELECT role_id FROM roles_users WHERE user_id=$1", userID).Scan(&roleID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return false, fmt.Errorf("%s:%w", op, auth.ErrUserNotFound)
		}
		return false, fmt.Errorf("%s:%w", op, err)
	}

	return roleID == models.AdminRole, nil
}
