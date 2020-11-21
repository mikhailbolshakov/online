package sdk

type ErrorResponse struct {
	Message string `json:"message"`
	Code    int    `json:"code"`
	Stack   string `json:"stack"`
}
