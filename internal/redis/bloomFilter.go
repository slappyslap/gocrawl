package redis

import (
	"GoCrawl/internal/log"
	redisbloom "github.com/RedisBloom/redisbloom-go"
	"github.com/gomodule/redigo/redis"
	"os"
	"time"
)

var rb *redisbloom.Client

func init() {
	pool := &redis.Pool{
		Dial: func() (redis.Conn, error) {
			c, err := redis.Dial("tcp", os.Getenv("REDIS_HOST")+":"+os.Getenv("REDIS_PORT"))
			if err != nil {
				log.Error("failed to connect to redis %s", err.Error())
				return nil, err
			}
			return c, nil
		},
		MaxIdle:     10,
		MaxActive:   50,
		Wait:        true,
		IdleTimeout: 240 * time.Second,
	}

	rb = redisbloom.NewClientFromPool(pool, "bloom-client-1")
}

func BloomReserve(key string, errorRate float64, capacity uint64) error {
	return rb.Reserve(key, errorRate, capacity)
}

func BloomAdd(key string, item string) (bool, error) {
	ok, err := rb.Add(key, item)

	if err != nil {
		log.Error("failed to add item %s to bloom key %s, err: %s", item, key, err.Error())
		return false, err
	}

	return ok, nil
}
