package sdk

import (
	"encoding/json"
	"errors"
	"fmt"
	uuid "github.com/satori/go.uuid"
	"net/http"
	"strconv"
	"strings"
)

//	Deprecated
func (sdk *Sdk) UserPushToken(userId uint) (result *ApiUserPushTokenResponseData, err *Error) {
	defer sdk.catchPanic("UserPushToken")
	sdk.subject = personalTopic

	id := strconv.FormatUint(uint64(userId), 10)
	path := strings.Replace(userPushTokenPath, "%d", id, 1)

	requestPrepare := &ApiRequest{
		Method: http.MethodGet,
		Path:   path,
	}

	request, marshalError := json.Marshal(requestPrepare)
	if marshalError != nil {
		return nil, sdk.errorLog(sdk.Error(marshalError, 1202, []byte(requestPrepare.String())))
	}

	responsePrepare, responseError := sdk.request(request)
	if responseError != nil {
		return nil, responseError
	}

	response := &ApiUserPushTokenResponse{}

	unmarshalError := json.Unmarshal(responsePrepare, response)
	if unmarshalError != nil {
		return nil, sdk.errorLog(sdk.Error(unmarshalError, 1010, responsePrepare))
	}

	return &response.Data, nil
}

func (sdk *Sdk) UserPush(pushRequest *ApiUserPushRequest) *Error {
	defer sdk.catchPanic("UserPush")
	sdk.subject = communicationTopic

	requestPrepare := &ApiRequest{
		Method: http.MethodPost,
		Path:   communicationPushPath,
		Body:   pushRequest,
	}

	request, err := json.Marshal(requestPrepare)
	if err != nil {
		return sdk.errorLog(sdk.Error(err, 1202, []byte(requestPrepare.String())))
	}

	sdk.request(request)

	return nil
}

func (sdk *Sdk) UserByToken(token string, account *AccountModel) *Error {
	defer sdk.catchPanic("UserByToken")
	sdk.subject = userTopic

	userToken := &UserTokenModel{AccessToken: token}

	requestPrepare := &ApiRequest{
		Method: http.MethodPost,
		Path:   userTokenPath,
		Body:   userToken,
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
		sdk.setlog("unmarshalError step 1", nil, []byte{})
		return sdk.errorLog(sdk.Error(unmarshalError, 1010, responsePrepare))
	}

	if response["sentry"] != nil {
		errorResponse := &ApiErrorResponse{}
		unmarshalError := json.Unmarshal(responsePrepare, errorResponse)
		if unmarshalError != nil {
			sdk.setlog("unmarshalError step 2", nil, []byte{})
			return sdk.errorLog(sdk.Error(unmarshalError, 1010, responsePrepare))
		}
		return sdk.errorLog(sdk.Error(errors.New(errorResponse.Error.Message), 1011, responsePrepare))
	}

	successResponse := &ApiAccountResponse{}
	unmarshalError = json.Unmarshal(responsePrepare, successResponse)
	if unmarshalError != nil {
		sdk.setlog("unmarshalError step 3", nil, []byte{})
		return sdk.errorLog(sdk.Error(unmarshalError, 1010, responsePrepare))
	}

	account.Id = successResponse.Data.Id
	//	#userCrutch
	/*
		account.FirstName = successResponse.Data.FirstName
		account.LastName = successResponse.Data.LastName
		account.MiddleName = successResponse.Data.MiddleName
		account.Photo = successResponse.Data.Photo
	*/

	return nil
}

