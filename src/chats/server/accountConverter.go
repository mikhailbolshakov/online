package server

import (
	"chats/proto"
	"chats/system"
	uuid "github.com/satori/go.uuid"
	a "chats/repository/account"
)

type AccountConverter struct{}

func ConvertAccountFromModel(model *a.Account) *Account {
	if model == nil || model.Id == uuid.Nil {
		return &Account{}
	}

	return &Account{
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

func (r *AccountConverter) CreateRequestFromProto(request *proto.CreatAccountRequest) (*CreateAccountRequest, *system.Error) {

	result := &CreateAccountRequest{
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

func (r *AccountConverter) CreateResponseProtoFromModel(request *CreateAccountResponse) (*proto.CreateAccountResponse, *system.Error) {

	result := &proto.CreateAccountResponse{
		Account: &proto.AccountResponse{
			Id: proto.FromUUID(request.AccountId),
		},
		Errors: ProtoErrorFromErrorRs(request.Errors),
	}

	return result, nil
}

func (r *AccountConverter) UpdateRequestFromProto(request *proto.UpdateAccountRequest) (*UpdateAccountRequest, *system.Error) {

	result := &UpdateAccountRequest{
		AccountId: AccountIdRequest{
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

func (r *AccountConverter) UpdateResponseProtoFromModel(request *UpdateAccountResponse) (*proto.UpdateAccountResponse, *system.Error) {

	result := &proto.UpdateAccountResponse{
		Errors: ProtoErrorFromErrorRs(request.Errors),
	}

	return result, nil
}

func (r *AccountConverter) GetByCriteriaRequestFromProto(request *proto.GetAccountsByCriteriaRequest) (*GetAccountsByCriteriaRequest, *system.Error) {

	result := &GetAccountsByCriteriaRequest{
		AccountId: AccountIdRequest{
			AccountId:  request.AccountId.AccountId.ToUUID(),
			ExternalId: request.AccountId.ExternalId,
		},
		Email: request.Email,
		Phone: request.Phone,
	}

	return result, nil
}

func (r *AccountConverter) GetByCriteriaResponseProtoFromModel(request *GetAccountsByCriteriaResponse) (*proto.GetAccountsByCriteriaResponse, *system.Error) {

	result := &proto.GetAccountsByCriteriaResponse{
		Accounts: []*proto.AccountItem{},
		Errors:   ProtoErrorFromErrorRs(request.Errors),
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

func (r *AccountConverter) SetOnlineStatusRequestFromProto(request *proto.SetOnlineStatusRequest) (*SetAccountOnlineStatusRequest, *system.Error) {

	result := &SetAccountOnlineStatusRequest{
		Account: &AccountIdRequest{
			AccountId:  request.AccountId.AccountId.ToUUID(),
			ExternalId: request.AccountId.ExternalId,
		},
		Status: request.Status,
	}

	return result, nil
}

func (r *AccountConverter) SetOnlineStatusResponseProtoFromModel(request *SetAccountOnlineStatusResponse) (*proto.SetOnlineStatusResponse, *system.Error) {

	result := &proto.SetOnlineStatusResponse{
		Errors: ProtoErrorFromErrorRs(request.Errors),
	}

	return result, nil
}

func (r *AccountConverter) GetOnlineStatusRequestFromProto(request *proto.GetOnlineStatusRequest) (*GetAccountOnlineStatusRequest, *system.Error) {

	result := &GetAccountOnlineStatusRequest{
		Account: &AccountIdRequest{
			AccountId:  request.AccountId.AccountId.ToUUID(),
			ExternalId: request.AccountId.ExternalId,
		},
	}

	return result, nil
}

func (r *AccountConverter) GetOnlineStatusResponseProtoFromModel(request *GetAccountOnlineStatusResponse) (*proto.GetOnlineStatusResponse, *system.Error) {

	result := &proto.GetOnlineStatusResponse{
		Status: request.Status,
		Errors: ProtoErrorFromErrorRs(request.Errors),
	}

	return result, nil
}
