package shortener

import (
	"context"
	"fmt"
	"log"
	"net/url"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/google/uuid"
)

const appidSecretCatalogKey = "url-shortern:appid-secret-catalog"
const appidSecretHeader = "url-shorten:appid-secret"

// RedisStorage redis storage
type RedisStorage struct {
	URLStorage
	client *redis.Client
	ctx    context.Context
}

// NewRedisStorage -
func NewRedisStorage() *RedisStorage {
	r := &RedisStorage{}
	r.ctx = context.Background()
	return r
}

// Open connect to redis
func (r *RedisStorage) Open(url *url.URL) error {
	pswd := ""
	if p, ok := url.User.Password(); ok {
		pswd = p
	}
	rdb := redis.NewClient(&redis.Options{
		Addr:     url.Host,
		Password: pswd, // no password set
		DB:       0,    // use default DB
	})
	_, err := rdb.Ping(r.ctx).Result()
	if err != nil {
		return err
	}
	r.client = rdb
	return nil
}

// Close close redis connection
func (r *RedisStorage) Close() error {
	if r.client != nil {
		return r.client.Close()
	}
	return nil
}

func (r *RedisStorage) key(id string) string {
	return fmt.Sprintf("SHORTEN:URL:%v", id)
}

// NewURL save new url
func (r *RedisStorage) NewURL(url string, id string, appid string, ttl int) error {
	key := r.key(id)
	var ok bool
	var err error
	ok, err = r.client.HSetNX(r.ctx, key, "url", url).Result()
	if err != nil {
		log.Println("save url error,", err)
		return err
	}
	if !ok {
		log.Println("save url shortid already exists")
		return ErrAlreadyExist
	}

	r.client.HSet(r.ctx, key, "appid", appid)
	r.client.HSet(r.ctx, key, "created_at", time.Now().String()).Result()
	_, err = r.client.Expire(r.ctx, key, time.Duration(ttl)*time.Hour).Result()
	if err != nil {
		log.Println("save redis url error, ", err)
		return err
	}
	return nil
}

// DeleteURLByID delete url from redis
func (r *RedisStorage) DeleteURLByID(id string) error {
	key := r.key(id)
	_, err := r.client.Del(r.ctx, key).Result()
	if err != nil {
		log.Println("del redis url error, ", err)
		return err
	}
	return nil
}

// FindByID find url from redis
func (r *RedisStorage) FindByID(id string) (string, error) {
	key := r.key(id)
	url, err := r.client.HGet(r.ctx, key, "url").Result()
	if err != nil {
		log.Println("get redis key value error ", err)
		return "", err
	}
	return url, err
}

func (r *RedisStorage) appidStoreKey(appid string) string {
	return fmt.Sprintf("%v:%v", appidSecretHeader, appid)
}

// RegisterAppID -
func (r *RedisStorage) RegisterAppID(appid string) (string, error) {
	var err error
	var secretList []string

	// check whether appid already exists!
	secretList, err = r.client.Keys(r.ctx, appidSecretHeader+":*").Result()
	if err != nil {
		return "", err
	}

	for _, key := range secretList {
		var id string
		id, err = r.client.Get(r.ctx, key).Result()
		if err == redis.Nil {
			continue
		} else if err != nil {
			log.Println("search appid error,", err)
			return "", err
		} else if err == nil {
			if id == appid {
				log.Println("appid already exists!!")
				return "", ErrRegisterApp
			}
		}
	}

	// generate new secret and save into redis
	var secret string
	for i := 0; i < 10; i++ {
		secret = uuid.New().String()
		_, err = r.client.Get(r.ctx, r.appidStoreKey(secret)).Result()
		if err == redis.Nil {
			r.client.Set(r.ctx, r.appidStoreKey(secret), appid, 0)
			return secret, nil
		} else if err != nil {
			return "", nil
		}
	}
	log.Println("register appid generate secret error")
	return "", ErrRegisterApp
}

// UnregisterAppID -
func (r *RedisStorage) UnregisterAppID(appid string) error {
	var err error
	var secretList []string

	// check whether appid already exists!
	secretList, err = r.client.Keys(r.ctx, appidSecretHeader+":*").Result()
	if err != nil {
		return err
	}

	for _, key := range secretList {
		var id string
		id, err = r.client.Get(r.ctx, key).Result()
		if err != nil {
			return err
		}
		if id == appid {
			r.client.Del(r.ctx, key)
		}
	}
	return nil
}

// VerifySecret -
func (r *RedisStorage) VerifySecret(secret string) (string, error) {
	var err error
	var appid string

	appid, err = r.client.Get(r.ctx, r.appidStoreKey(secret)).Result()
	if err == redis.Nil {
		return "", ErrAppIDNotFound
	} else if err != nil {
		return "", err
	}
	return appid, nil
}
