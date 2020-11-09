package sentry

type SystemError struct {
	Error   error  `json:"sentry"`
	Message string `json:"message"`
	Code    int    `json:"code"`
	Data    []byte `json:"data"`
}
