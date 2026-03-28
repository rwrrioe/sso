package models

import "time"

type ResetCode struct {
	UserID   string
	CodeHash string
	TTL      time.Duration
	Used     bool
}
