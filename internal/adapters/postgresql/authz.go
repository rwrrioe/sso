package postgresql

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	domainerrors "github.com/rwrrioe/sso/internal/domain/errors"
	"github.com/rwrrioe/sso/internal/domain/models"
)

func (s *Storage) IsAdmin(ctx context.Context, userID uuid.UUID) (bool, error) {
	const op = "postgresql.IsAdmin"
	var roleID int64

	err := s.db.QueryRow(ctx, "SELECT role_id FROM roles_users WHERE user_id=$1", userID).Scan(&roleID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return false, fmt.Errorf("%s:%w", op, domainerrors.ErrUserNotFound)
		}
		return false, fmt.Errorf("%s:%w", op, err)
	}

	return roleID == models.AdminRole, nil
}
