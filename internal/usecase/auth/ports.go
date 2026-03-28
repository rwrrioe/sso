package auth

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/rwrrioe/sso/internal/domain/models"
)

type UserSaver interface {
	SaveUser(
		ctx context.Context,
		email string,
		passHash []byte,
	) (uid uuid.UUID, err error)
}

type UserProvider interface {
	User(ctx context.Context, email string) (*models.User, error)
	IsAdmin(ctx context.Context, userID uuid.UUID) (bool, error)

	SetNewPassword(ctx context.Context, email string, passHash []byte) error
}

type AppProvider interface {
	App(ctx context.Context, appID int) (*models.App, error)
}

type RefreshTokenProvider interface {
	SaveRefreshToken(
		ctx context.Context,
		token, uid string,
		appID int,
		expiresAt time.Duration,
	) (string, error)

	RefreshToken(ctx context.Context, token string) (*models.RefreshToken, error)
	DeleteToken(ctx context.Context, token string) error
	MarkUsed(ctx context.Context, token string) error
}

type ResetTokenProvider interface {
	SaveResetToken(
		ctx context.Context,
		token, email string,
		ttl time.Duration,
	) error

	ResetToken(
		ctx context.Context,
		token string,
	) (*models.ResetToken, error)

	MarkResetTokenUsed(ctx context.Context, token string) error
}
