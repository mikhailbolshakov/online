package tests

import (
	"chats/sdk"
	"encoding/json"
	uuid "github.com/satori/go.uuid"
	"testing"
)

func TestNewChat_Success(t *testing.T) {

	sdkService, err := initSdk()
	if err != nil {
		t.Error(err.Error())
	}

	newChatRq := sdk.ChatNewRequest{
		ApiRequestModel: sdk.ApiRequestModel{
			Method: "POST",
			Path:   "/chats/new",
		},
		Body:            sdk.ChatNewRequestBody{
			ReferenceId: "123",
		},
	}

	dataRq, err := json.Marshal(newChatRq)
	if err != nil {
		t.Error(err.Error())
	}

	dataRs, e := sdkService.
		Subject(topic).
		Request(dataRq)
	if e != nil {
		t.Errorf("%s %d %s", e.Error, e.Code, e.Message)
	}

	newChatRs := &sdk.ChatNewResponse{Data: sdk.ChatNewResponseData{}}
	err = json.Unmarshal(dataRs, newChatRs)
	if err != nil {
		t.Error(err)
	}

	if newChatRs.Data.ChatId == uuid.Nil {
		t.Error("Chat creation error")
	}

}
