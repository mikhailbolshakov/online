package redis

import (
	"chats/infrastructure"
	"chats/models"
	"chats/sdk"
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

func (r *Redis) GetAccountById(id uuid.UUID, account *sdk.AccountModel, sdkConn *sdk.Sdk) *sentry.SystemError {
	uid := uuid.UUID.String(id)

	//	get from redis
	key := "account:" + uid
	val, _ := r.Instance.Get(key).Bytes()
	if val == nil {
		account.Id = id
		//	get from nats
		err := sdkConn.UserById(account)
		if err != nil {
			return &sentry.SystemError{
				Error:   err.Error,
				Message: err.Message,
				Code:    err.Code,
			}
		}

		//	set to redis
		accountData, _ := json.Marshal(account)
		setErr := r.Instance.Set(key, accountData, r.Ttl).Err()
		if setErr != nil {
			redisError := &sentry.SystemError{
				Error:   setErr,
				Message: RedisSetError,
				Code:    RedisSetErrorCode,
				Data:    []byte(accountData),
			}
			infrastructure.SetError(redisError)

			return redisError
		}
	} else {
		err := json.Unmarshal(val, account)
		if err != nil {
			return &sentry.SystemError{
				Error:   err,
				Message: infrastructure.UnmarshallingError,
				Code:    infrastructure.UnmarshallingErrorCode,
				Data:    val,
			}
		}
	}

	return nil
}

func (r *Redis) DoctorSpecialization(id uint, ds *sdk.ApiDoctorSpecializationResponesData, sdkConn *sdk.Sdk) *sentry.SystemError {
	uid := strconv.FormatUint(uint64(id), 10)

	//	get from redis
	key := "doctorSpecialization:" + uid
	val, _ := r.Instance.Get(key).Bytes()
	if val == nil {
		sdkDs, err := sdkConn.DoctorSpecialization(id)
		if err != nil {
			return &sentry.SystemError{
				Error:   err.Error,
				Message: err.Message,
				Code:    err.Code,
			}
		}

		ds.Id = sdkDs.Id
		ds.Title = sdkDs.Title

		//	set to redis
		dsData, _ := json.Marshal(ds)
		setErr := r.Instance.Set(key, dsData, r.Ttl).Err()
		if setErr != nil {
			redisError := &sentry.SystemError{
				Error:   setErr,
				Message: RedisSetError,
				Code:    RedisSetErrorCode,
				Data:    []byte(dsData),
			}
			infrastructure.SetError(redisError)

			return redisError
		}
	} else {
		err := json.Unmarshal(val, ds)
		if err != nil {
			return &sentry.SystemError{
				Error:   err,
				Message: infrastructure.UnmarshallingError,
				Code:    infrastructure.UnmarshallingErrorCode,
				Data:    val,
			}
		}
	}

	return nil
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

func (r *Redis) VagueUserById(userModel *sdk.AccountModel, role string, referenceId uuid.UUID, sdkConn *sdk.Sdk) *sentry.SystemError {
	user := uuid.UUID.String(userModel.Id)
	ref := uuid.UUID.String(referenceId)
	key := "user_id:" + user + ".user_type:" + role + ".order_id:" + ref

	val, _ := r.Instance.Get(key).Bytes()
	if val == nil {
		sdkConn.VagueUserById(userModel, role, "" /*referenceId*/)
		um, _ := json.Marshal(userModel)
		setErr := r.Instance.Set(key, um, r.Ttl).Err()
		if setErr != nil {
			redisError := &sentry.SystemError{
				Error:   setErr,
				Message: RedisSetError,
				Code:    RedisSetErrorCode,
				Data:    []byte(um),
			}
			infrastructure.SetError(redisError)

			return redisError
		}
	} else {
		err := json.Unmarshal(val, userModel)
		if err != nil {
			return &sentry.SystemError{
				Error:   err,
				Message: infrastructure.UnmarshallingError,
				Code:    infrastructure.UnmarshallingErrorCode,
				Data:    val,
			}
		}
	}

	return nil
}
