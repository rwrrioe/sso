package auth

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/google/uuid"
	jwtlib "github.com/rwrrioe/sso/internal/lib/jwt"
	"github.com/rwrrioe/sso/internal/lib/logger/sl"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrUserExists         = errors.New("user already exists")
	ErrUserNotFound       = errors.New("user not found")
	ErrAppNotFound        = errors.New("app not found")
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrInvalidCode        = errors.New("invalid reset code")
	ErrCodeExpired        = errors.New("reset code has expired")
	ErrCodeAlreadyUsed    = errors.New("reset code was already used")
)

type AuthConfig struct {
	tokenTTL time.Duration
	codeTTL  time.Duration
}

type Auth struct {
	log          *slog.Logger
	config       *AuthConfig
	usrSaver     UserSaver
	usrProvider  UserProvider
	appProvider  AppProvider
	mailProvider MailProvider
	codeProvider CodeProvider
}

func New(
	log *slog.Logger,
	userSaver UserSaver,
	userProvider UserProvider,
	appProvider AppProvider,
	mailProvider MailProvider,
	CodeProvider CodeProvider,
	cfg *AuthConfig,
) *Auth {
	return &Auth{
		log:          log,
		config:       cfg,
		usrSaver:     userSaver,
		usrProvider:  userProvider,
		appProvider:  appProvider,
		mailProvider: mailProvider,
		codeProvider: CodeProvider,
	}
}

func (a *Auth) RegisterNewUser(
	ctx context.Context,
	email string,
	password string) (uuid.UUID, error) {
	const op = "Auth.RegisterNewUser"

	log := a.log.With(
		slog.String("op", op),
		slog.String("email", email),
	)

	log.Info("registering user")

	passHash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		log.Error("failed to generate password hash", sl.Err(err))
		return uuid.Nil, fmt.Errorf("%s: %w", op, err)
	}

	id, err := a.usrSaver.SaveUser(ctx, email, passHash)
	if err != nil {
		log.Error("failed to save user", sl.Err(err))
		return uuid.Nil, fmt.Errorf("%s: %w", op, err)
	}

	return id, nil
}

func (a *Auth) Login(
	ctx context.Context,
	email string,
	password string,
	appID int,
) (string, error) {
	const op = "Auth.Login"

	log := a.log.With(
		slog.String("op", op),
		slog.String("username", email),
	)

	log.Info("attempting to login user")

	user, err := a.usrProvider.User(ctx, email)
	if err != nil {
		if errors.Is(err, ErrUserNotFound) {
			a.log.Warn("user not found", sl.Err(err))
			return "", fmt.Errorf("%s: %w", op, ErrInvalidCredentials)
		}

		a.log.Error("failed to get user", sl.Err(err))

		return "", fmt.Errorf("%s: %w", op, err)
	}

	if err := bcrypt.CompareHashAndPassword(user.PassHash, []byte(password)); err != nil {
		a.log.Info("invalid credentials", sl.Err(err))

		return "", fmt.Errorf("%s: %w", op, ErrInvalidCredentials)
	}

	app, err := a.appProvider.App(ctx, appID)
	if err != nil {
		a.log.Error("failed to find an appId", sl.Err(err))
		return "", fmt.Errorf("%s: %w", op, err)
	}

	log.Info("user logged successfully")

	token, err := jwtlib.NewToken(*user, *app, a.config.tokenTTL)
	if err != nil {
		a.log.Error("failed to generate token", sl.Err(err))

		return "", fmt.Errorf("%s: %w", op, err)
	}

	log.Info("token successfully returned")
	return token, nil
}

func (a *Auth) IsAdmin(ctx context.Context, userID uuid.UUID) (bool, error) {
	const op = "Auth.IsAdmin"

	log := a.log.With(
		slog.String("op", op),
		slog.String("user_id", userID.String()),
	)

	log.Info("checking if user is admin")

	isAdmin, err := a.usrProvider.IsAdmin(ctx, userID)
	if err != nil {
		return false, fmt.Errorf("%s:%w", op, err)
	}

	log.Info("checked if user is admin", slog.Bool("is_admin", isAdmin))

	return isAdmin, nil
}

func (a *Auth) SendCode(ctx context.Context, email string) error {
	const op = "auth.SendResetCode"
	var code string

	log := a.log.With(
		slog.String("op:", op),
		slog.String("code:", code),
	)

	uid, err := a.getUid(ctx, email)
	if err != nil {
		if errors.Is(err, ErrUserNotFound) {
			return fmt.Errorf("%s:%w", op, ErrUserNotFound)
		}

		return fmt.Errorf("%s:%w", op, err)
	}

	expiresAt := time.Now().Add(a.config.codeTTL)
	if _, err := a.codeProvider.SaveCode(ctx, code, uid.String(), expiresAt); err != nil {
		log.Error("failed to save code", sl.Err(err))
		return fmt.Errorf("%s:%w", op, err)
	}

	if err := a.mailProvider.SendCode(ctx, email, code); err != nil {
		return fmt.Errorf("%s:%w", op, err)
	}

	log.Info("code was successfully sent", slog.String("code:", code))
	return nil
}

func (a *Auth) VerifyCode(ctx context.Context, code string) error {
	const op = "auth.VerifyResetCode"

	sendedCode, err := a.codeProvider.Code(ctx, code)

	if errors.Is(err, ErrInvalidCode) {
		return fmt.Errorf("%s:%w", op, ErrInvalidCode)
	}

	if sendedCode.Used {
		return fmt.Errorf("%s:%w", op, ErrCodeAlreadyUsed)
	}

	if sendedCode.ExpiresAt.Before(time.Now()) {
		return fmt.Errorf("%s:%w", op, ErrCodeExpired)
	}

	if err := a.codeProvider.MarkUsed(ctx, code); err != nil {
		return fmt.Errorf("%s:%w", op, err)
	}

	return nil
}

func (a *Auth) ResetPassword(ctx context.Context, email string) error {
	const op = "auth.ResetPassword"

	log := a.log.With(
		slog.String("op", op),
		slog.String("email", email),
	)

	if err := a.usrProvider.ResetPassword(ctx, email); err != nil {
		if errors.Is(err, ErrUserNotFound) {
			return fmt.Errorf("%s:%w", op, ErrUserNotFound)
		}

		return fmt.Errorf("%s:%w", op, err)
	}

	log.Info("user password was successfully reset")
	return nil
}

func (a *Auth) CreateNewPassword(ctx context.Context, email, password string) error {
	const op = "auth.CreateNewPassword"

	passHash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		a.log.Error("failed to generate password hash", sl.Err(err))
		return fmt.Errorf("%s: %w", op, err)
	}

	if err := a.usrProvider.SetNewPassword(ctx, email, passHash); err != nil {
		if errors.Is(err, ErrUserNotFound) {
			return fmt.Errorf("%s:%w", op, ErrUserNotFound)
		}

		return fmt.Errorf("%s:%w", op, err)
	}

	return nil
}

func (a *Auth) getUid(ctx context.Context, email string) (uuid.UUID, error) {
	user, err := a.usrProvider.User(ctx, email)
	if err != nil {
		if errors.Is(err, ErrUserNotFound) {
			return uuid.Nil, ErrUserNotFound
		}

		return uuid.Nil, err
	}

	return user.ID, nil
}
