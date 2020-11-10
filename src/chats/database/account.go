package database

import (
	"chats/models"
	uuid "github.com/satori/go.uuid"
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

func (s *Storage) GetAccount(accountId uuid.UUID, externalId string) (*models.Account, error) {

	//account := &models.Account{}
	//s.Redis.GetAccountById(accountId, account, nil)
	return &models.Account{}, nil
}