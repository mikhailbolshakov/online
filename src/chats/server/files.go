package server
//
//import (
//	"chats/system"
//	uuid "github.com/satori/go.uuid"
//)
//
//func (fm *FileModel) string() string {
//	return "Id: " + fm.Id
//}
//
//func File(file *FileModel, chatId, accountId uuid.UUID) (err *system.Error) {
//	return nil
//	//defer catchPanic("File")
//	//subject = storageTopic
//	//
//	//if len(file.Id) == 0 {
//	//	return errorLog(Error(nil, 1401, []byte(file.string())))
//	//} else if chatId == uuid.Nil {
//	//	return errorLog(Error(nil, 1402, []byte(file.string())))
//	//} else if accountId == uuid.Nil {
//	//	return errorLog(Error(nil, 1403, []byte(file.string())))
//	//}
//	//
//	//requestPrepare := &ApiRequest{
//	//	Method: http.MethodGet,
//	//	Path:   storageFilePath,
//	//	Body: ApiFileRequest{
//	//		FileId:    file.Id,
//	//		ChatId:    chatId,
//	//		AccountId: accountId,
//	//	},
//	//}
//	//
//	//request, marshalError := json.Marshal(requestPrepare)
//	//if marshalError != nil {
//	//	return errorLog(Error(marshalError, 1202, []byte(requestPrepare.String())))
//	//}
//	//
//	//responsePrepare, err := request(request)
//	//if err != nil {
//	//	return err
//	//}
//	//
//	//var response map[string]interface{}
//	//unmarshalError := json.Unmarshal(responsePrepare, &response)
//	//if unmarshalError != nil {
//	//	return errorLog(Error(unmarshalError, 1010, responsePrepare))
//	//}
//	//
//	//if response["sentry"] != nil {
//	//	errorResponse := &ApiErrorResponse{}
//	//	unmarshalError := json.Unmarshal(responsePrepare, errorResponse)
//	//	if unmarshalError != nil {
//	//		return errorLog(Error(unmarshalError, 1010, responsePrepare))
//	//	}
//	//	return errorLog(Error(errors.New(errorResponse.Error.Message), 1011, responsePrepare))
//	//}
//	//
//	//successResponse := &ApiFileResponse{}
//	//unmarshalError = json.Unmarshal(responsePrepare, successResponse)
//	//if unmarshalError != nil {
//	//	return errorLog(Error(unmarshalError, 1010, responsePrepare))
//	//}
//	//
//	//*file = successResponse.Data
//	//
//	//return nil
//}
