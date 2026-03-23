package models

import "time"

type TokenPair struct {
	AccessToken  string
	RefreshToken string
}

type RefreshToken struct {
	ID         string
	UserID     string
	Email      string
	AppID      int
	ExpirestAt time.Time
}
