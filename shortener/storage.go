package shortener

import (
	"errors"
	"fmt"
	"net/url"
)

// URLStorage url storage interface
type URLStorage interface {
	Open(url *url.URL) error

	Close() error

	NewURL(url string, id string, ttl int) error

	DeleteURLByID(id string) error

	FindByID(id string) (string, error)
}

// ErrIDNotFound id not found
var ErrIDNotFound = errors.New("ID not found")

// InitStorage initialize storage
func InitStorage(connectURL string) (URLStorage, error) {
	u, err := url.Parse(connectURL)
	if err != nil {
		return nil, err
	}

	var r URLStorage
	switch u.Scheme {
	case "redis":
		r = &RedisStorage{}
		break
	case "mysql":
		r = &DatabaseStorage{}
		break
	case "postgresql":
		r = &DatabaseStorage{}
		break
	default:
		return nil, fmt.Errorf("not support storage type <%v>", u.Scheme)
	}
	if err := r.Open(u); err != nil {
		return nil, err
	}
	return r, nil
}
