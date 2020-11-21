package system

import uuid "github.com/satori/go.uuid"

func Uuid() uuid.UUID {
	id, _ := uuid.NewV4()
	return id
}

func Uint8ToBool(i uint8) bool {
	return !(i == 0)
}

func BoolToUint8(i bool) uint8 {
	if i {
		return 1
	} else {
		return 0
	}
}
