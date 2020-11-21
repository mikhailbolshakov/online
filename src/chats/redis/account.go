package redis

import (
	"chats/system"
	"chats/models"
	"encoding/json"
	uuid "github.com/satori/go.uuid"
)

func (r *Redis) GetAccountModelById(id uuid.UUID) (*models.Account, *system.Error) {
	uid := uuid.UUID.String(id)

	key := "account_id:" + uid
	val, _ := r.Instance.Get(key).Bytes()

	account := &models.Account{}

	if val != nil {

		err := json.Unmarshal(val, account)
		if err != nil {
			return nil, &system.Error{
				Error:   err,
				Message: system.UnmarshallingError,
				Code:    system.UnmarshallingErrorCode,
				Data:    val,
			}
		}

		return account, nil

	}

	return account, nil

}

func (r *Redis) GetAccountModelByExternalId(externalId string) (*models.Account, *system.Error) {

	key := "account_ext_id:" + externalId
	val, _ := r.Instance.Get(key).Bytes()

	account := &models.Account{}
	if val != nil {

		err := json.Unmarshal(val, account)
		if err != nil {
			return nil, &system.Error{
				Error:   err,
				Message: system.UnmarshallingError,
				Code:    system.UnmarshallingErrorCode,
				Data:    val,
			}
		}
	}

	return account, nil

}

func (r *Redis) GetAccountOnlineStatus(accountId uuid.UUID) (string, *system.Error) {

	if accountId == uuid.Nil {
		return "", nil
	}

	key := "online:" + accountId.String()
	val, _ := r.Instance.Get(key).Result()

	return val, nil

}

func (r *Redis) SetAccountOnlineStatus(accountId uuid.UUID, status string) *system.Error {

	if accountId == uuid.Nil {
		return nil
	}

	key := "online:" + accountId.String()
	err := r.Instance.Set(key, status, r.Ttl).Err()
	if err != nil {
		return system.E(err)
	}

	return nil

}

func (r *Redis) SetAccount(account *models.Account) *system.Error {
	uid := uuid.UUID.String(account.Id)
	key := "account_id:" + uid

	marshal, _ := json.Marshal(account)

	setErr := r.Instance.Set(key, marshal, r.Ttl).Err()
	if setErr != nil {
		redisError := &system.Error{
			Error:   setErr,
			Message: RedisSetError,
			Code:    RedisSetErrorCode,
			Data:    marshal,
		}
		system.ErrHandler.SetError(redisError)

		return redisError
	}

	if account.ExternalId != "" {
		key := "account_ext_id:" + account.ExternalId

		setErr := r.Instance.Set(key, marshal, r.Ttl).Err()
		if setErr != nil {
			redisError := &system.Error{
				Error:   setErr,
				Message: RedisSetError,
				Code:    RedisSetErrorCode,
				Data:    marshal,
			}
			system.ErrHandler.SetError(redisError)

			return redisError
		}
	}

	return nil
}

func (r *Redis) DeleteAccounts(accountIds []uuid.UUID, externalIds []string) *system.Error {
	var keys []string
	for _, accId := range accountIds {
		keys = append(keys,  "account_id:" + uuid.UUID.String(accId))
		keys = append(keys,  "online:" + uuid.UUID.String(accId))
	}
	for _, extId := range externalIds {
		keys = append(keys,  "account_ext_id:" + extId)
	}
	r.Instance.Del(keys...)
	return nil
}

