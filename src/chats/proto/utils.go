package proto

import (
	"chats/app"
	"chats/system"
	uuid "github.com/satori/go.uuid"
	"time"
)

func (u *UUID) ToUUID() uuid.UUID {

	if u == nil || u.Value == "" {
		return uuid.Nil
	} else {
		value, err := uuid.FromString(u.Value)
		if err != nil {
			app.E().SetError(&system.Error{
				Message: "UUID convertion error",
				Error:   err,
			})
		}
		return value
	}
}

func FromUUID(u uuid.UUID) *UUID {
	if u == uuid.Nil {
		return &UUID{Value: ""}
	} else {
		return &UUID{Value: u.String()}
	}
}

func ToTimestamp(t *time.Time) *Timestamp {
	if t != nil {
		return &Timestamp{
			Value: t.String(),
		}
	} else {
		return nil
	}
}

func Err(e *system.Error) *Error {
	return &Error{
		Code:    int32(e.Code),
		Message: e.Message,
	}
}
