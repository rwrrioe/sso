package redis

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	domainerrors "github.com/rwrrioe/sso/internal/domain/errors"
	"github.com/rwrrioe/sso/internal/domain/models"
)

type resetTokenDTO struct {
	Token string `json:"reset_token"`
	Email string `json:"email"`
	Used  bool   `json:"used"`
}

func (s *Storage) SaveResetToken(
	ctx context.Context,
	token, email string,
	ttl time.Duration,
) error {
	const op = "redis.SaveResetToken"

	key := resetTokenKey(token)

	dto := resetTokenDTO{
		Token: token,
		Email: email,
		Used:  false,
	}

	b, err := json.Marshal(dto)
	if err != nil {
		s.log.Error("failed to marshal code")
		return fmt.Errorf("%s:%w", op, err)
	}

	if err := s.client.Set(ctx, key, b, ttl).Err(); err != nil {
		s.log.Error("failed to save the token")
		return fmt.Errorf("%s:%w", op, err)
	}

	return nil
}

func (s *Storage) ResetToken(
	ctx context.Context,
	token string,
) (*models.ResetToken, error) {
	const op = "redis.ResetToken"

	key := resetTokenKey(token)

	val, err := s.client.Get(ctx, key).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			s.log.Error("reset token not found")
			return nil, fmt.Errorf("%s:%w", op, domainerrors.ErrInvalidResetToken)
		}

		s.log.Error("failed to get reset token")
		return nil, fmt.Errorf("%s:%w", op, err)
	}

	var dto resetTokenDTO
	if err := json.Unmarshal([]byte(val), &dto); err != nil {
		s.log.Error("failed to unmarshal reset token")
		return nil, fmt.Errorf("%s:%w", op, err)
	}

	return &models.ResetToken{
		Token: dto.Token,
		Email: dto.Email,
		Used:  dto.Used,
	}, nil

}

func (s *Storage) MarkResetTokenUsed(ctx context.Context, token string) error {
	const op = "redis.MarkResetTokenUsed"

	key := codeKey(token)

	resetToken, err := s.ResetToken(ctx, token)
	if err != nil {
		if errors.Is(err, domainerrors.ErrInvalidCode) {
			s.log.Error("reset token not found")
			return fmt.Errorf("%s:%w", op, domainerrors.ErrInvalidCode)
		}

		s.log.Error("failed to get reset token")
		return fmt.Errorf("%s:%w", op, err)
	}

	resetToken.Used = true
	b, err := json.Marshal(resetToken)
	if err != nil {
		s.log.Error("failed to marshal reset token")
		return fmt.Errorf("%s:%w", op, err)
	}

	if err = s.client.Set(ctx, key, b, resetToken.TTL).Err(); err != nil {
		s.log.Error("failed to mark reset token used")
		return fmt.Errorf("%s:%w", op, err)
	}

	return nil
}

// helpers

func resetTokenKey(token string) string {
	return fmt.Sprintf("reset_token:%s", token)
}
