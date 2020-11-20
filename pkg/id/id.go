package id

import (
	"github.com/google/uuid"
)

// NewID returns new string ID in a UUID v4 format
func NewID() string {
	return uuid.New().String()
}
