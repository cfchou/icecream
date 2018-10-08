package model

import "time"

type ApiKey struct {
	ClientID string
	Expiry time.Time
	Key string
	Revoked bool
}
