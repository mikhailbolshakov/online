package database

import (
	"chats/models"
	"chats/system"
	uuid "github.com/satori/go.uuid"
	"time"
)

func (s *Storage) CreateAccount(accountModel *models.Account) (uuid.UUID, *system.Error) {

	result := s.Instance.Create(accountModel)

	if result.Error != nil {
		return uuid.Nil, &system.Error{Error: result.Error}
	}

	return accountModel.Id, nil
}

func (s *Storage) UpdateAccount(accountModel *models.Account) *system.Error {

	result := s.Instance.
		Model(&models.Account{}).
		Where("id = ?::uuid", accountModel.Id).
		Updates(&models.Account{
			FirstName:  accountModel.FirstName,
			MiddleName: accountModel.MiddleName,
			LastName:   accountModel.LastName,
			Email:      accountModel.Email,
			Phone:      accountModel.Phone,
			AvatarUrl:  accountModel.AvatarUrl,
			BaseModel: models.BaseModel{
				UpdatedAt: time.Now(),
			},
		})

	if result.Error != nil {
		return &system.Error{Error: result.Error}
	}

	s.Redis.DeleteAccounts([]uuid.UUID{accountModel.Id}, []string{accountModel.ExternalId})

	return nil
}

func (s *Storage) CreateOnlineStatus(onlineStatusModel *models.OnlineStatus) (uuid.UUID, *system.Error) {

	result := s.Instance.Create(onlineStatusModel)

	if result.Error != nil {
		return uuid.Nil, &system.Error{Error: result.Error}
	}

	return onlineStatusModel.Id, nil
}

func (s *Storage) UpdateOnlineStatus(accountId uuid.UUID, status string) *system.Error {

	err := s.Redis.SetAccountOnlineStatus(accountId, status)
	if err != nil {
		return err
	}

	// update DB async
	go func() {
		err := s.Instance.
			Model(&models.OnlineStatus{}).
			Where("account_id = ?::uuid", accountId).
			Updates(&models.OnlineStatus{
				Status: status,
				BaseModel: models.BaseModel{
					UpdatedAt: time.Now(),
				},
			}).Error
		system.ErrHandler.SetError(system.E(err))
	}()

	return nil

}

func (s *Storage) GetOnlineStatus(accountId uuid.UUID) (string, *system.Error) {

	status, err := s.Redis.GetAccountOnlineStatus(accountId)
	if err != nil {
		return "", err
	}

	if status != "" {
		return status, nil
	}

	model := &models.OnlineStatus{}
	e := s.Instance.
		Where("account_id = ?::uuid", accountId).
		First(model).
		Error
	if e != nil {
		return "", system.E(e)
	}

	if model.Id != uuid.Nil {
		s.Redis.SetAccountOnlineStatus(model.AccountId, model.Status)
		return model.Status, nil
	}
	return "", nil

}

func (s *Storage) GetAccount(accountId uuid.UUID, externalId string) (*models.Account, *system.Error) {

	if accountId == uuid.Nil && externalId == "" {
		return nil, &system.Error{
			Message: "AccountId and ExternalId cannot be empty both",
			Code:    0,
		}
	}

	if accountId != uuid.Nil {

		account, err := s.Redis.GetAccountModelById(accountId)
		if err != nil {
			return nil, err
		}

		if account.Id != uuid.Nil {
			return account, nil
		}

		accountModel := &models.Account{}
		s.Instance.First(accountModel, "id = ?::uuid", accountId)

		if accountModel.Id != uuid.Nil {
			s.Redis.SetAccount(accountModel)
		}
		return accountModel, nil
	}

	if externalId != "" {

		account, err := s.Redis.GetAccountModelByExternalId(externalId)
		if err != nil {
			return nil, err
		}

		if account.Id != uuid.Nil {
			return account, nil
		}

		accountModel := &models.Account{}
		s.Instance.Where("external_id = ?", externalId).First(accountModel)
		if accountModel.Id != uuid.Nil {
			s.Redis.SetAccount(accountModel)
		}
		return accountModel, nil
	}

	return nil, nil
}

func (s *Storage) GetAccountsByCriteria(critera *models.GetAccountsCriteria) ([]models.Account, *system.Error) {

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

		var result []models.Account
		if account != nil {
			result = append(result, *account)
		}

		return result, nil

	}

	q := s.Instance.Model(&models.Account{})
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

	var result []models.Account
	for rows.Next() {
		account := &models.Account{}
		err := s.Instance.ScanRows(rows, account)
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
