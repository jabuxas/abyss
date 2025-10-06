package app

import "time"

type PasteMetadata struct {
	ExpiresAt    *time.Time `json:"expires_at,omitempty"`
	PasswordHash []byte     `json:"password_hash,omitempty"`
}
