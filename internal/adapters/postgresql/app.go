package postgresql

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	domainerrors "github.com/rwrrioe/sso/internal/domain/errors"
	"github.com/rwrrioe/sso/internal/domain/models"
)

func (s *Storage) App(ctx context.Context, id int) (*models.App, error) {
	const op = "postgresql.App"

	var app models.App

	err := s.db.QueryRow(ctx,
		"SELECT id,name,secret FROM apps WHERE id=$1",
		id).Scan(
		&app.ID,
		&app.Name,
		&app.Secret,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("%s:%w", op, domainerrors.ErrAppNotFound)
		}
		return nil, fmt.Errorf("%s:%w", op, err)
	}

	return &app, nil
}
