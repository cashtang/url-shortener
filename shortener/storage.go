package shortener

import (
	"errors"
	"fmt"
	"log"
	"net/url"
)

// URLStorage url storage interface

// URLMeta -
type URLMeta struct {
	LongURL   string
	AppID     string
	CreatedAt string
}

// URLStorage -
type URLStorage interface {
	Open(url *url.URL) error

	Close() error

	NewURL(url string, id string, appid string, ttl int) error

	DeleteURLByID(id string) error

	FindByID(id string) (*URLMeta, error)

	RegisterAppID(appid string) (string, error)

	UnregisterAppID(appid string) error

	VerifySecret(secret string) (string, error)
}

// ErrIDNotFound id not found
var ErrIDNotFound = errors.New("ID not found")

// ErrAlreadyExist -
var ErrAlreadyExist = errors.New("ID Already exists")

// ErrAppIDNotFound -
var ErrAppIDNotFound = errors.New("Appid not found")

// ErrRegisterApp -
var ErrRegisterApp = errors.New("Register app error")

// InitStorage initialize storage
func InitStorage(connectURL string) (URLStorage, error) {
	u, err := url.Parse(connectURL)
	if err != nil {
		return nil, err
	}

	var r URLStorage
	switch u.Scheme {
	case "redis":
		r = NewRedisStorage()
		break
	case "mysql":
		r = NewDatabaseStorage()
		break
	case "postgresql":
		r = NewDatabaseStorage()
		break
	default:
		return nil, fmt.Errorf("not support storage type <%v>", u.Scheme)
	}
	if err := r.Open(u); err != nil {
		log.Println("connect to storage error,", err)
		return nil, err
	}
	log.Println("connect to storage success!!!")
	return r, nil
}
