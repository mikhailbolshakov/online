package system

const (

	ApplicationPanicCatched    = 100

	ApplicationErrorCode    = 200
	UnmarshallingErrorCode  = 201
	ParseErrorCode          = 205
	LoadLocationErrorCode   = 210
	MessageTooLongErrorCode = 220
	SdkConnectionErrorCode  = 230
	CronResponseErrorCode   = 301

	WsCreateAccountInvalidTypeErrorCode = 2001
	WsCreateAccountEmptyAccountErrorCode = 2002
	WsConnectAccountSystemCode = 2003
	AccountIncorrectOnlineStatus = 2004
	AccountNotFoundById = 2005
	AccountOnlineStatusWithoutLiveConnection = 2006

	NoRoomFoundByIdCode = 3001
	NoRoomFoundByReferenceCode = 3002
	RoomAlreadyClosedCode = 3003
	NotSubscribedAccountCode = 3004

	IncorrectRequestCode = 4000

)

var errList = Errors {
	ApplicationPanicCatched: "Перехвачена паника в методе %s",

	200: "Общая ошибка приложения",
	201: "Ошибка при анмаршаллинге",
	205: "Ошибка установки временной зоны",
	210: "Длина сообщения превышает установленный лимит",
	220: "Длина сообщения превышает установленный лимит",
	230: "Ошибка соединения с шиной",
	301: "Ошибка ответа от cron",

	1000: "Общая ошибка сервиса %s",
	1001: "Ошибка соединения с шиной",
	1002: "Неизвестная ошибка сервиса %s",

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

	WsCreateAccountInvalidTypeErrorCode: "Неверный тип аккаунта",
	WsCreateAccountEmptyAccountErrorCode: "Пустое значение наименование аккаунта",
	WsConnectAccountSystemCode: "Невозможно создать соединение для системного аккаунта",
	AccountIncorrectOnlineStatus: "Некорректный онлайн статус",
	AccountNotFoundById: "Аккаунт не найден по ИД %s",
	AccountOnlineStatusWithoutLiveConnection: "невозможно установить статус %s при отсутствие открытого соединения",

	NoRoomFoundByIdCode: "Комната не найдена по ИД %s",
	NoRoomFoundByReferenceCode: "Комната не найдена по referenceId %s",
	RoomAlreadyClosedCode: "Комната уже закрыта",
	NotSubscribedAccountCode: "Аккаунт %s не подписан на комнату %s",

	IncorrectRequestCode: "Некорректный запрос",

}


const (
	MysqlErrorCode                    = 300 //	общая ошибка
	MysqlConnectionProblem            = "Ошибка соединения с базой данных mysql"
	MysqlConnectionProblemCode        = 301
	MysqlChatCreateError              = "Ошибка при попытке создать новый чат"
	MysqlChatCreateErrorCode          = 321
	MysqlChatUserSubscribeError       = "Ошибка при попытке подписать пользователя на чат"
	MysqlChatUserSubscribeErrorCode   = 322
	MysqlChatUserUnsubscribeError     = "Ошибка при попытке отписать пользователя от чата"
	MysqlChatUserUnsubscribeErrorCode = 323
	MysqlChatCreateMessageError       = "Ошибка при попытке создания нового сообщения"
	MysqlChatCreateMessageErrorCode   = 324
	MysqlChatSubscribeEmpty           = "У чата нет подписчиков"
	MysqlChatSubscribeEmptyCode       = 325
	MysqlChatSubscribeChangeError     = "Ошибка при попытке поменять подписчика на чат"
	MysqlChatSubscribeChangeErrorCode = 326
	MysqlChatNotExists                = "Чат не найден"
	MysqlChatNotExistsCode            = 330
	MysqlChatChangeStatusError        = "Ошибка при попытке изменить статус чата"
	MysqlChatChangeStatusErrorCode    = 331
	MysqlUserChatListError            = "Ошибка при выборе списка чатов пользователя"
	MysqlUserChatListErrorCode        = 340
	MysqlChatInfoError                = "Ошибка при получении информации о чате"
	MysqlChatInfoErrorCode            = 341
	MysqlChatMessageParamsError       = "Ошибка при получении списка параметров сообщения"
	MysqlChatMessageParamsErrorCode   = 342
	MysqlChatIdIncorrect              = "Передан некорректный параметр chatId"
	MysqlChatIdIncorrectCode          = 350
	MysqlChatMessageTypeIncorrect     = "Передан некорректный параметр type"
	MysqlChatMessageTypeIncorrectCode = 351
	MysqlChatAccessDenied             = "Пользователь не подписан на чат"
	PrivateChatRecipientNotFoundAmongSubscribersCode = 352

	MysqlChatAccessDeniedCode         = 501

	DbAccountOnlineStatusUpdateError = "Ошибка при обновлении online статуса"
	DbAccountOnlineStatusUpdateErrorCode = 601
	DbAccountOnlineStatusGetError = "Ошибка при получении online статуса"
	DbAccountOnlineStatusGetErrorCode = 602
	DbAccountOnlineStatusGetMoreThenOneError = "Найдено более одного online статуса"
	DbAccountOnlineStatusGetMoreThenOneErrorCode = 603
	DbAccountOnlineStatusGetNotFoundError = "online статус не найден"
	DbAccountOnlineStatusGetNotFoundErrorCode = 604
)

const (
	RedisConnectionProblem     = "Ошибка соединения с базой данных redis"
	RedisConnectionProblemCode = 501
	RedisSetError              = "Ошибка записи в redis"
	RedisSetErrorCode          = 510
	RedisTTLNotExist           = "Не передана переменная окружения REDIS_TTL"
	RedisTTLNotExistCode       = 520
	RedisGetError              = "Ошибка получения из redis"
	RedisGetErrorCode          = 530
)

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