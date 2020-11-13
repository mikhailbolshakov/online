package database

import (
	"chats/models"
	"chats/sentry"
	uuid "github.com/satori/go.uuid"
	"time"
)

func (s *Storage) CreateAccount(accountModel *models.Account) (uuid.UUID, error) {

	result := s.Instance.Create(accountModel)

	if result.Error != nil {
		return uuid.Nil, result.Error
	}

	return accountModel.Id, nil
}

func (s *Storage) CreateOnlineStatus(onlineStatusModel *models.OnlineStatus) (uuid.UUID, error) {

	result := s.Instance.Create(onlineStatusModel)

	if result.Error != nil {
		return uuid.Nil, result.Error
	}

	return onlineStatusModel.Id, nil
}

func (s *Storage) UpdateOnlineStatus(onlineStatusModel *models.OnlineStatus) *sentry.SystemError {

	err := s.Instance.
		Model(onlineStatusModel).
		Updates(&models.OnlineStatus{
			Status: onlineStatusModel.Status,
			BaseModel: models.BaseModel{
				UpdatedAt: time.Now(),
			},
		}).Error

	if err != nil {
		return &sentry.SystemError{
			Error: err,
		}
	}

	return nil

}

func (s *Storage) GetOnlineStatus(accountId uuid.UUID) (*models.OnlineStatus, *sentry.SystemError) {

	items := []models.OnlineStatus{}
	err := s.Instance.
		Where("account_id = ?", accountId).
		Find(&items).
		Error

	if err != nil {
		return nil, &sentry.SystemError{
			Error:   err,
			Message: DbAccountOnlineStatusGetError,
			Code:    DbAccountOnlineStatusGetErrorCode,
		}
	}

	if len(items) > 1 {
		return nil, &sentry.SystemError{
			Error:   err,
			Message: DbAccountOnlineStatusGetMoreThenOneError,
			Code:    DbAccountOnlineStatusGetMoreThenOneErrorCode,
		}
	}

	if len(items) == 0 {
		return nil, &sentry.SystemError{
			Error:   err,
			Message: DbAccountOnlineStatusGetNotFoundError,
			Code:    DbAccountOnlineStatusGetNotFoundErrorCode,
		}
	}

	return &items[0], nil

}

func (s *Storage) GetAccount(accountId uuid.UUID, externalId string) (*models.Account, *sentry.SystemError) {

	if accountId != uuid.Nil {
		account, err := s.Redis.GetAccountModelById(accountId)
		if err != nil {
			return nil, err
		}
		if account.Id != uuid.Nil {
			return account, nil
		}
		accountModel := &models.Account{}
		s.Instance.First(accountModel, "id = ?", accountId)
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
		return accountModel, nil
	}
	return nil, nil
}
