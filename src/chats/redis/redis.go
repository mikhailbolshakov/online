package redis

import (
	"chats/infrastructure"
	"chats/models"
	"chats/sentry"
	"encoding/json"
	"github.com/go-redis/redis"
	uuid "github.com/satori/go.uuid"
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
		infrastructure.SetError(&sentry.SystemError{
			Error:   err,
			Message: RedisConnectionProblem + "; attempt: " + strconv.FormatUint(uint64(attempt), 10),
			Code:    RedisConnectionProblemCode,
		})

		infrastructure.Reconnect(RedisConnectionProblem, &attempt)

		return Init(attempt)
	}

	ttl, err := strconv.ParseUint(os.Getenv("REDIS_TTL"), 10, 64)
	if err != nil {
		infrastructure.SetError(&sentry.SystemError{
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

func (r *Redis) Chat(chat *models.Chat) *sentry.SystemError {
	uid := uuid.UUID.String(chat.Id)
	key := "chat:" + uid
	val, err := r.Instance.Get(key).Bytes()

	if err != nil || val == nil {
		chat.Id = uuid.Nil
		return &sentry.SystemError{
			Error:   err,
			Message: RedisSetError,
			Code:    RedisSetErrorCode,
			Data:    val,
		}
	}

	err = json.Unmarshal(val, chat)
	if err != nil {
		chat.Id = uuid.Nil
		return &sentry.SystemError{
			Error:   err,
			Message: RedisSetError,
			Code:    RedisSetErrorCode,
			Data:    val,
		}
	}

	return nil

}

func (r *Redis) SetChat(chat *models.Chat) *sentry.SystemError {
	uid := uuid.UUID.String(chat.Id)
	key := "chat:" + uid

	marshal, _ := json.Marshal(chat)

	setErr := r.Instance.Set(key, marshal, r.Ttl).Err()
	if setErr != nil {
		redisError := &sentry.SystemError{
			Error:   setErr,
			Message: RedisSetError,
			Code:    RedisSetErrorCode,
			Data:    marshal,
		}
		infrastructure.SetError(redisError)

		return redisError
	}

	return nil
}
