package database

//	mysql errors 300

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
	PinbaConnectionProblem            = "Ошибка соединения с Pinba"
	PinbaConnectionProblemCode        = 381
	MysqlChatAccessDenied             = "Пользователь не подписан на чат"
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
