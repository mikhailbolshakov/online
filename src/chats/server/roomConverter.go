package server

import (
	"chats/proto"
	"chats/sdk"
	"chats/system"
	uuid "github.com/satori/go.uuid"
)

type RoomConverter struct{}

func (r *RoomConverter) CreateRequestFromProto(request *proto.CreateRoomRequest) (*sdk.CreateRoomRequest, *system.Error) {
	result := &sdk.CreateRoomRequest{
		Room: &sdk.RoomRequest{
			ReferenceId: request.ReferenceId,
			Chat:        request.Chat,
			Video:       request.Video,
			Audio:       request.Audio,
		},
	}

	result.Room.Subscribers = []sdk.SubscriberRequest{}

	for _, item := range request.Subscribers {
		result.Room.Subscribers = append(result.Room.Subscribers, sdk.SubscriberRequest{
			Account: &sdk.AccountIdRequest{
				AccountId:  item.Account.AccountId.ToUUID(),
				ExternalId: item.Account.ExternalId,
			},
			Role: item.Role,
		})
	}

	return result, nil
}

func (r *RoomConverter) CreateResponseProtoFromModel(request *sdk.CreateRoomResponse) (*proto.CreateRoomResponse, *system.Error) {
	result := &proto.CreateRoomResponse{
		Errors: []*proto.Error{},
	}

	if request.Result.Id != uuid.Nil {
		result.Result = &proto.RoomResponse{
			Id: proto.FromUUID(request.Result.Id),
			Hash: request.Result.Hash,
		}
	}

	result.Errors = proto.ErrorRs(request.Errors)

	return result, nil
}

func (r *RoomConverter) SubscribeRequestFromProto(request *proto.RoomSubscribeRequest) (*sdk.RoomSubscribeRequest, *system.Error) {

	result := &sdk.RoomSubscribeRequest{
		RoomId:      request.RoomId.ToUUID(),
		ReferenceId: request.ReferenceId,
		Subscribers: []sdk.SubscriberRequest{},
	}

	for _, item := range request.Subscribers {
		result.Subscribers = append(result.Subscribers, sdk.SubscriberRequest{
			Account: &sdk.AccountIdRequest{
				AccountId:  item.Account.AccountId.ToUUID(),
				ExternalId: item.Account.ExternalId,
			},
			Role: item.Role,
		})
	}

	return result, nil
}

func (r *RoomConverter) SubscribeResponseProtoFromModel(request *sdk.RoomSubscribeResponse) (*proto.RoomSubscribeResponse, *system.Error) {

	result := &proto.RoomSubscribeResponse{
		Errors: []*proto.Error{},
	}

	result.Errors = proto.ErrorRs(request.Errors)

	return result, nil
}

func (r *RoomConverter) GetByCriteriaRequestFromProto(request *proto.GetRoomsByCriteriaRequest) (*sdk.GetRoomsByCriteriaRequest, *system.Error) {
	result := &sdk.GetRoomsByCriteriaRequest{}

	if request.AccountId != nil {
		result.AccountId = &sdk.AccountIdRequest{
			AccountId:  request.AccountId.AccountId.ToUUID(),
			ExternalId: request.AccountId.ExternalId,
		}
	} else {
		result.AccountId = &sdk.AccountIdRequest{}
	}

	result.ReferenceId = request.ReferenceId
	result.RoomId = request.RoomId.ToUUID()
	result.WithClosed = request.WithClosed
	result.WithSubscribers = request.WithSubscribers

	return result, nil
}

func (r *RoomConverter) GetByCriteriaResponseProtoFromModel(request *sdk.GetRoomsByCriteriaResponse) (*proto.GetRoomsByCriteriaResponse, *system.Error) {

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

func (r *RoomConverter) CloseRoomRequestFromProto(request *proto.CloseRoomRequest) (*sdk.CloseRoomRequest, *system.Error) {

	result := &sdk.CloseRoomRequest{
		RoomId:      request.RoomId.ToUUID(),
		ReferenceId: request.ReferenceId,
	}

	return result, nil
}

func (r *RoomConverter) CloseRoomResponseProtoFromModel(request *sdk.CloseRoomResponse) (*proto.CloseRoomResponse, *system.Error) {

	result := &proto.CloseRoomResponse{
		Errors: proto.ErrorRs(request.Errors),
	}

	return result, nil

}