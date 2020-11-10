package sdk

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
)

//	get first specialization for doctor by userId
func (sdk *Sdk) DoctorSpecialization(userId uint) (result *ApiDoctorSpecializationResponesData, err *Error) {
	defer sdk.catchPanic("DoctorSpecialization")
	sdk.subject = doctorsTopic

	id := strconv.FormatUint(uint64(userId), 10)
	path := strings.Replace(doctorsSpecializationPath, "%d", id, 1)

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

	response := &ApiDoctorSpecializationRespones{}

	unmarshalError := json.Unmarshal(responsePrepare, response)
	if unmarshalError != nil {
		return nil, sdk.errorLog(sdk.Error(unmarshalError, 1010, responsePrepare))
	}

	return &response.Data, nil
}
