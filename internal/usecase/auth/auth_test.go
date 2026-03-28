package auth_test

import (
	"context"
	"errors"
	"log/slog"
	"os"
	"testing"
	"time"

	"github.com/google/uuid"
	"go.uber.org/mock/gomock"
	"golang.org/x/crypto/bcrypt"

	domainerrors "github.com/rwrrioe/sso/internal/domain/errors"
	"github.com/rwrrioe/sso/internal/domain/models"
	"github.com/rwrrioe/sso/internal/usecase/auth"
	"github.com/rwrrioe/sso/internal/usecase/auth/mocks"
)

func newTestAuth(
	t *testing.T,
	usrSaver auth.UserSaver,
	usrProvider auth.UserProvider,
	appProvider auth.AppProvider,
	refreshTokenProvider auth.RefreshTokenProvider,
	resetTokenProvider auth.ResetTokenProvider,
) *auth.Auth {
	t.Helper()
	log := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	cfg := &auth.Config{
		AccessTokenTTL:  15 * time.Minute,
		RefreshTokenTTL: 720 * time.Hour,
		ResetTokenTTL:   1 * time.Hour,
	}
	return auth.New(log, usrSaver, usrProvider, appProvider, refreshTokenProvider, resetTokenProvider, cfg)
}

func TestAuth_RegisterNewUser(t *testing.T) {
	ctx := context.Background()
	userID := uuid.New()

	tests := []struct {
		name    string
		email   string
		pass    string
		setup   func(saver *mocks.MockUserSaver)
		wantErr error
	}{
		{
			name:  "success",
			email: "test@example.com",
			pass:  "password123",
			setup: func(saver *mocks.MockUserSaver) {
				saver.EXPECT().
					SaveUser(ctx, "test@example.com", gomock.Any()).
					Return(userID, nil)
			},
		},
		{
			name:  "user already exists",
			email: "exists@example.com",
			pass:  "password123",
			setup: func(saver *mocks.MockUserSaver) {
				saver.EXPECT().
					SaveUser(ctx, "exists@example.com", gomock.Any()).
					Return(uuid.Nil, domainerrors.ErrUserExists)
			},
			wantErr: domainerrors.ErrUserExists,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			saver := mocks.NewMockUserSaver(ctrl)
			tt.setup(saver)

			a := newTestAuth(t, saver, nil, nil, nil, nil)
			got, err := a.RegisterNewUser(ctx, tt.email, tt.pass)

			if tt.wantErr != nil {
				if !errors.Is(err, tt.wantErr) {
					t.Errorf("expected error %v, got %v", tt.wantErr, err)
				}
				return
			}
			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}
			if got == uuid.Nil {
				t.Error("expected valid uuid, got nil")
			}
		})
	}
}

