package postgresql

import (
	"context"
	"errors"
	"fmt"
	"os"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/rwrrioe/sso/internal/domain/models"
	"github.com/rwrrioe/sso/internal/storage"
)

type Storage struct {
	db *pgx.Conn
}

func New(ctx context.Context) (*Storage, error) {
	const op = "storage.postgres.New"

	connStr := fmt.Sprintf("postgres://%s:%s@%s:%s/%s",
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASS"),
		os.Getenv("DB_HOST"),
		os.Getenv("DB_PORT"),
		os.Getenv("DB_NAME"))
	db, err := pgx.Connect(ctx, connStr)
	if err != nil {
		return nil, fmt.Errorf("%s:%w", op, err)
	}

	return &Storage{db: db}, nil
}

func (s *Storage) SaveUser(ctx context.Context, email string, passHash []byte) (int64, error) {
	const op = "storage.postgresql.SaveUser"

	var id int64
	err := s.db.QueryRow(ctx,
		"INSERT INTO users(email, pass_hash) VALUES ($1, $2) RETURNING id",
		email, passHash).Scan(&id)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return 0, fmt.Errorf("%s:%w", op, storage.ErrUserExists)
		}
		return 0, fmt.Errorf("%s:%w", op, err)
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
			return nil, fmt.Errorf("%s:%w", op, storage.ErrUserNotFound)
		}

		return nil, fmt.Errorf("%s:%w", op, err)
	}

	return &user, nil
}

func (s *Storage) App(ctx context.Context, id int) (*models.App, error) {
	const op = "storage.postgresql.App"

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
			return nil, fmt.Errorf("%s:%w", op, storage.ErrAppNotFound)
		}
		return nil, fmt.Errorf("%s:%w", op, err)
	}

	return &app, nil
}

func (s *Storage) IsAdmin(ctx context.Context, userID int64) (bool, error) {
	const op = "storage.postgresql.IsAdmin"
	var roleID int64

	err := s.db.QueryRow(ctx, "SELECT role_id FROM roles_users WHERE user_id=$1", userID).Scan(&roleID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return false, fmt.Errorf("%s:%w", op, storage.ErrUserNotFound)
		}
		return false, fmt.Errorf("%s:%w", op, err)
	}

	return roleID == models.AdminRole, nil
}
