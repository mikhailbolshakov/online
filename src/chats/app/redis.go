package app

import (
	"chats/system"
	"github.com/go-redis/redis"
	"os"
	"strconv"
	"time"
)

const DEFAULT_TTL = 2 * 60 * 60

type Redis struct {
	Instance *redis.Client
	Ttl      time.Duration
}

func InitRedis(attempt uint) *Redis {
	client := redis.NewClient(&redis.Options{
		Addr:     os.Getenv("REDIS_HOST") + ":" + os.Getenv("REDIS_PORT"),
		Password: os.Getenv("REDIS_PASSWORD"),
		DB:       0,
	})
	_, err := client.Ping().Result()
	if err != nil {
		Instance.ErrorHandler.SetError(&system.Error{
			Error:   err,
			Message: system.RedisConnectionProblem + "; attempt: " + strconv.FormatUint(uint64(attempt), 10),
			Code:    system.RedisConnectionProblemCode,
		})

		Instance.Inf.Reconnect(system.RedisConnectionProblem, &attempt)

		return InitRedis(attempt)
	}

	ttl, err := strconv.ParseUint(os.Getenv("REDIS_TTL"), 10, 64)
	if err != nil {
		Instance.ErrorHandler.SetError(&system.Error{
			Error:   err,
			Message: system.RedisTTLNotExist,
			Code:    system.RedisTTLNotExistCode,
		})
		ttl = DEFAULT_TTL
	}

	return &Redis{
		Instance: client,
		Ttl:      time.Duration(ttl) * time.Second,
	}
}


