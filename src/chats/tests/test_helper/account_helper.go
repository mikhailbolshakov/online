package test_helper

import (
	"chats/sdk"
	"encoding/json"
	"errors"
	uuid "github.com/satori/go.uuid"
)

func CreateAccountExt(sdkService *sdk.Sdk, account *sdk.Account) (uuid.UUID, error) {

	createAccountRq := &sdk.CreateAccountRequest{
		ApiRequestModel: sdk.ApiRequestModel{
			Method: "POST",
			Path:   "/accounts/account",
		},
		Body: *account,
	}

	dataRq, err := json.Marshal(createAccountRq)

	if err != nil {
		return uuid.Nil, err
	}

	rsData, e := sdkService.Subject(topic).Request(dataRq)
	if e != nil {
		return uuid.Nil, errors.New(e.Message)
	}

	response := &sdk.CreateAccountResponse{}

	err = json.Unmarshal(rsData, response)
	if err != nil {
		return uuid.Nil, err
	}

	return response.AccountId, nil
}

func CreateAccount(sdkService *sdk.Sdk, externalId string) (uuid.UUID, error) {

	return CreateAccountExt(sdkService, &sdk.Account{
		Account:    "testAccount",
		Type:       "user",
		ExternalId: externalId,
		FirstName:  "Иванов",
		MiddleName: "Иванович",
		LastName:   "Иванов",
		Email:      "ivanov@gmail.com",
		Phone:      "+79107895632",
		AvatarUrl:  "https://s3.adacta.ru/ivanov",
	})
}

func GetAccountById(sdkService *sdk.Sdk, id uuid.UUID) (*sdk.GetAccountResponse, error) {

	getAccountByIdRq := &sdk.GetAccountByIdRequest{
		ApiRequestModel: sdk.ApiRequestModel{
			Method: "GET",
			Path:   "/accounts/account",
		},
		Body:            sdk.GetAccountByIdRequestBody{Account: &sdk.AccountRequest{
			AccountId:  id,
			ExternalId: "",
		}},
	}

	dataRq, err := json.Marshal(getAccountByIdRq)

	if err != nil {
		return nil, err
	}

	rsData, e := sdkService.Subject(topic).Request(dataRq)
	if e != nil {
		return nil, errors.New(e.Message)
	}

	response := &sdk.GetAccountResponse{}

	err = json.Unmarshal(rsData, response)
	if err != nil {
		return nil, err
	}

	return response, nil
}

func GetAccountByExternalId(sdkService *sdk.Sdk, externalId string) (*sdk.GetAccountResponse, error) {

	getAccountByIdRq := &sdk.GetAccountByIdRequest{
		ApiRequestModel: sdk.ApiRequestModel{
			Method: "GET",
			Path:   "/accounts/account",
		},
		Body:            sdk.GetAccountByIdRequestBody{Account: &sdk.AccountRequest{
			AccountId:  uuid.Nil,
			ExternalId: externalId,
		}},
	}

	dataRq, err := json.Marshal(getAccountByIdRq)

	if err != nil {
		return nil, err
	}

	rsData, e := sdkService.Subject(topic).Request(dataRq)
	if e != nil {
		return nil, errors.New(e.Message)
	}

	response := &sdk.GetAccountResponse{}

	err = json.Unmarshal(rsData, response)
	if err != nil {
		return nil, err
	}

	return response, nil
}

func SetAccountOnlineStatus(sdkService *sdk.Sdk, accountId uuid.UUID, status string) (*sdk.BoolResponseData, error) {

	setOnlineStatusRq := &sdk.UpdateAccountOnlineStatusRequest{
		ApiRequestModel: sdk.ApiRequestModel{
			Method: "PUT",
			Path:   "/account/online-status",
		},
		Body:            sdk.AccountOnlineStatus{
			Account: &sdk.AccountRequest{
				AccountId:  accountId,
				ExternalId: "",
			},
			Status:  "online",
		},
	}

	dataRq, err := json.Marshal(setOnlineStatusRq)

	if err != nil {
		return nil, err
	}

	rsData, e := sdkService.Subject(topic).Request(dataRq)
	if e != nil {
		return nil, errors.New(e.Message)
	}

	response := &sdk.BoolResponseData{}

	err = json.Unmarshal(rsData, response)
	if err != nil {
		return nil, err
	}

	return response, nil
}