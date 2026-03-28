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

type codeDTO struct {
	UserID   string `json:"uid"`
	CodeHash string `json:"code_hash"`
	Used     bool   `json:"used"`
}

func (s *Storage) SaveCode(
	ctx context.Context,
	uid, codeHash string,
	ttl time.Duration,
) error {
	const op = "redis.SaveCode"

	key := codeKey(codeHash)

	dto := codeDTO{
		UserID:   uid,
		CodeHash: codeHash,
		Used:     false,
	}

	b, err := json.Marshal(dto)
	if err != nil {
		s.log.Error("failed to marshal code")
		return fmt.Errorf("%s:%w", op, err)
	}

	if err := s.client.Set(ctx, key, b, ttl).Err(); err != nil {
		s.log.Error("failed to save the code", codeHash)
		return fmt.Errorf("%s:%w", op, err)
	}

	return nil
}

func (s *Storage) Code(
	ctx context.Context,
	codeHash string,
) (*models.ResetCode, error) {
	const op = "redis.Code"

	key := codeKey(codeHash)

	val, err := s.client.Get(ctx, key).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			s.log.Error("code not found")
			return nil, fmt.Errorf("%s:%w", op, domainerrors.ErrInvalidCode)
		}

		s.log.Error("failed to get code")
		return nil, fmt.Errorf("%s:%w", op, err)
	}

	var dto codeDTO
	if err := json.Unmarshal([]byte(val), &dto); err != nil {
		s.log.Error("failed to unmarshal code")
		return nil, fmt.Errorf("%s:%w", op, err)
	}

	return &models.ResetCode{
		UserID:   dto.UserID,
		CodeHash: dto.CodeHash,
		Used:     dto.Used,
	}, nil
}

func (s *Storage) MarkCodeUsed(
	ctx context.Context,
	codeHash string,
	ttl time.Duration,
) error {
	const op = "redis.MarkCodeUsed"

	key := codeKey(codeHash)

	code, err := s.Code(ctx, codeHash)
	if err != nil {
		if errors.Is(err, domainerrors.ErrInvalidCode) {
			s.log.Error("code not found")
			return fmt.Errorf("%s:%w", op, domainerrors.ErrInvalidCode)
		}

		s.log.Error("failed to get code")
		return fmt.Errorf("%s:%w", op, err)
	}

	code.Used = true
	b, err := json.Marshal(code)
	if err != nil {
		s.log.Error("failed to marshal code")
		return fmt.Errorf("%s:%w", op, err)
	}

	if err = s.client.Set(ctx, key, b, ttl).Err(); err != nil {
		s.log.Error("failed to mark code used")
		return fmt.Errorf("%s:%w", op, err)
	}

	return nil
}

// helpers

func codeKey(codeHash string) string {
	return fmt.Sprintf("code:%s", codeHash)
}
