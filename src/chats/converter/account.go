package converter

import (
	"chats/models"
	"chats/sdk"
	uuid "github.com/satori/go.uuid"
)

func ConvertAccountFromModel(model *models.Account) *sdk.Account {
	if model == nil || model.Id == uuid.Nil {
		return &sdk.Account {}
	}

	return &sdk.Account{
		Id:         model.Id,
		Account:    model.Account,
		Type:       model.Type,
		ExternalId: model.ExternalId,
		FirstName:  model.FirstName,
		MiddleName: model.MiddleName,
		LastName:   model.LastName,
		Email:      model.Email,
		Phone:      model.Phone,
		AvatarUrl:  model.AvatarUrl,
	}
}

func ConvertAccoutnFromExpandedAccountModel(model *models.ExpandedAccountModel) *sdk.Account {
	if model == nil || model.Id == uuid.Nil {
		return &sdk.Account {}
	}

	return &sdk.Account {
		Id:         model.Account.Id,
		Account:    model.Account.Account,
		Type:       model.Account.Type,
		ExternalId: model.Account.ExternalId,
		FirstName:  model.Account.FirstName,
		MiddleName: model.Account.MiddleName,
		LastName:   model.Account.LastName,
		Email:      model.Account.Email,
		Phone:      model.Account.Phone,
		AvatarUrl:  model.Account.AvatarUrl,
	}
}
