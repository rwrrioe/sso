package code_test

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"log/slog"
	"math/big"
	"os"
	"testing"
	"time"

	"github.com/google/uuid"
	"go.uber.org/mock/gomock"

	domainerrors "github.com/rwrrioe/sso/internal/domain/errors"
	"github.com/rwrrioe/sso/internal/domain/models"
	"github.com/rwrrioe/sso/internal/usecase/code"
	"github.com/rwrrioe/sso/internal/usecase/code/mocks"
)

func newTestCode(
	t *testing.T,
	codeProvider code.CodeProvider,
	mailProvider code.MailProvider,
	usrProvider code.UserProvider,
) *code.Code {
	t.Helper()
	log := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	cfg := &code.Config{
		CodeTTL: 5 * time.Minute,
	}
	return code.New(log, codeProvider, mailProvider, usrProvider, cfg)
}

func TestCode_SendCode(t *testing.T) {
	ctx := context.Background()
	userID := uuid.New()
	user := &models.User{ID: userID, Email: "test@example.com"}
	opts := &code.Options{CodeType: code.TypeResetCode}

	tests := []struct {
		name    string
		email   string
		setup   func(cp *mocks.MockCodeProvider, mp *mocks.MockMailProvider, up *mocks.MockUserProvider)
		wantErr error
	}{
		{
			name:  "success",
			email: "test@example.com",
			setup: func(cp *mocks.MockCodeProvider, mp *mocks.MockMailProvider, up *mocks.MockUserProvider) {
				up.EXPECT().User(ctx, "test@example.com").Return(user, nil)
				cp.EXPECT().SaveCode(ctx, userID.String(), gomock.Any(), 5*time.Minute).Return(nil)
				mp.EXPECT().SendCode(ctx, "test@example.com", gomock.Any(), code.TypeResetCode).Return(nil)
			},
		},
		{
			name:  "user not found",
			email: "unknown@example.com",
			setup: func(cp *mocks.MockCodeProvider, mp *mocks.MockMailProvider, up *mocks.MockUserProvider) {
				up.EXPECT().User(ctx, "unknown@example.com").Return(nil, domainerrors.ErrUserNotFound)
			},
			wantErr: domainerrors.ErrUserNotFound,
		},
		{
			name:  "save code error",
			email: "test@example.com",
			setup: func(cp *mocks.MockCodeProvider, mp *mocks.MockMailProvider, up *mocks.MockUserProvider) {
				up.EXPECT().User(ctx, "test@example.com").Return(user, nil)
				cp.EXPECT().SaveCode(ctx, userID.String(), gomock.Any(), 5*time.Minute).Return(errors.New("storage error"))
			},
			wantErr: errors.New("storage error"),
		},
		{
			name:  "mail send error",
			email: "test@example.com",
			setup: func(cp *mocks.MockCodeProvider, mp *mocks.MockMailProvider, up *mocks.MockUserProvider) {
				up.EXPECT().User(ctx, "test@example.com").Return(user, nil)
				cp.EXPECT().SaveCode(ctx, userID.String(), gomock.Any(), 5*time.Minute).Return(nil)
				mp.EXPECT().SendCode(ctx, "test@example.com", gomock.Any(), code.TypeResetCode).Return(errors.New("mail error"))
			},
			wantErr: errors.New("mail error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			cp := mocks.NewMockCodeProvider(ctrl)
			mp := mocks.NewMockMailProvider(ctrl)
			up := mocks.NewMockUserProvider(ctrl)
			tt.setup(cp, mp, up)

			c := newTestCode(t, cp, mp, up)
			err := c.SendCode(ctx, tt.email, opts)

			if tt.wantErr != nil {
				if err == nil {
					t.Errorf("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

func TestCode_VerifyCode(t *testing.T) {
	ctx := context.Background()

	validCode := "123456"
	validCodeHash := hashCode(validCode)

	tests := []struct {
		name    string
		code    string
		setup   func(cp *mocks.MockCodeProvider)
		wantErr error
	}{
		{
			name: "success",
			code: validCode,
			setup: func(cp *mocks.MockCodeProvider) {
				cp.EXPECT().Code(ctx, validCodeHash).Return(&models.ResetCode{
					CodeHash: validCodeHash,
					Used:     false,
				}, nil)
				cp.EXPECT().MarkCodeUsed(ctx, validCodeHash, 5*time.Minute).Return(nil)
			},
		},
		{
			name: "invalid code",
			code: "000000",
			setup: func(cp *mocks.MockCodeProvider) {
				cp.EXPECT().Code(ctx, gomock.Any()).Return(nil, domainerrors.ErrInvalidCode)
			},
			wantErr: domainerrors.ErrInvalidCode,
		},
		{
			name: "code already used",
			code: validCode,
			setup: func(cp *mocks.MockCodeProvider) {
				cp.EXPECT().Code(ctx, validCodeHash).Return(&models.ResetCode{
					CodeHash: validCodeHash,
					Used:     true,
				}, nil)
			},
			wantErr: domainerrors.ErrCodeAlreadyUsed,
		},
		{
			name: "mark used error",
			code: validCode,
			setup: func(cp *mocks.MockCodeProvider) {
				cp.EXPECT().Code(ctx, validCodeHash).Return(&models.ResetCode{
					CodeHash: validCodeHash,
					Used:     false,
				}, nil)
				cp.EXPECT().MarkCodeUsed(ctx, validCodeHash, 5*time.Minute).Return(errors.New("storage error"))
			},
			wantErr: errors.New("storage error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			cp := mocks.NewMockCodeProvider(ctrl)
			tt.setup(cp)

			c := newTestCode(t, cp, nil, nil)
			err := c.VerifyCode(ctx, tt.code)

			if tt.wantErr != nil {
				if err == nil {
					t.Errorf("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

func Test_generateCode(t *testing.T) {
	for i := 0; i < 100; i++ {
		got := generateCode()
		if len(got) != 6 {
			t.Errorf("expected 6 digits, got %d: %s", len(got), got)
		}
		for _, ch := range got {
			if ch < '0' || ch > '9' {
				t.Errorf("expected only digits, got: %s", got)
			}
		}
	}
}

func Test_hashCode(t *testing.T) {
	tests := []struct {
		name string
		code string
		want string
	}{
		{
			name: "same input same hash",
			code: "123456",
			want: hashCode("123456"),
		},
		{
			name: "different input different hash",
			code: "654321",
			want: hashCode("654321"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := hashCode(tt.code)
			if got != tt.want {
				t.Errorf("hashCode() = %v, want %v", got, tt.want)
			}
			if len(got) != 64 {
				t.Errorf("expected sha256 hex length 64, got %d", len(got))
			}
		})
	}
}

// helpers

func generateCode() string {
	n, _ := rand.Int(rand.Reader, big.NewInt(900000))
	return fmt.Sprintf("%06d", n.Int64()+100000)
}

func hashCode(code string) string {
	h := sha256.Sum256([]byte(code))
	return hex.EncodeToString(h[:])
}
