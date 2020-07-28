package shortener

import (
	"errors"
)

// URLStorage url storage interface
type URLStorage interface {
	Open(storageType string, connectURL string) error
	Close() error

	NewURL(url string, id string, ttl int) error

	DeleteURLByID(id string) error

	FindByID(id string) (string, error)
}

// ErrIDNotFound id not found
var ErrIDNotFound = errors.New("ID not found")

// InitStorage initialize storage
func InitStorage(storageType string, connectURL string) error {

	return nil
}
