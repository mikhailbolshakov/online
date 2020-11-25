package account

import (
	"chats/app"
	"chats/system"
	"encoding/json"
	uuid "github.com/satori/go.uuid"
)

func (r *Repository) redisGetAccountModelById(id uuid.UUID) (*Account, *system.Error) {
	uid := uuid.UUID.String(id)

	key := "account_id:" + uid
	val, _ := r.Redis.Instance.Get(key).Bytes()

	account := &Account{}

	if val != nil {

		err := json.Unmarshal(val, account)
		if err != nil {
			return nil, &system.Error{
				Error:   err,
				Message: system.GetError(system.UnmarshallingErrorCode),
				Code:    system.UnmarshallingErrorCode,
				Data:    val,
			}
		}

		return account, nil

	}

	return account, nil

}

func (r *Repository) redisGetAccountModelByExternalId(externalId string) (*Account, *system.Error) {

	key := "account_ext_id:" + externalId
	val, _ := r.Redis.Instance.Get(key).Bytes()

	account := &Account{}
	if val != nil {

		err := json.Unmarshal(val, account)
		if err != nil {
			return nil, &system.Error{
				Error:   err,
				Message: system.GetError(system.UnmarshallingErrorCode),
				Code:    system.UnmarshallingErrorCode,
				Data:    val,
			}
		}
	}

	return account, nil

}

func (r *Repository) redisGetAccountOnlineStatus(accountId uuid.UUID) (string, *system.Error) {

	if accountId == uuid.Nil {
		return "", nil
	}

	key := "online:" + accountId.String()
	val, _ := r.Redis.Instance.Get(key).Result()

	return val, nil

}

func (r *Repository) redisSetAccountOnlineStatus(accountId uuid.UUID, status string) *system.Error {

	if accountId == uuid.Nil {
		return nil
	}

	key := "online:" + accountId.String()
	err := r.Redis.Instance.Set(key, status, r.Redis.Ttl).Err()
	if err != nil {
		return system.E(err)
	}

	return nil

}

func (r *Repository) redisSetAccount(account *Account) *system.Error {
	uid := uuid.UUID.String(account.Id)
	key := "account_id:" + uid

	marshal, _ := json.Marshal(account)

	setErr := r.Redis.Instance.Set(key, marshal, r.Redis.Ttl).Err()
	if setErr != nil {
		redisError := &system.Error{
			Error:   setErr,
			Message: system.GetError(system.RedisSetErrorCode),
			Code:    system.RedisSetErrorCode,
			Data:    marshal,
		}
		app.E().SetError(redisError)

		return redisError
	}

	if account.ExternalId != "" {
		key := "account_ext_id:" + account.ExternalId

		setErr := r.Redis.Instance.Set(key, marshal, r.Redis.Ttl).Err()
		if setErr != nil {
			redisError := &system.Error{
				Error:   setErr,
				Message: system.GetError(system.RedisSetErrorCode),
				Code:    system.RedisSetErrorCode,
				Data:    marshal,
			}
			app.E().SetError(redisError)

			return redisError
		}
	}

	return nil
}

func (r *Repository) redisDeleteAccounts(accountIds []uuid.UUID, externalIds []string) *system.Error {
	var keys []string
	for _, accId := range accountIds {
		keys = append(keys,  "account_id:" + uuid.UUID.String(accId))
		keys = append(keys,  "online:" + uuid.UUID.String(accId))
	}
	for _, extId := range externalIds {
		keys = append(keys,  "account_ext_id:" + extId)
	}
	r.Redis.Instance.Del(keys...)
	return nil
}

