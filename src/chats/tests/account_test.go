package tests

import (
	"chats/sdk"
	"encoding/json"
	uuid "github.com/satori/go.uuid"
	"testing"
)

func TestCreateAccount_Success(t *testing.T) {

	sdkService, err := initSdk()
	if err != nil {
		t.Error(err.Error(), sdkService)
	}

	createAccountRq := &sdk.CreateAccountRequest{
		ApiRequestModel: sdk.ApiRequestModel{
			Method: "POST",
			Path:   "/account",
		},
		Body: sdk.CreateAccountRequestBody{
			Account:    "testAccount",
			Type:       "user",
			ExternalId: "112233",
			FirstName:  "Иванов",
			MiddleName: "Иванович",
			LastName:   "Иванов",
			Email:      "ivanov@gmail.com",
			Phone:      "+79107895632",
			AvatarUrl:  "https://s3.adacta.ru/ivanov",
		},
	}

	dataRq, err := json.Marshal(createAccountRq)

	if err != nil {
		t.Error(err.Error(), sdkService)
	}

	rsData, e := sdkService.Subject(topic).Request(dataRq)
	if e != nil {
		t.Error(e.Error, sdkService)
	}

	response := &sdk.CreateAccountResponse{}

	err = json.Unmarshal(rsData, response)
	if err != nil {
		t.Error(err.Error(), sdkService)
	}

	if response.AccountId == uuid.Nil {
		t.Error("AccountId is empty")
	}

}
