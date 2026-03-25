package code

import (
	"context"
	"time"

	"github.com/rwrrioe/sso/internal/domain/models"
)

type CodeProvider interface {
	//TODO polish reset code provider
	SaveCode(
		ctx context.Context,
		code, uid string,
		expiresAt time.Time,
	) (string, error)

	Code(
		ctx context.Context,
		code string,
	) (*models.ResetCode, error)

	MarkUsed(ctx context.Context, code string) error
}

type MailProvider interface {
	SendCode(ctx context.Context, email, code string) error
}

type UserProvider interface {
	User(ctx context.Context, email string) (*models.User, error)
}
