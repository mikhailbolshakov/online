package server

//	websocketError 100

const (
	WsEmptyToken                   = "Не указан token"
	WsEmptyTokenCode               = 101
	WsCreateClientResponse         = "Ошибка при формировании ответа клиенту"
	WsCreateClientResponseCode     = 102
	WsUserIdentification           = "Ошибка при получении пользователя по токену"
	WsUserIdentificationCode       = 103
	WsUpgradeProblem               = "Ошибка обновления соединения"
	WsUpgradeProblemCode           = 104
	WsUniqueIdGenerateProblem      = "Ошибка генерации уникального идентификатора пользователя"
	WsUniqueIdGenerateProblemCode  = 105
	WsEventTypeNotExists           = "Вызываемое событие не поддерживается"
	WsEventTypeNotExistsCode       = 106
	WsChangeMessageStatusError     = "Ошибка при изменение статуса сообщения"
	WsChangeMessageStatusErrorCode = 107
	WsSendMessageError             = ">>> PANIC: Ошибка при отправке сообщения в web-socket"
	WsSendMessageErrorCode         = 108
	WsConnReadMessageError         = ">>> FATAL: Ошибка чтения сообщения"
	WsConnReadMessageErrorCode     = 109
)
