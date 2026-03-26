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
		uid, codeHash string,
		expiresAt time.Time,
	) error

	Code(
		ctx context.Context,
		codeHash string,
	) (*models.ResetCode, error)

	MarkUsed(ctx context.Context, codeHash string) error
}

type MailProvider interface {
	SendCode(
		ctx context.Context,
		email, code string,
		codeType CodeType,
	) error
}

type UserProvider interface {
	User(ctx context.Context, email string) (*models.User, error)
}
