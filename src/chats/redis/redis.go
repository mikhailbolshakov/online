package redis

import (
	"chats/models"
	"chats/service"
	"encoding/json"
	"github.com/go-redis/redis"
	"gitlab.medzdrav.ru/health-service/go-sdk"
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
		service.SetError(&models.SystemError{
			Error:   err,
			Message: RedisConnectionProblem + "; attempt: " + strconv.FormatUint(uint64(attempt), 10),
			Code:    RedisConnectionProblemCode,
		})

		service.Reconnect(RedisConnectionProblem, &attempt)

		return Init(attempt)
	}

	ttl, err := strconv.ParseUint(os.Getenv("REDIS_TTL"), 10, 64)
	if err != nil {
		service.SetError(&models.SystemError{
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

//	deprecated
func (r *Redis) GetUserById(id uint, user *sdk.UserModel, sdkConn *sdk.Sdk) *models.SystemError {
	uid := strconv.FormatUint(uint64(id), 10)

	//	get from redis
	key := "user:" + uid
	val, _ := r.Instance.Get(key).Bytes()
	if val == nil {
		user.Id = id
		//	get from nats
		err := sdkConn.UserById(user)
		if err != nil {
			return &models.SystemError{
				Error:   err.Error,
				Message: err.Message,
				Code:    err.Code,
			}
		}

		//	set to redis
		userData, _ := json.Marshal(user)
		setErr := r.Instance.Set(key, userData, r.Ttl).Err()
		if setErr != nil {
			redisError := &models.SystemError{
				Error:   setErr,
				Message: RedisSetError,
				Code:    RedisSetErrorCode,
				Data:    []byte(userData),
			}
			service.SetError(redisError)

			return redisError
		}
	} else {
		err := json.Unmarshal(val, user)
		if err != nil {
			return &models.SystemError{
				Error:   err,
				Message: service.UnmarshallingError,
				Code:    service.UnmarshallingErrorCode,
				Data:    val,
			}
		}
	}

	return nil
}

func (r *Redis) DoctorSpecialization(id uint, ds *sdk.ApiDoctorSpecializationResponesData, sdkConn *sdk.Sdk) *models.SystemError {
	uid := strconv.FormatUint(uint64(id), 10)

	//	get from redis
	key := "doctorSpecialization:" + uid
	val, _ := r.Instance.Get(key).Bytes()
	if val == nil {
		sdkDs, err := sdkConn.DoctorSpecialization(id)
		if err != nil {
			return &models.SystemError{
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
			redisError := &models.SystemError{
				Error:   setErr,
				Message: RedisSetError,
				Code:    RedisSetErrorCode,
				Data:    []byte(dsData),
			}
			service.SetError(redisError)

			return redisError
		}
	} else {
		err := json.Unmarshal(val, ds)
		if err != nil {
			return &models.SystemError{
				Error:   err,
				Message: service.UnmarshallingError,
				Code:    service.UnmarshallingErrorCode,
				Data:    val,
			}
		}
	}

	return nil
}

func (r *Redis) Chat(chat *models.Chat) {
	uid := strconv.FormatUint(uint64(chat.ID), 10)
	key := "chat:" + uid + ":order"
	val, _ := r.Instance.Get(key).Bytes()

	if val == nil {
		chat.ID = 0
	} else {
		err := json.Unmarshal(val, chat)
		if err != nil {
			chat.ID = 0
		}
	}
}

func (r *Redis) SetChat(chat *models.Chat) *models.SystemError {
	uid := strconv.FormatUint(uint64(chat.ID), 10)
	key := "chat:" + uid + ":order"

	marshal, _ := json.Marshal(chat) //	sorry

	setErr := r.Instance.Set(key, marshal, r.Ttl).Err()
	if setErr != nil {
		redisError := &models.SystemError{
			Error:   setErr,
			Message: RedisSetError,
			Code:    RedisSetErrorCode,
			Data:    marshal,
		}
		service.SetError(redisError)

		return redisError
	}

	return nil
}

func (r *Redis) VagueUserById(userModel *sdk.UserModel, userType string, orderId uint, sdkConn *sdk.Sdk) *models.SystemError {
	user := strconv.FormatUint(uint64(userModel.Id), 10)
	order := strconv.FormatUint(uint64(orderId), 10)
	key := "user_id:" + user + ".user_type:" + userType + ".order_id:" + order

	val, _ := r.Instance.Get(key).Bytes()
	if val == nil {
		sdkConn.VagueUserById(userModel, userType, orderId)
		um, _ := json.Marshal(userModel)
		setErr := r.Instance.Set(key, um, r.Ttl).Err()
		if setErr != nil {
			redisError := &models.SystemError{
				Error:   setErr,
				Message: RedisSetError,
				Code:    RedisSetErrorCode,
				Data:    []byte(um),
			}
			service.SetError(redisError)

			return redisError
		}
	} else {
		err := json.Unmarshal(val, userModel)
		if err != nil {
			return &models.SystemError{
				Error:   err,
				Message: service.UnmarshallingError,
				Code:    service.UnmarshallingErrorCode,
				Data:    val,
			}
		}
	}

	return nil
}
