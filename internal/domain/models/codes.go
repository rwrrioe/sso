package models

import "time"

type ResetCode struct {
	Code      string
	ExpiresAt time.Time
	Used      bool
}
