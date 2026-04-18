package code

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"log/slog"
	"math/big"
	"time"

	"github.com/google/uuid"
	domainerrors "github.com/rwrrioe/sso/internal/domain/errors"
	"github.com/rwrrioe/sso/internal/lib/logger/sl"
)

type CodeType string

const (
	TypeResetCode             CodeType = "reset code"
	TypeEmailVerificationCode CodeType = "email verification code"
	Type2FACode               CodeType = "2FA code"
)

// Options  - for every send code call
type Options struct {
	CodeType CodeType
}

// Config - for usecase
type Config struct {
	CodeTTL time.Duration
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
	codeProvider CodeProvider,
	mailProvider MailProvider,
	usrProvider UserProvider,
	config *Config,
) *Code {
	return &Code{
		log:          log,
		config:       config,
		codeProvider: codeProvider,
		mailProvider: mailProvider,
		usrProvider:  usrProvider,
	}
}

func (c *Code) SendCode(
	ctx context.Context,
	email string,
	opts *Options,
) error {
	const op = "auth.SendResetCode"
	code := generateCode()

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

	codeHash := hashCode(code)

	if err := c.codeProvider.SaveCode(ctx, uid.String(), codeHash, c.config.CodeTTL); err != nil {
		log.Error("failed to save code", sl.Err(err))
		return fmt.Errorf("%s:%w", op, err)
	}

	if err := c.mailProvider.SendCode(ctx, email, code, opts.CodeType); err != nil {
		return fmt.Errorf("%s:%w", op, err)
	}

	// TODO remove in release
	log.Info("code was successfully sent", slog.String("code:", code))
	return nil
}

func (c *Code) VerifyCode(ctx context.Context, code string) error {
	const op = "auth.VerifyCode"

	codeHash := hashCode(code)

	sentCode, err := c.codeProvider.Code(ctx, codeHash)

	if err != nil {
		if errors.Is(err, domainerrors.ErrInvalidCode) {
			return fmt.Errorf("%s:%w", op, domainerrors.ErrInvalidCode)
		}

		return fmt.Errorf("%s:%w", op, err)
	}

	if sentCode.Used {
		return fmt.Errorf("%s:%w", op, domainerrors.ErrCodeAlreadyUsed)
	}

	if err := c.codeProvider.MarkCodeUsed(
		ctx,
		codeHash,
		c.config.CodeTTL); err != nil {
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

func generateCode() string {
	n, _ := rand.Int(rand.Reader, big.NewInt(900000))
	return fmt.Sprintf("%06d", n.Int64()+100000)
}

func hashCode(code string) string {
	h := sha256.Sum256([]byte(code))
	return hex.EncodeToString(h[:])
}
