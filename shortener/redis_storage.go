package shortener

import (
	"context"
	"fmt"
	"log"
	"net/url"
	"time"

	"github.com/go-redis/redis/v8"
)

// RedisStorage redis storage
type RedisStorage struct {
	URLStorage
	client *redis.Client
	ctx    context.Context
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
func (r *RedisStorage) NewURL(url string, id string, ttl int) error {
	key := r.key(id)
	_, err := r.client.SetNX(r.ctx, key, url, time.Duration(ttl)*time.Hour).Result()
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
	value, err := r.client.Get(r.ctx, key).Result()
	if err != nil {
		log.Println("get redis key value error ", err)
		return "", err
	}
	return value, err
}
