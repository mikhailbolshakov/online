package redis

import (
	"chats/infrastructure"
	"chats/models"
	"chats/sentry"
	"encoding/json"
	uuid "github.com/satori/go.uuid"
)

func (r *Redis) GetAccountModelById(id uuid.UUID) (*models.Account, *sentry.SystemError) {
	uid := uuid.UUID.String(id)

	key := "account_id:" + uid
	val, _ := r.Instance.Get(key).Bytes()

	account := &models.Account{}

	if val != nil {

		err := json.Unmarshal(val, account)
		if err != nil {
			return nil, &sentry.SystemError{
				Error:   err,
				Message: infrastructure.UnmarshallingError,
				Code:    infrastructure.UnmarshallingErrorCode,
				Data:    val,
			}
		}

		return account, nil

	}

	return account, nil

}

func (r *Redis) GetAccountModelByExternalId(externalId string) (*models.Account, *sentry.SystemError) {

	key := "account_ext_id:" + externalId
	val, _ := r.Instance.Get(key).Bytes()

	account := &models.Account{}
	if val != nil {

		err := json.Unmarshal(val, account)
		if err != nil {
			return nil, &sentry.SystemError{
				Error:   err,
				Message: infrastructure.UnmarshallingError,
				Code:    infrastructure.UnmarshallingErrorCode,
				Data:    val,
			}
		}
	}

	return account, nil

}

func (r *Redis) SetAccount(account *models.Account) *sentry.SystemError {
	uid := uuid.UUID.String(account.Id)
	key := "account_id:" + uid

	marshal, _ := json.Marshal(account)

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

	if account.ExternalId != "" {
		key := "account_ext_id:" + account.ExternalId

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
	}

	return nil
}