//	deprecated
func (sdk *Sdk) UserById(user *AccountModel) (err *Error) {
	defer sdk.catchPanic("UserById")
	sdk.subject = personalTopic
	userIdModel := &AccountIdModel{Id: user.Id}
	requestPrepare := &ApiRequest{
		Method: http.MethodGet,
		Path:   fmt.Sprintf(userIdPath, userIdModel.Id),
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
		sdk.setlog("unmarshalError step 1", nil, []byte{})
		return sdk.errorLog(sdk.Error(unmarshalError, 1010, responsePrepare))
	}

	if response["sentry"] != nil {
		errorResponse := &ApiErrorResponse{}
		unmarshalError := json.Unmarshal(responsePrepare, errorResponse)
		if unmarshalError != nil {
			sdk.setlog("unmarshalError step 2", nil, []byte{})
			return sdk.errorLog(sdk.Error(unmarshalError, 1010, responsePrepare))
		}
		return sdk.errorLog(sdk.Error(errors.New(errorResponse.Error.Message), 1011, responsePrepare))
	}

	successResponse := &ApiPersonalResponse{}
	unmarshalError = json.Unmarshal(responsePrepare, successResponse)
	if unmarshalError != nil {
		sdk.setlog("unmarshalError step 3", nil, []byte{})
		return sdk.errorLog(sdk.Error(unmarshalError, 1010, responsePrepare))
	}

	user.Id = successResponse.Data.AccountId
	user.FirstName = successResponse.Data.FirstName
	user.LastName = successResponse.Data.LastName
	user.MiddleName = successResponse.Data.MiddleName
	user.Photo = successResponse.Data.Photo

	return nil
}

func (sdk *Sdk) VagueUserById(user *AccountModel, userType string, referenceId string) (error *Error) {
	switch userType {
	case "client":
		sdk.patientByConsultation(0 /*consultationId*/, user)
		break
	case "doctor":
		sdk.DoctorById(user)
		break
	case "operator":
		sdk.OperatorById(user)
		break
	case "bot":
		sdk.OperatorById(user)
		break
	}

	return nil
}

func (sdk *Sdk) PatientById(user *AccountModel, patientId uuid.UUID) (err *Error) {
	defer sdk.catchPanic("PatientById")
	sdk.subject = personalTopic

	userId := user.Id

	user.Id = patientId

	requestPrepare := &ApiRequest{
		Method: http.MethodGet,
		Path:   fmt.Sprintf(patientIdPath, patientId),
	}

	err = sdk.abstractUserById(user, requestPrepare)
	if err != nil {
		return err
	}

	user.Id = userId

	return nil
}

func (sdk *Sdk) OperatorById(user *AccountModel) (err *Error) {
	defer sdk.catchPanic("OperatorById")
	sdk.subject = personalTopic

	requestPrepare := &ApiRequest{
		Method: http.MethodGet,
		Path:   fmt.Sprintf(personalUserIdPath, user.Id),
	}

	err = sdk.abstractUserById(user, requestPrepare)
	if err != nil {
		return err
	}
	return nil
}

func (sdk *Sdk) DoctorById(account *AccountModel) (err *Error) {
	defer sdk.catchPanic("DoctorById")
	sdk.subject = doctorsTopic

	requestPrepare := &ApiRequest{
		Method: http.MethodGet,
		Path:   doctorIdPath,
	}
	doctorId := uuid.UUID.String(account.Id)
	requestPrepare.Path = strings.Replace(requestPrepare.Path, "%d", doctorId, 1)

	err = sdk.abstractUserById(account, requestPrepare)
	if err != nil {
		return err
	}
	return nil
}

func (sdk *Sdk) abstractUserById(user *AccountModel, requestPrepare *ApiRequest) (err *Error) {
	userIdModel := &AccountIdModel{Id: user.Id}
	userId := uuid.UUID.String(userIdModel.Id)
	requestPrepare.Path = strings.Replace(requestPrepare.Path, "%d", userId, 1)

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

	successResponse := &ApiPersonalResponse{}
	unmarshalError = json.Unmarshal(responsePrepare, successResponse)
	if unmarshalError != nil {
		return sdk.errorLog(sdk.Error(unmarshalError, 1010, responsePrepare))
	}

	user.Id = successResponse.Data.AccountId
	user.FirstName = successResponse.Data.FirstName
	user.LastName = successResponse.Data.LastName
	user.MiddleName = successResponse.Data.MiddleName
	user.Photo = successResponse.Data.Photo

	return nil
}

func (sdk *Sdk) patientByConsultation(consultationId uint, user *AccountModel) {
	consultation, err := sdk.consultation(consultationId)
	if err != nil {
		return
	}

	consultationResponse := &ConsultationResponseResponse{}

	unmarshalErr := json.Unmarshal(consultation, consultationResponse)
	if unmarshalErr != nil {
		return
	}

	if consultationResponse.Data.PatientId != uuid.Nil {
		sdk.PatientById(user, consultationResponse.Data.PatientId)
	} else {
		sdk.UserById(user) //	todo
	}
}
