package app

import "time"

type PasteMetadata struct {
	ExpiresAt *time.Time `json:"expires_at,omitempty"`
}
