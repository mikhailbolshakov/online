package account

import (
	"chats/app"
	"chats/system"
	uuid "github.com/satori/go.uuid"
	rep "chats/repository"
	"time"
)

type Repository struct {
	Storage *app.Storage
	Redis *app.Redis
}

func CreateRepository(storage *app.Storage) *Repository {
	return &Repository{
		Storage: storage,
		Redis:   storage.Redis,
	}
}

func (s *Repository) CreateAccount(accountModel *Account) (uuid.UUID, *system.Error) {

	result := s.Storage.Instance.Create(accountModel)

	if result.Error != nil {
		return uuid.Nil, &system.Error{Error: result.Error}
	}

	return accountModel.Id, nil
}

func (s *Repository) UpdateAccount(accountModel *Account) *system.Error {

	result := s.Storage.Instance.
		Model(&Account{}).
		Where("id = ?::uuid", accountModel.Id).
		Updates(&Account{
			FirstName:  accountModel.FirstName,
			MiddleName: accountModel.MiddleName,
			LastName:   accountModel.LastName,
			Email:      accountModel.Email,
			Phone:      accountModel.Phone,
			AvatarUrl:  accountModel.AvatarUrl,
			BaseModel: rep.BaseModel{
				UpdatedAt: time.Now(),
			},
		})

	if result.Error != nil {
		return &system.Error{Error: result.Error}
	}

	s.redisDeleteAccounts([]uuid.UUID{accountModel.Id}, []string{accountModel.ExternalId})

	return nil
}

func (s *Repository) CreateOnlineStatus(onlineStatusModel *OnlineStatus) (uuid.UUID, *system.Error) {

	result := s.Storage.Instance.Create(onlineStatusModel)

	if result.Error != nil {
		return uuid.Nil, &system.Error{Error: result.Error}
	}

	return onlineStatusModel.Id, nil
}

func (s *Repository) UpdateOnlineStatus(accountId uuid.UUID, status string) *system.Error {

	err := s.redisSetAccountOnlineStatus(accountId, status)
	if err != nil {
		return err
	}

	// update DB async
	go func() {
		err := s.Storage.Instance.
			Model(&OnlineStatus{}).
			Where("account_id = ?::uuid", accountId).
			Updates(&OnlineStatus{
				Status: status,
				BaseModel: rep.BaseModel{
					UpdatedAt: time.Now(),
				},
			}).Error
		if err != nil {
			app.E().SetError(system.E(err))
		}
	}()

	return nil

}

func (s *Repository) GetOnlineStatus(accountId uuid.UUID) (string, *system.Error) {

	status, err := s.redisGetAccountOnlineStatus(accountId)
	if err != nil {
		return "", err
	}

	if status != "" {
		return status, nil
	}

	model := &OnlineStatus{}
	e := s.Storage.Instance.
		Where("account_id = ?::uuid", accountId).
		First(model).
		Error
	if e != nil {
		return "", system.E(e)
	}

	if model.Id != uuid.Nil {
		s.redisSetAccountOnlineStatus(model.AccountId, model.Status)
		return model.Status, nil
	}
	return "", nil

}

func (s *Repository) GetAccount(accountId uuid.UUID, externalId string) (*Account, *system.Error) {

	if accountId == uuid.Nil && externalId == "" {
		return nil, &system.Error{
			Message: "AccountId and ExternalId cannot be empty both",
			Code:    0,
		}
	}

	if accountId != uuid.Nil {

		account, err := s.redisGetAccountModelById(accountId)
		if err != nil {
			return nil, err
		}

		if account.Id != uuid.Nil {
			return account, nil
		}

		accountModel := &Account{}
		s.Storage.Instance.First(accountModel, "id = ?::uuid", accountId)

		if accountModel.Id != uuid.Nil {
			s.redisSetAccount(accountModel)
		}
		return accountModel, nil
	}

	if externalId != "" {

		account, err := s.redisGetAccountModelByExternalId(externalId)
		if err != nil {
			return nil, err
		}

		if account.Id != uuid.Nil {
			return account, nil
		}

		accountModel := &Account{}
		s.Storage.Instance.Where("external_id = ?", externalId).First(accountModel)
		if accountModel.Id != uuid.Nil {
			s.redisSetAccount(accountModel)
		}
		return accountModel, nil
	}

	return nil, nil
}

func (s *Repository) GetAccountsByCriteria(critera *GetAccountsCriteria) ([]Account, *system.Error) {

	if critera.ExternalId == "" && critera.AccountId == uuid.Nil && critera.Email == "" && critera.Phone == "" {
		return nil, &system.Error{
			Message: "All parameters are empty",
			Code:    0,
		}
	}

	if critera.AccountId != uuid.Nil || critera.ExternalId != "" {
		account, err := s.GetAccount(critera.AccountId, critera.ExternalId)
		if err != nil {
			return nil, err
		}

		var result []Account
		if account != nil {
			result = append(result, *account)
		}

		return result, nil

	}

	q := s.Storage.Instance.Model(&Account{})
	if critera.Phone != "" {
		q = q.Where("phone = ?", critera.Phone)
	}

	if critera.Email != "" {
		q = q.Where("email = ?", critera.Email)
	}

	rows, err := q.Rows()
	if err != nil {
		return nil, &system.Error{
			Error:   err,
			Message: "Query error",
		}
	}

	var result []Account
	for rows.Next() {
		account := &Account{}
		err := s.Storage.Instance.ScanRows(rows, account)
		if err != nil {
			return nil, &system.Error{
				Error:   err,
				Message: "Query error (scan rows)",
			}
		}
		result = append(result, *account)
	}

	return result, nil

}
