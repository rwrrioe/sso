package models

import "time"

type TokenPair struct {
	AccessToken  string
	RefreshToken string
}

type RefreshToken struct {
	Token      string
	UserID     string
	Email      string
	AppID      int
	ExpirestAt time.Time
}

type ResetToken struct {
	Token string
	Email string
	Used  bool
	TTL   time.Duration
}
