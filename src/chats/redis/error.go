package redis

//	redis errors 500

const (
	RedisConnectionProblem     = "Ошибка соединения с базой данных redis"
	RedisConnectionProblemCode = 501
	RedisSetError              = "Ошибка записи в redis"
	RedisSetErrorCode          = 510
	RedisTTLNotExist           = "Не передана переменная окружения REDIS_TTL"
	RedisTTLNotExistCode       = 520
)
