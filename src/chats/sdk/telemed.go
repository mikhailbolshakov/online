package sdk

import (
	"encoding/json"
	"fmt"
	uuid "github.com/satori/go.uuid"
	"net/http"
	"strconv"
	"strings"
)

func (sdk *Sdk) consultation(consultationId uint) ([]byte, *Error) {
	defer sdk.catchPanic("Consultation")
	message := &ApiRequest{
		Method: http.MethodGet,
		Path:   consultationPath,
	}
	consultation := strconv.FormatUint(uint64(consultationId), 10)
	message.Path = strings.Replace(message.Path, "%d", consultation, 1)

	request, marshalErr := json.Marshal(message)
	if marshalErr != nil {
		return nil, sdk.errorLog(sdk.Error(marshalErr, 1202, []byte(message.String())))
	}

	responsePrepare, err := sdk.Subject(telemedTopic).request(request)
	if err != nil {
		return nil, err
	}

	return responsePrepare, nil
}

func (sdk *Sdk) ChangeConnectionStatus(userId uuid.UUID, online bool) {
	defer sdk.catchPanic("ChangeConnectionStatus")
	message := &ApiRequest{
		Method: http.MethodPut,
		Path:   consultationWebsocketPath,
		Body: &ApiTelemedRequest{
			AccountId: userId,
			Online:    online,
		},
	}

	request, err := json.Marshal(message)
	if err != nil {
		sdk.errorLog(sdk.Error(err, 1202, []byte(message.String())))
		return
	}

	fmt.Println("ChangeConnectionStatus.sdk.subject:", telemedTopic, string(request))
	sdk.Subject(telemedTopic).request(request)
}

func (sdk *Sdk) UserConsultationJoin(referenceId string, userId uuid.UUID) {
	defer sdk.catchPanic("UserConsultationJoin")

	type order struct {
		ReferenceId string `json:"reference_id"`
		AccountId   uuid.UUID `json:"account_id"`
	}

	message := &ApiRequest{
		Method: http.MethodPut,
		Path:   consultationWebsocketPath,
		Body: &order{
			ReferenceId: referenceId,
			AccountId:   userId,
		},
	}

	request, err := json.Marshal(message)
	if err != nil {
		sdk.errorLog(sdk.Error(err, 1202, []byte(message.String())))
		return
	}

	fmt.Println("UserConsultationJoin.sdk.subject:", telemedTopic, string(request))
	sdk.Subject(telemedTopic).request(request)
}
