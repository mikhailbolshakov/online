package redis

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

func Init(attempt uint) *Redis {
	client := redis.NewClient(&redis.Options{
		Addr:     os.Getenv("REDIS_HOST") + ":" + os.Getenv("REDIS_PORT"),
		Password: os.Getenv("REDIS_PASSWORD"),
		DB:       0,
	})
	_, err := client.Ping().Result()
	if err != nil {
		system.ErrHandler.SetError(&system.Error{
			Error:   err,
			Message: RedisConnectionProblem + "; attempt: " + strconv.FormatUint(uint64(attempt), 10),
			Code:    RedisConnectionProblemCode,
		})

		system.Reconnect(RedisConnectionProblem, &attempt)

		return Init(attempt)
	}

	ttl, err := strconv.ParseUint(os.Getenv("REDIS_TTL"), 10, 64)
	if err != nil {
		system.ErrHandler.SetError(&system.Error{
			Error:   err,
			Message: RedisTTLNotExist,
			Code:    RedisTTLNotExistCode,
		})
		ttl = DEFAULT_TTL
	}

	return &Redis{
		Instance: client,
		Ttl:      time.Duration(ttl) * time.Second,
	}
}