func TestAuth_Login(t *testing.T) {
	ctx := context.Background()
	userID := uuid.New()
	passHash, _ := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)

	user := &models.User{
		ID:       userID,
		Email:    "test@example.com",
		PassHash: passHash,
	}
	app := &models.App{ID: 1, Name: "test", Secret: "secret"}

	tests := []struct {
		name    string
		email   string
		pass    string
		appID   int
		setup   func(p *mocks.MockUserProvider, ap *mocks.MockAppProvider, rp *mocks.MockRefreshTokenProvider)
		wantErr error
	}{
		{
			name:  "success",
			email: "test@example.com",
			pass:  "password123",
			appID: 1,
			setup: func(p *mocks.MockUserProvider, ap *mocks.MockAppProvider, rp *mocks.MockRefreshTokenProvider) {
				p.EXPECT().User(ctx, "test@example.com").Return(user, nil)
				ap.EXPECT().App(ctx, 1).Return(app, nil)
				rp.EXPECT().SaveRefreshToken(ctx, gomock.Any(), userID.String(), 1, gomock.Any()).Return("", nil)
			},
		},
		{
			name:  "user not found",
			email: "unknown@example.com",
			pass:  "password123",
			appID: 1,
			setup: func(p *mocks.MockUserProvider, ap *mocks.MockAppProvider, rp *mocks.MockRefreshTokenProvider) {
				p.EXPECT().User(ctx, "unknown@example.com").Return(nil, domainerrors.ErrUserNotFound)
			},
			wantErr: domainerrors.ErrInvalidCredentials,
		},
		{
			name:  "wrong password",
			email: "test@example.com",
			pass:  "wrongpassword",
			appID: 1,
			setup: func(p *mocks.MockUserProvider, ap *mocks.MockAppProvider, rp *mocks.MockRefreshTokenProvider) {
				p.EXPECT().User(ctx, "test@example.com").Return(user, nil)
			},
			wantErr: domainerrors.ErrInvalidCredentials,
		},
		{
			name:  "app not found",
			email: "test@example.com",
			pass:  "password123",
			appID: 99,
			setup: func(p *mocks.MockUserProvider, ap *mocks.MockAppProvider, rp *mocks.MockRefreshTokenProvider) {
				p.EXPECT().User(ctx, "test@example.com").Return(user, nil)
				ap.EXPECT().App(ctx, 99).Return(nil, domainerrors.ErrAppNotFound)
			},
			wantErr: domainerrors.ErrAppNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			provider := mocks.NewMockUserProvider(ctrl)
			appProvider := mocks.NewMockAppProvider(ctrl)
			refreshProvider := mocks.NewMockRefreshTokenProvider(ctrl)
			tt.setup(provider, appProvider, refreshProvider)

			a := newTestAuth(t, nil, provider, appProvider, refreshProvider, nil)
			got, err := a.Login(ctx, tt.email, tt.pass, tt.appID)

			if tt.wantErr != nil {
				if !errors.Is(err, tt.wantErr) {
					t.Errorf("expected error %v, got %v", tt.wantErr, err)
				}
				return
			}
			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}
			if got.AccessToken == "" || got.RefreshToken == "" {
				t.Error("expected non-empty token pair")
			}
		})
	}
}

func TestAuth_Logout(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name    string
		token   string
		setup   func(rp *mocks.MockRefreshTokenProvider)
		wantErr error
	}{
		{
			name:  "success",
			token: "valid-token",
			setup: func(rp *mocks.MockRefreshTokenProvider) {
				rp.EXPECT().MarkUsed(ctx, "valid-token").Return(nil)
			},
		},
		{
			name:  "invalid token",
			token: "invalid-token",
			setup: func(rp *mocks.MockRefreshTokenProvider) {
				rp.EXPECT().MarkUsed(ctx, "invalid-token").Return(domainerrors.ErrInvalidToken)
			},
			wantErr: domainerrors.ErrInvalidToken,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			refreshProvider := mocks.NewMockRefreshTokenProvider(ctrl)
			tt.setup(refreshProvider)

			a := newTestAuth(t, nil, nil, nil, refreshProvider, nil)
			err := a.Logout(ctx, tt.token)

			if tt.wantErr != nil {
				if !errors.Is(err, tt.wantErr) {
					t.Errorf("expected error %v, got %v", tt.wantErr, err)
				}
				return
			}
			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

func TestAuth_IsAdmin(t *testing.T) {
	ctx := context.Background()
	userID := uuid.New()

	tests := []struct {
		name    string
		setup   func(p *mocks.MockUserProvider)
		want    bool
		wantErr error
	}{
		{
			name: "is admin",
			setup: func(p *mocks.MockUserProvider) {
				p.EXPECT().IsAdmin(ctx, userID).Return(true, nil)
			},
			want: true,
		},
		{
			name: "is not admin",
			setup: func(p *mocks.MockUserProvider) {
				p.EXPECT().IsAdmin(ctx, userID).Return(false, nil)
			},
			want: false,
		},
		{
			name: "user not found",
			setup: func(p *mocks.MockUserProvider) {
				p.EXPECT().IsAdmin(ctx, userID).Return(false, domainerrors.ErrUserNotFound)
			},
			wantErr: domainerrors.ErrUserNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			provider := mocks.NewMockUserProvider(ctrl)
			tt.setup(provider)

			a := newTestAuth(t, nil, provider, nil, nil, nil)
			got, err := a.IsAdmin(ctx, userID)

			if tt.wantErr != nil {
				if !errors.Is(err, tt.wantErr) {
					t.Errorf("expected error %v, got %v", tt.wantErr, err)
				}
				return
			}
			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}
			if got != tt.want {
				t.Errorf("expected %v, got %v", tt.want, got)
			}
		})
	}
}

