package account

import (
	"chats/app"
	"chats/system"
	"encoding/json"
	"github.com/go-redis/redis"
	uuid "github.com/satori/go.uuid"
)

func (r *Repository) redisGetAccountModelById(id uuid.UUID) (*Account, *system.Error) {
	uid := uuid.UUID.String(id)

	key := "account_id:" + uid
	val, err := r.Redis.Instance.Get(key).Bytes()

	if err != nil {
		if err != redis.Nil {
			return nil, system.SysErr(nil, system.RedisGetErrorCode, nil)
		}
	}

	account := &Account{}

	if val != nil {
		err := json.Unmarshal(val, account)
		if err != nil {
			return nil, system.SysErr(nil, system.UnmarshallingErrorCode, val)
		}
		app.L().Debugf("Account found in redis: %s", key)
	}

	return account, nil

}

func (r *Repository) redisGetAccountModelByExternalId(externalId string) (*Account, *system.Error) {

	key := "account_ext_id:" + externalId
	val, err := r.Redis.Instance.Get(key).Bytes()

	if err != nil {
		if err != redis.Nil {
			return nil, system.SysErr(nil, system.RedisGetErrorCode, nil)
		}
	}

	account := &Account{}
	if val != nil {

		err := json.Unmarshal(val, account)
		if err != nil {
			return nil, system.SysErr(nil, system.UnmarshallingErrorCode, val)
		}
		app.L().Debugf("Account found in redis: %s", key)
	}

	return account, nil

}

func (r *Repository) redisGetAccountOnlineStatus(accountId uuid.UUID) (string, *system.Error) {

	if accountId == uuid.Nil {
		return "", nil
	}

	key := "online:" + accountId.String()
	val, err := r.Redis.Instance.Get(key).Result()

	if err != nil {
		if err != redis.Nil {
			return "", system.SysErr(nil, system.RedisGetErrorCode, nil)
		}
	}
	app.L().Debugf("Account online status found in redis: %s", key)
	return val, nil

}

func (r *Repository) redisSetAccountOnlineStatus(accountId uuid.UUID, status string) *system.Error {

	if accountId == uuid.Nil {
		return nil
	}

	key := "online:" + accountId.String()
	err := r.Redis.Instance.Set(key, status, r.Redis.Ttl).Err()
	if err != nil {
		return system.SysErr(nil, system.RedisSetErrorCode, nil)
	}
	app.L().Debugf("Account online status set in redis: %s", key)

	return nil

}

func (r *Repository) redisSetAccount(account *Account) *system.Error {
	uid := uuid.UUID.String(account.Id)
	key := "account_id:" + uid

	marshal, _ := json.Marshal(account)

	setErr := r.Redis.Instance.Set(key, marshal, r.Redis.Ttl).Err()
	if setErr != nil {
		return app.E().SetError(system.SysErr(nil, system.RedisSetErrorCode, nil))
	}
	app.L().Debugf("Account set in redis: %s", key)

	if account.ExternalId != "" {
		key := "account_ext_id:" + account.ExternalId

		setErr := r.Redis.Instance.Set(key, marshal, r.Redis.Ttl).Err()
		if setErr != nil {
			return app.E().SetError(system.SysErr(nil, system.RedisSetErrorCode, nil))
		}
		app.L().Debugf("Account set in redis: %s", key)
	}

	return nil
}

func (r *Repository) redisDeleteAccounts(accountIds []uuid.UUID, externalIds []string) *system.Error {
	var keys []string
	for _, accId := range accountIds {
		keys = append(keys, "account_id:"+uuid.UUID.String(accId))
		keys = append(keys, "online:"+uuid.UUID.String(accId))
	}
	for _, extId := range externalIds {
		keys = append(keys, "account_ext_id:"+extId)
	}
	r.Redis.Instance.Del(keys...)
	app.L().Debugf("Account deletes redis: %s", keys)
	return nil
}
