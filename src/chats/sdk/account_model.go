package sdk

import uuid "github.com/satori/go.uuid"

type AccountRequest struct {
	AccountId  uuid.UUID `json:"account_id"`
	ExternalId string    `json:"external_id"`
}

type CreateAccountRequest struct {
	ApiRequestModel
	Body Account `json:"body"`
}

type CreateAccountResponse struct {
	AccountId uuid.UUID `json:"account_id"`
}

type GetAccountByIdRequestBody struct {
	Account *AccountRequest `json:"account"`
}

type GetAccountByIdRequest struct {
	ApiRequestModel
	Body GetAccountByIdRequestBody `json:"body"`
}

type Account struct {
	Id		   uuid.UUID `json:"id"`
	Account    string `json:"account"`
	Type       string `json:"type"`
	ExternalId string `json:"externalId"`
	FirstName  string `json:"firstName"`
	MiddleName string `json:"middleName"`
	LastName   string `json:"lastName"`
	Email      string `json:"email"`
	Phone      string `json:"phone"`
	AvatarUrl  string `json:"avatarUrl"`
}

type GetAccountResponse struct {
	Data Account `json:"data"`
}

type AccountOnlineStatus struct {
	Account *AccountRequest `json:"account"`
	Status string `json:"status"`
}

type UpdateAccountOnlineStatusRequest struct {
	ApiRequestModel
	Body AccountOnlineStatus `json:"body"`
}