func TestAuth_GenerateResetToken(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name    string
		email   string
		setup   func(rp *mocks.MockResetTokenProvider)
		wantErr error
	}{
		{
			name:  "success",
			email: "test@example.com",
			setup: func(rp *mocks.MockResetTokenProvider) {
				rp.EXPECT().
					SaveResetToken(ctx, gomock.Any(), "test@example.com", gomock.Any()).
					Return(nil)
			},
		},
		{
			name:  "storage error",
			email: "test@example.com",
			setup: func(rp *mocks.MockResetTokenProvider) {
				rp.EXPECT().
					SaveResetToken(ctx, gomock.Any(), "test@example.com", gomock.Any()).
					Return(errors.New("storage error"))
			},
			wantErr: errors.New("storage error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			resetProvider := mocks.NewMockResetTokenProvider(ctrl)
			tt.setup(resetProvider)

			a := newTestAuth(t, nil, nil, nil, nil, resetProvider)
			got, err := a.GenerateResetToken(ctx, tt.email)

			if tt.wantErr != nil {
				if err == nil {
					t.Errorf("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}
			if got == "" {
				t.Error("expected non-empty reset token")
			}
		})
	}
}

func TestAuth_CreateNewPassword(t *testing.T) {
	ctx := context.Background()

	validToken := &models.ResetToken{
		Token: "valid-token",
		Email: "test@example.com",
		Used:  false,
	}

	tests := []struct {
		name       string
		resetToken string
		password   string
		email      string
		setup      func(rp *mocks.MockResetTokenProvider, up *mocks.MockUserProvider)
		wantErr    error
	}{
		{
			name:       "success",
			resetToken: "valid-token",
			password:   "newpassword123",
			email:      "test@example.com",
			setup: func(rp *mocks.MockResetTokenProvider, up *mocks.MockUserProvider) {
				rp.EXPECT().ResetToken(ctx, "valid-token").Return(validToken, nil)
				up.EXPECT().SetNewPassword(ctx, "test@example.com", gomock.Any()).Return(nil)
			},
		},
		{
			name:       "invalid reset token",
			resetToken: "bad-token",
			password:   "newpassword123",
			email:      "test@example.com",
			setup: func(rp *mocks.MockResetTokenProvider, up *mocks.MockUserProvider) {
				rp.EXPECT().ResetToken(ctx, "bad-token").Return(nil, domainerrors.ErrInvalidToken)
			},
			wantErr: domainerrors.ErrInvalidResetToken,
		},
		{
			name:       "token already used",
			resetToken: "used-token",
			password:   "newpassword123",
			email:      "test@example.com",
			setup: func(rp *mocks.MockResetTokenProvider, up *mocks.MockUserProvider) {
				rp.EXPECT().ResetToken(ctx, "used-token").Return(&models.ResetToken{
					Token: "used-token",
					Email: "test@example.com",
					Used:  true,
				}, nil)
			},
			wantErr: domainerrors.ErrResetTokenAlreadyUsed,
		},
		{
			name:       "email mismatch",
			resetToken: "valid-token",
			password:   "newpassword123",
			email:      "other@example.com",
			setup: func(rp *mocks.MockResetTokenProvider, up *mocks.MockUserProvider) {
				rp.EXPECT().ResetToken(ctx, "valid-token").Return(validToken, nil)
			},
			wantErr: domainerrors.ErrInvalidResetToken,
		},
		{
			name:       "user not found",
			resetToken: "valid-token",
			password:   "newpassword123",
			email:      "test@example.com",
			setup: func(rp *mocks.MockResetTokenProvider, up *mocks.MockUserProvider) {
				rp.EXPECT().ResetToken(ctx, "valid-token").Return(validToken, nil)
				up.EXPECT().SetNewPassword(ctx, "test@example.com", gomock.Any()).Return(domainerrors.ErrUserNotFound)
			},
			wantErr: domainerrors.ErrUserNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			resetProvider := mocks.NewMockResetTokenProvider(ctrl)
			userProvider := mocks.NewMockUserProvider(ctrl)
			tt.setup(resetProvider, userProvider)

			a := newTestAuth(t, nil, userProvider, nil, nil, resetProvider)
			err := a.CreateNewPassword(ctx, tt.resetToken, tt.password, tt.email)

			if tt.wantErr != nil {
				if !errors.Is(err, tt.wantErr) {
					t.Errorf("expected error %v, got %v", tt.wantErr, err)
				}
				return
			}
			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}
