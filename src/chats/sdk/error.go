package sdk

type SdkErrors map[int]string

// TODO: change to system.Error
type Error struct {
	Error   error
	Message string
	Code    int
	Data    []byte
}

// TODO: move to system
var errList = SdkErrors{
	1000: "Общая ошибка сервиса " + serviceName,
	1001: "Ошибка соединения с шиной",
	1002: "Неизвестная ошибка сервиса " + serviceName,

	1010: "Ошибка при анмаршалинге",
	1011: "Ошибка ответа", //	ResponseError
	1012: "Пустое тело ответа",
	1101: "Не указан топик",

	1201: "Ошибка запроса", //	RequestError
	1202: "Ошибка подготовки запроса",
	1203: "Ошибка запроса: Метод не найден",
	1204: "Ошибка запроса: Тип не найден",

	1300: "", //	user
	1400: "", //	file
	1401: "Передан некорректный параметр fileId",
	1402: "Передан некорректный параметр chatId",
	1403: "Передан некорректный параметр userId",
	1404: "Передан некорректный параметр type",
	1405: "Передан некорректный параметр orderId",
}

func GetError(code int) string {
	if _, ok := errList[code]; ok {
		return errList[code]
	}
	return errList[1002]
}

func (sdk *Sdk) Error(err error, id int, data []byte) *Error {
	return &Error{
		Error:   err,
		Message: GetError(id),
		Code:    id,
		Data:    data,
	}
}
