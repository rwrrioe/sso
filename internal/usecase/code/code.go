package code

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/google/uuid"
	domainerrors "github.com/rwrrioe/sso/internal/domain/errors"
	"github.com/rwrrioe/sso/internal/lib/logger/sl"
)

type Config struct {
	codeTTL time.Duration
}

type Code struct {
	log          *slog.Logger
	config       *Config
	codeProvider CodeProvider
	mailProvider MailProvider
	usrProvider  UserProvider
}

func New(
	log *slog.Logger,
	config *Config,
	codeProvider CodeProvider,
	mailProvider MailProvider,
	usrProvider UserProvider,
) *Code {
	return &Code{
		log:          log,
		config:       config,
		codeProvider: codeProvider,
		mailProvider: mailProvider,
		usrProvider:  usrProvider,
	}
}

func (c *Code) SendCode(ctx context.Context, email string) error {
	const op = "auth.SendResetCode"
	var code string

	log := c.log.With(
		slog.String("op", op),
		slog.String("code", code),
	)

	uid, err := c.getUid(ctx, email)
	if err != nil {
		if errors.Is(err, domainerrors.ErrUserNotFound) {
			return fmt.Errorf("%s:%w", op, domainerrors.ErrUserNotFound)
		}

		return fmt.Errorf("%s:%w", op, err)
	}

	expiresAt := time.Now().Add(c.config.codeTTL)
	if _, err := c.codeProvider.SaveCode(ctx, code, uid.String(), expiresAt); err != nil {
		log.Error("failed to save code", sl.Err(err))
		return fmt.Errorf("%s:%w", op, err)
	}

	if err := c.mailProvider.SendCode(ctx, email, code); err != nil {
		return fmt.Errorf("%s:%w", op, err)
	}

	log.Info("code was successfully sent", slog.String("code:", code))
	return nil
}

func (c *Code) VerifyCode(ctx context.Context, code string) error {
	const op = "auth.VerifyResetCode"

	sendedCode, err := c.codeProvider.Code(ctx, code)

	if err != nil {
		if errors.Is(err, domainerrors.ErrInvalidCode) {
			return fmt.Errorf("%s:%w", op, domainerrors.ErrInvalidCode)
		}

		return fmt.Errorf("%s:%w", op, err)
	}

	if sendedCode.Used {
		return fmt.Errorf("%s:%w", op, domainerrors.ErrCodeAlreadyUsed)
	}

	if sendedCode.ExpiresAt.Before(time.Now()) {
		return fmt.Errorf("%s:%w", op, domainerrors.ErrCodeExpired)
	}

	if err := c.codeProvider.MarkUsed(ctx, code); err != nil {
		return fmt.Errorf("%s:%w", op, err)
	}

	return nil
}

func (c *Code) getUid(ctx context.Context, email string) (uuid.UUID, error) {
	user, err := c.usrProvider.User(ctx, email)
	if err != nil {
		if errors.Is(err, domainerrors.ErrUserNotFound) {
			return uuid.Nil, domainerrors.ErrUserNotFound
		}

		return uuid.Nil, err
	}

	return user.ID, nil
}
