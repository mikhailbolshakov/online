package server

import (
	"chats/proto"
	"chats/system"
	uuid "github.com/satori/go.uuid"
)

type RoomConverter struct{}

func (r *RoomConverter) CreateRequestFromProto(request *proto.CreateRoomRequest) (*CreateRoomRequest, *system.Error) {
	result := &CreateRoomRequest{
		Room: &RoomRequest{
			ReferenceId: request.ReferenceId,
			Chat:        request.Chat,
			Video:       request.Video,
			Audio:       request.Audio,
		},
	}

	result.Room.Subscribers = []SubscriberRequest{}

	for _, item := range request.Subscribers {
		result.Room.Subscribers = append(result.Room.Subscribers, SubscriberRequest{
			Account: &AccountIdRequest{
				AccountId:  item.Account.AccountId.ToUUID(),
				ExternalId: item.Account.ExternalId,
			},
			Role: item.Role,
			AsSystemAccount: item.AsSystemAccount,
		})
	}

	return result, nil
}

func (r *RoomConverter) CreateResponseProtoFromModel(request *CreateRoomResponse) (*proto.CreateRoomResponse, *system.Error) {
	result := &proto.CreateRoomResponse{
		Errors: []*proto.Error{},
	}

	if request.Result.Id != uuid.Nil {
		result.Result = &proto.RoomResponse{
			Id: proto.FromUUID(request.Result.Id),
			Hash: request.Result.Hash,
		}
	}

	result.Errors = ProtoErrorFromErrorRs(request.Errors)

	return result, nil
}

func (r *RoomConverter) SubscribeRequestFromProto(request *proto.RoomSubscribeRequest) (*RoomSubscribeRequest, *system.Error) {

	result := &RoomSubscribeRequest{
		RoomId:      request.RoomId.ToUUID(),
		ReferenceId: request.ReferenceId,
		Subscribers: []SubscriberRequest{},
	}

	for _, item := range request.Subscribers {
		result.Subscribers = append(result.Subscribers, SubscriberRequest{
			Account: &AccountIdRequest{
				AccountId:  item.Account.AccountId.ToUUID(),
				ExternalId: item.Account.ExternalId,
			},
			Role: item.Role,
			AsSystemAccount: item.AsSystemAccount,
		})
	}

	return result, nil
}

func (r *RoomConverter) SubscribeResponseProtoFromModel(request *RoomSubscribeResponse) (*proto.RoomSubscribeResponse, *system.Error) {

	result := &proto.RoomSubscribeResponse{
		Errors: []*proto.Error{},
	}

	result.Errors = ProtoErrorFromErrorRs(request.Errors)

	return result, nil
}

func (r *RoomConverter) GetByCriteriaRequestFromProto(request *proto.GetRoomsByCriteriaRequest) (*GetRoomsByCriteriaRequest, *system.Error) {
	result := &GetRoomsByCriteriaRequest{}

	if request.AccountId != nil {
		result.AccountId = &AccountIdRequest{
			AccountId:  request.AccountId.AccountId.ToUUID(),
			ExternalId: request.AccountId.ExternalId,
		}
	} else {
		result.AccountId = &AccountIdRequest{}
	}

	result.ReferenceId = request.ReferenceId
	result.RoomId = request.RoomId.ToUUID()
	result.WithClosed = request.WithClosed
	result.WithSubscribers = request.WithSubscribers

	return result, nil
}

func (r *RoomConverter) GetByCriteriaResponseProtoFromModel(request *GetRoomsByCriteriaResponse) (*proto.GetRoomsByCriteriaResponse, *system.Error) {

	result := &proto.GetRoomsByCriteriaResponse{
		Rooms: []*proto.GetRoomResponse{},
		Errors: []*proto.Error{},
	}

	for _, item := range request.Rooms {
		room := &proto.GetRoomResponse{
			Id:          proto.FromUUID(item.Id),
			Hash:        item.Hash,
			ReferenceId: item.ReferenceId,
			Chat:        item.Chat,
			Video:       item.Video,
			Audio:       item.Audio,
			ClosedAt:    proto.ToTimestamp(item.ClosedAt),
			Subscribers: []*proto.GetSubscriberResponse{},
		}

		for _, s := range item.Subscribers {
			room.Subscribers = append(room.Subscribers, &proto.GetSubscriberResponse{
				Id:            proto.FromUUID(s.Id),
				AccountId:     proto.FromUUID(s.AccountId),
				Role:          s.Role,
				UnSubscribeAt: proto.ToTimestamp(s.UnSubscribeAt),
			})
		}

		result.Rooms = append(result.Rooms, room)
	}

	return result, nil

}

func (r *RoomConverter) CloseRoomRequestFromProto(request *proto.CloseRoomRequest) (*CloseRoomRequest, *system.Error) {

	result := &CloseRoomRequest{
		RoomId:      request.RoomId.ToUUID(),
		ReferenceId: request.ReferenceId,
	}

	return result, nil
}

func (r *RoomConverter) CloseRoomResponseProtoFromModel(request *CloseRoomResponse) (*proto.CloseRoomResponse, *system.Error) {

	result := &proto.CloseRoomResponse{
		Errors: ProtoErrorFromErrorRs(request.Errors),
	}

	return result, nil

}

func (r *RoomConverter) SendChatMessageRequestFromProto(request *proto.SendChatMessagesRequest) (*SendChatMessagesRequest, *system.Error) {

	result := &SendChatMessagesRequest{
		SenderAccountId: request.SenderAccountId.ToUUID(),
		Type:            request.Type,
		Data:            SendChatMessagesDataRequest{
			Messages: []SendChatMessageDataRequest{},
		},
	}

	for _, m := range request.Data.Messages {
		result.Data.Messages = append(result.Data.Messages, SendChatMessageDataRequest{
			ClientMessageId:    m.ClientMessageId,
			RoomId:             m.RoomId.ToUUID(),
			Type:               m.Type,
			Text:               m.Text,
			Params:             m.Params,
			RecipientAccountId: m.RecipientAccountId.ToUUID(),
		})
	}

	return result, nil
}

func (r *RoomConverter) SendChatMessageResponseProtoFromModel(request *SendChatMessageResponse) (*proto.SendChatMessageResponse, *system.Error) {

	result := &proto.SendChatMessageResponse{
		Errors: ProtoErrorFromErrorRs(request.Errors),
	}

	return result, nil

}

func (r *RoomConverter) UnsubscribeRequestFromProto(request *proto.RoomUnsubscribeRequest) (*RoomUnsubscribeRequest, *system.Error) {

	result := &RoomUnsubscribeRequest{
		RoomId:      request.RoomId.ToUUID(),
		ReferenceId: request.ReferenceId,
		AccountId:   AccountIdRequest{
			AccountId:  request.AccountId.AccountId.ToUUID(),
			ExternalId: request.AccountId.ExternalId,
		},
	}

	return result, nil
}

func (r *RoomConverter) UnsubscribeResponseProtoFromModel(request *RoomUnsubscribeResponse) (*proto.RoomUnsubscribeResponse, *system.Error) {

	result := &proto.RoomUnsubscribeResponse{
		Errors: ProtoErrorFromErrorRs(request.Errors),
	}

	return result, nil

}