package server

import (
	"chats/proto"
)

func ProtoErrorFromErrorRs(source []ErrorResponse) []*proto.Error {
	var result []*proto.Error
	if len(source) > 0 {
		for _, e := range source {
			result = append(result, &proto.Error{
				Code:    int32(e.Code),
				Message: e.Message,
			})
		}
	}
	return result
}