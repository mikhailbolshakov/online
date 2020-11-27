package app

import (
	"chats/system"
)

type ErrorHandler struct {
	Sentry *Sentry
}

func (h *ErrorHandler) SetError(err *system.Error) *system.Error {

	if h.Sentry != nil {
		h.Sentry.SetError(err)
	} else {
		L().Debug("Tools.Error message:", err.Message, "Code:", err.Code, "Data:", string(err.Data))
	}
	return err
}

func (h *ErrorHandler) SetPanic(f func()) {

	if h.Sentry != nil {
		h.Sentry.SetPanic(f)
	} else {
		defer func() {
			if err := recover(); err != nil {
				L().Debug("Panic!!!!", err)
			}
		}()
		f()
	}
}

func (h *ErrorHandler) CatchPanic(method string) {
	if r := recover(); r != nil {
		_, ok := r.(error)
		if ok {
			h.SetError(system.SysErrf(nil, system.ApplicationPanicCatched, nil, method))
		}
	}
}


func (h *ErrorHandler) Close() {
	if h.Sentry != nil {
		h.Sentry.Close()
	}
}
