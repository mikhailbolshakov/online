package server

import (
	"chats/models"
	"chats/proto"
	"chats/sdk"
	"chats/system"
	uuid "github.com/satori/go.uuid"
)

type AccountConverter struct{}

func ConvertAccountFromModel(model *models.Account) *sdk.Account {
	if model == nil || model.Id == uuid.Nil {
		return &sdk.Account{}
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

func ConvertAccountFromExpandedAccountModel(model *models.ExpandedAccountModel) *sdk.Account {
	if model == nil || model.Id == uuid.Nil {
		return &sdk.Account{}
	}

	return &sdk.Account{
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

func (r *AccountConverter) CreateRequestFromProto(request *proto.CreatAccountRequest) (*sdk.CreateAccountRequest, *system.Error) {

	result := &sdk.CreateAccountRequest{
		Account:    request.Account,
		Type:       request.Type,
		ExternalId: request.ExternalId,
		FirstName:  request.FirstName,
		MiddleName: request.MiddleName,
		LastName:   request.LastName,
		Email:      request.Email,
		Phone:      request.Phone,
		AvatarUrl:  request.AvatarUrl,
	}

	return result, nil
}

func (r *AccountConverter) CreateResponseProtoFromModel(request *sdk.CreateAccountResponse) (*proto.CreateAccountResponse, *system.Error) {

	result := &proto.CreateAccountResponse{
		Account: &proto.AccountResponse{
			Id: proto.FromUUID(request.AccountId),
		},
		Errors: proto.ErrorRs(request.Errors),
	}

	return result, nil
}

func (r *AccountConverter) UpdateRequestFromProto(request *proto.UpdateAccountRequest) (*sdk.UpdateAccountRequest, *system.Error) {

	result := &sdk.UpdateAccountRequest{
		AccountId: sdk.AccountIdRequest{
			AccountId:  request.AccountId.AccountId.ToUUID(),
			ExternalId: request.AccountId.ExternalId,
		},
		FirstName:  request.FirstName,
		MiddleName: request.MiddleName,
		LastName:   request.LastName,
		Email:      request.Email,
		Phone:      request.Phone,
		AvatarUrl:  request.AvatarUrl,
	}

	return result, nil
}

func (r *AccountConverter) UpdateResponseProtoFromModel(request *sdk.UpdateAccountResponse) (*proto.UpdateAccountResponse, *system.Error) {

	result := &proto.UpdateAccountResponse{
		Errors: proto.ErrorRs(request.Errors),
	}

	return result, nil
}

func (r *AccountConverter) GetByCriteriaRequestFromProto(request *proto.GetAccountsByCriteriaRequest) (*sdk.GetAccountsByCriteriaRequest, *system.Error) {

	result := &sdk.GetAccountsByCriteriaRequest{
		AccountId: sdk.AccountIdRequest{
			AccountId:  request.AccountId.AccountId.ToUUID(),
			ExternalId: request.AccountId.ExternalId,
		},
		Email: request.Email,
		Phone: request.Phone,
	}

	return result, nil
}

func (r *AccountConverter) GetByCriteriaResponseProtoFromModel(request *sdk.GetAccountsByCriteriaResponse) (*proto.GetAccountsByCriteriaResponse, *system.Error) {

	result := &proto.GetAccountsByCriteriaResponse{
		Accounts: []*proto.AccountItem{},
		Errors:   proto.ErrorRs(request.Errors),
	}

	for _, i := range request.Accounts {
		result.Accounts = append(result.Accounts, &proto.AccountItem{
			Id:         proto.FromUUID(i.Id),
			Account:    i.Account,
			Type:       i.Type,
			ExternalId: i.ExternalId,
			FirstName:  i.FirstName,
			MiddleName: i.MiddleName,
			LastName:   i.LastName,
			Email:      i.Email,
			Phone:      i.Phone,
			AvatarUrl:  i.AvatarUrl,
		})
	}

	return result, nil
}

func (r *AccountConverter) SetOnlineStatusRequestFromProto(request *proto.SetOnlineStatusRequest) (*sdk.SetAccountOnlineStatusRequest, *system.Error) {

	result := &sdk.SetAccountOnlineStatusRequest{
		Account: &sdk.AccountIdRequest{
			AccountId:  request.AccountId.AccountId.ToUUID(),
			ExternalId: request.AccountId.ExternalId,
		},
		Status: request.Status,
	}

	return result, nil
}

func (r *AccountConverter) SetOnlineStatusResponseProtoFromModel(request *sdk.SetAccountOnlineStatusResponse) (*proto.SetOnlineStatusResponse, *system.Error) {

	result := &proto.SetOnlineStatusResponse{
		Errors: proto.ErrorRs(request.Errors),
	}

	return result, nil
}

func (r *AccountConverter) GetOnlineStatusRequestFromProto(request *proto.GetOnlineStatusRequest) (*sdk.GetAccountOnlineStatusRequest, *system.Error) {

	result := &sdk.GetAccountOnlineStatusRequest{
		Account: &sdk.AccountIdRequest{
			AccountId:  request.AccountId.AccountId.ToUUID(),
			ExternalId: request.AccountId.ExternalId,
		},
	}

	return result, nil
}

func (r *AccountConverter) GetOnlineStatusResponseProtoFromModel(request *sdk.GetAccountOnlineStatusResponse) (*proto.GetOnlineStatusResponse, *system.Error) {

	result := &proto.GetOnlineStatusResponse{
		Status: request.Status,
		Errors: proto.ErrorRs(request.Errors),
	}

	return result, nil
}
