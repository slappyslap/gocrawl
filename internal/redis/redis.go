package redis

import (
	"GoCrawl/internal/log"
	"context"
	"fmt"
	"github.com/joho/godotenv"
	"os"
	"time"

	"github.com/go-redsync/redsync/v4"
	goredisSync "github.com/go-redsync/redsync/v4/redis"
	"github.com/go-redsync/redsync/v4/redis/goredis/v9"
	"github.com/redis/go-redis/v9"
)

var Redis *redis.Client
var pool goredisSync.Pool
var rs *redsync.Redsync
var lockKey = "lock:"
var ctx = context.TODO()

func init() {
	err := godotenv.Load()
	if err != nil {
		log.Error("Error loading .env file")
	}

	log.Info("Initializing Redis: %s", os.Getenv("REDIS_HOST"))
	option := &redis.Options{
		Addr:           fmt.Sprintf("%s:%s", os.Getenv("REDIS_HOST"), os.Getenv("REDIS_PORT")),
		Password:       os.Getenv("REDIS_PASS"),
		DB:             0,
		PoolSize:       10, // Limite max de connexions simultanées
		MinIdleConns:   5,  // Nombre minimum de connexions inactives à maintenir
		MaxActiveConns: 1000,
		PoolTimeout:    4 * time.Second, // Temps d'attente max pour obtenir une connexion du pool
	}

	fmt.Println(option.Addr)

	Redis = redis.NewClient(option)
	pong, err := Redis.Ping(ctx).Result()

	if err != nil {
		panic(fmt.Errorf("failed to inititialize Redis %s", err.Error()))
	}

	pool = goredis.NewPool(Redis)
	rs = redsync.New(pool)

	log.Info("Redis Initialized %s", pong)
}

func LPush(key string, members ...interface{}) error {
	return Redis.LPush(ctx, key, members...).Err()
}

func LPop(key string) string {
	val, err := Redis.LPop(ctx, key).Result()
	if err != nil {
		return ""
	}
	return val
}

func RPush(key string, members ...interface{}) error {
	return Redis.RPush(ctx, key, members...).Err()
}

func HIncryBy(key string, field string, incr int64) error {
	return Redis.HIncrBy(ctx, key, field, incr).Err()
}

func HLen(key string) int64 {
	return Redis.HLen(ctx, key).Val()
}

func Exists(key string) bool {
	res, err := Redis.Exists(ctx, key).Result()
	if err != nil {
		log.Error("failed to run redis cmd exists on key %s, err:%s", key, err.Error())
		return false
	}
	if res == 0 {
		return false
	}
	return true
}

func GetQueueSize(key string) int64 {
	return Redis.LLen(ctx, key).Val()
}
