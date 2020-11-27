package system

import "fmt"

type Errors map[int]string

type Error struct {
	Error   error  `json:"sentry"`
	Message string `json:"message"`
	Code    int    `json:"code"`
	Data    []byte `json:"data"`
	Stack   string `json:"stack"`
}

func E(ee error) *Error {
	return &Error{Error: ee, Message: ee.Error()}
}

func SysErr(err error, id int, data []byte) *Error {
	return &Error{
		Error:   err,
		Message: GetError(id),
		Code:    id,
		Data:    data,
	}
}

func SysErrf(err error, id int, data []byte, a...interface{}) *Error {
	return &Error{
		Error:   err,
		Message: GetErrorf(id, a),
		Code:    id,
		Data:    data,
	}
}

func GetError(code int) string {
	if _, ok := errList[code]; ok {
		return errList[code]
	}
	return errList[1002]
}

func GetErrorf(code int, a...interface{}) string {
	return fmt.Sprintf(GetError(code) + "\n", a)
}

func MarshalError1011(err error, params []byte) *Error {
	return &Error{
		Error:   err,
		Message: GetError(1011),
		Code:    1011,
		Data:    params,
	}
}

func UnmarshalError1010(err error, params []byte) *Error {
	return abstractError(&Error{
		Error: err,
		Data:  params,
	}, 1010)
}

func UnmarshalRequestError1201(err error, params []byte) *Error {
	return abstractError(&Error{
		Error: err,
		Data:  params,
	}, 1201)
}

func UnmarshalRequestTypeError1204(err error, params []byte) *Error {
	return abstractError(&Error{
		Error: err,
		Data:  params,
	}, 1204)
}

func abstractError(err *Error, code int) *Error {
	err.Message = GetError(code)
	err.Code = code
	return err
}


