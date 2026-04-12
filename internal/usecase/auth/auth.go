package auth

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/google/uuid"
	domainerrors "github.com/rwrrioe/sso/internal/domain/errors"
	"github.com/rwrrioe/sso/internal/domain/models"
	jwtlib "github.com/rwrrioe/sso/internal/lib/jwt"
	"github.com/rwrrioe/sso/internal/lib/logger/sl"
	"golang.org/x/crypto/bcrypt"
)

type Config struct {
	AccessTokenTTL  time.Duration
	RefreshTokenTTL time.Duration
	ResetTokenTTL   time.Duration
}

type Auth struct {
	log                  *slog.Logger
	config               *Config
	usrSaver             UserSaver
	usrProvider          UserProvider
	appProvider          AppProvider
	refreshTokenProvider RefreshTokenProvider
	resetTokenProvider   ResetTokenProvider
}

func New(
	log *slog.Logger,
	userSaver UserSaver,
	userProvider UserProvider,
	appProvider AppProvider,
	refreshTokenProvider RefreshTokenProvider,
	resetTokenProvider ResetTokenProvider,
	cfg *Config,
) *Auth {
	return &Auth{
		log:                  log,
		config:               cfg,
		usrSaver:             userSaver,
		usrProvider:          userProvider,
		appProvider:          appProvider,
		refreshTokenProvider: refreshTokenProvider,
		resetTokenProvider:   resetTokenProvider,
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
) (*models.TokenPair, error) {
	const op = "Auth.Login"

	log := a.log.With(
		slog.String("op", op),
		slog.String("username", email),
	)

	log.Info("attempting to login user")

	user, err := a.usrProvider.User(ctx, email)
	if err != nil {
		if errors.Is(err, domainerrors.ErrUserNotFound) {
			a.log.Warn("user not found", sl.Err(err))
			return nil, fmt.Errorf("%s: %w", op, domainerrors.ErrInvalidCredentials)
		}

		a.log.Error("failed to get user", sl.Err(err))

		return nil, fmt.Errorf("%s: %w", op, err)
	}

	if err := bcrypt.CompareHashAndPassword(user.PassHash, []byte(password)); err != nil {
		a.log.Info("invalid credentials", sl.Err(err))

		return nil, fmt.Errorf("%s: %w", op, domainerrors.ErrInvalidCredentials)
	}

	app, err := a.appProvider.App(ctx, appID)
	if err != nil {
		a.log.Error("failed to find an appId", sl.Err(err))
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	log.Info("user logged successfully")

	token, err := jwtlib.NewToken(user.ID.String(), user.Email, *app, a.config.AccessTokenTTL)
	if err != nil {
		a.log.Error("failed to generate token", sl.Err(err))

		return nil, fmt.Errorf("%s: %w", op, err)
	}

	log.Info("access token successfully created")

	// generate acc token

	refToken := uuid.New().String()
	if _, err := a.refreshTokenProvider.SaveRefreshToken(
		ctx,
		refToken,
		user.ID.String(),
		appID,
		a.config.RefreshTokenTTL); err != nil {
		a.log.Error("failed to save refresh token")
		return nil, fmt.Errorf("%s:%w", op, err)
	}

	log.Info("token pair successfully returned")
	return &models.TokenPair{
		AccessToken:  token,
		RefreshToken: refToken,
	}, nil
}

// refresh token flow

func (a *Auth) RegenerateToken(
	ctx context.Context,
	refreshToken string,
) (*models.TokenPair, error) {
	const op = "auth.RegenerateToken"

	//validate ref token
	refToken, err := a.refreshTokenProvider.RefreshToken(ctx, refreshToken)
	if err != nil {
		if errors.Is(err, domainerrors.ErrInvalidToken) {
			return nil, fmt.Errorf("%s:%w", op, domainerrors.ErrInvalidToken)
		}

		return nil, fmt.Errorf("%s:%w", op, err)
	}

	if refToken.ExpirestAt.Before(time.Now()) {
		return nil, fmt.Errorf("%s:%w", op, domainerrors.ErrTokenExpired)
	}

	log := a.log.With(
		slog.String("op", op),
		slog.String("email", refToken.Email),
	)

	// get app for new acc token
	app, err := a.appProvider.App(ctx, refToken.AppID)
	if err != nil {
		return nil, fmt.Errorf("%s:%w", op, domainerrors.ErrAppNotFound)
	}

	// gen new tokens
	token, err := jwtlib.NewToken(refToken.UserID, refToken.Email, *app, a.config.AccessTokenTTL)
	if err != nil {
		a.log.Error("failed to generate token", sl.Err(err))

		return nil, fmt.Errorf("%s: %w", op, err)
	}

	log.Info("access token successfully created")

	if err := a.refreshTokenProvider.MarkUsed(ctx, refreshToken); err != nil {
		log.Error("failed to mark old ref token used")
		return nil, fmt.Errorf("%s:%w", op, err)
	}

	newRefToken := uuid.New().String()
	if _, err := a.refreshTokenProvider.SaveRefreshToken(
		ctx,
		newRefToken,
		refToken.UserID,
		refToken.AppID,
		a.config.RefreshTokenTTL); err != nil {
		a.log.Error("failed to save refresh token")
		return nil, fmt.Errorf("%s:%w", op, err)
	}

	log.Info("token pair successfully returned")
	return &models.TokenPair{
		AccessToken:  token,
		RefreshToken: newRefToken,
	}, nil
}

// reset password flow

func (a *Auth) GenerateResetToken(
	ctx context.Context,
	email string,
) (string, error) {
	const op = "auth.GenerateResetToken"

	log := a.log.With(
		slog.String("op", op),
		slog.String("email", email),
	)

	resToken := uuid.New().String()
	if err := a.resetTokenProvider.SaveResetToken(ctx, resToken, email, a.config.ResetTokenTTL); err != nil {
		a.log.Error("failed to save reset token")
		return "", fmt.Errorf("%s:%w", op, err)
	}

	log.Info("reset token successfully generated")
	return resToken, nil
}

func (a *Auth) CreateNewPassword(
	ctx context.Context,
	resetToken, password, email string) error {

	const op = "auth.CreateNewPassword"

	log := a.log.With(
		slog.String("op", op),
		slog.String("email", email),
	)

	// validate resetToken
	if err := a.validateResetToken(ctx, resetToken, email); err != nil {
		return fmt.Errorf("%s:%w", op, err)
	}

	// set new password
	passHash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		a.log.Error("failed to generate password hash", sl.Err(err))
		return fmt.Errorf("%s: %w", op, err)
	}

	if err := a.usrProvider.SetNewPassword(ctx, email, passHash); err != nil {
		if errors.Is(err, domainerrors.ErrUserNotFound) {
			return fmt.Errorf("%s:%w", op, domainerrors.ErrUserNotFound)
		}

		return fmt.Errorf("%s:%w", op, err)
	}

	log.Info("new password successfully created")
	return nil
}

// logout

func (a *Auth) Logout(ctx context.Context, refreshToken string) error {
	const op = "auth.Logout"

	log := slog.With(
		slog.String("refresh_token", refreshToken),
	)

	if err := a.refreshTokenProvider.MarkUsed(ctx, refreshToken); err != nil {
		if errors.Is(err, domainerrors.ErrInvalidToken) {
			log.Warn("invalid token")
			return fmt.Errorf("%s:%w", op, domainerrors.ErrInvalidToken)
		}

		log.Error("failed to mark old ref token used")
		return fmt.Errorf("%s:%w", op, err)
	}

	log.Info("logged out")
	return nil
}

// authz

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

// helpers

func (a *Auth) validateResetToken(
	ctx context.Context,
	resetToken, email string) error {
	const op = "auth.validateResetToken"

	token, err := a.resetTokenProvider.ResetToken(ctx, resetToken)
	if err != nil {
		if errors.Is(err, domainerrors.ErrInvalidToken) {
			return fmt.Errorf("%s:%w", op, domainerrors.ErrInvalidResetToken)
		}

		return fmt.Errorf("%s:%w", op, err)
	}

	if token.Used {
		return fmt.Errorf("%s:%w", op, domainerrors.ErrResetTokenAlreadyUsed)
	}

	if token.Email != email {
		return fmt.Errorf("%s:%w", op, domainerrors.ErrInvalidResetToken)
	}

	return nil
}
