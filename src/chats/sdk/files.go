package sdk

import (
	uuid "github.com/satori/go.uuid"
	"net/http"
	"encoding/json"
	"errors"
)

func (fm *FileModel) string() string {
	return "Id: " + fm.Id
}

func (sdk *Sdk) File(file *FileModel, chatId, accountId uuid.UUID) (err *Error) {
	defer sdk.catchPanic("File")
	sdk.subject = storageTopic

	if len(file.Id) == 0 {
		return sdk.errorLog(sdk.Error(nil, 1401, []byte(file.string())))
	} else if chatId == uuid.Nil {
		return sdk.errorLog(sdk.Error(nil, 1402, []byte(file.string())))
	} else if accountId == uuid.Nil {
		return sdk.errorLog(sdk.Error(nil, 1403, []byte(file.string())))
	}

	requestPrepare := &ApiRequest{
		Method: http.MethodGet,
		Path:   storageFilePath,
		Body: ApiFileRequest{
			FileId:    file.Id,
			ChatId:    chatId,
			AccountId: accountId,
		},
	}

	request, marshalError := json.Marshal(requestPrepare)
	if marshalError != nil {
		return sdk.errorLog(sdk.Error(marshalError, 1202, []byte(requestPrepare.String())))
	}

	responsePrepare, err := sdk.request(request)
	if err != nil {
		return err
	}

	var response map[string]interface{}
	unmarshalError := json.Unmarshal(responsePrepare, &response)
	if unmarshalError != nil {
		return sdk.errorLog(sdk.Error(unmarshalError, 1010, responsePrepare))
	}

	if response["sentry"] != nil {
		errorResponse := &ApiErrorResponse{}
		unmarshalError := json.Unmarshal(responsePrepare, errorResponse)
		if unmarshalError != nil {
			return sdk.errorLog(sdk.Error(unmarshalError, 1010, responsePrepare))
		}
		return sdk.errorLog(sdk.Error(errors.New(errorResponse.Error.Message), 1011, responsePrepare))
	}

	successResponse := &ApiFileResponse{}
	unmarshalError = json.Unmarshal(responsePrepare, successResponse)
	if unmarshalError != nil {
		return sdk.errorLog(sdk.Error(unmarshalError, 1010, responsePrepare))
	}

	*file = successResponse.Data

	return nil
}
