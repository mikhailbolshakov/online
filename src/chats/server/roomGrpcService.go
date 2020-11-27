package server

import (
	"chats/proto"
	"context"
)

type RoomGrpcService struct {
	ws *WsServer
	proto.UnimplementedRoomServer
}

func (s *RoomGrpcService) Create(ctx context.Context, rq *proto.CreateRoomRequest) (*proto.CreateRoomResponse, error) {

	errorRs := &proto.CreateRoomResponse{}
	c := &RoomConverter{}
	modelRq, err := c.CreateRequestFromProto(rq)
	if err != nil {
		errorRs.Errors = []*proto.Error{ proto.Err(err) }
		return errorRs, nil
	}

	modelRs, err := s.ws.CreateRoom(modelRq)
	if err != nil {
		errorRs.Errors = []*proto.Error{ proto.Err(err) }
		return errorRs, nil
	}

	protoRs, err := c.CreateResponseProtoFromModel(modelRs)
	if err != nil {
		errorRs.Errors = []*proto.Error{ proto.Err(err) }
		return errorRs, nil
	}

	return protoRs, nil

}

func (s *RoomGrpcService) Subscribe(ctx context.Context, rq *proto.RoomSubscribeRequest) (*proto.RoomSubscribeResponse, error) {

	errorRs := &proto.RoomSubscribeResponse{}
	c := &RoomConverter{}
	modelRq, err := c.SubscribeRequestFromProto(rq)
	if err != nil {
		errorRs.Errors = []*proto.Error{ proto.Err(err) }
		return errorRs, nil
	}

	modelRs, err := s.ws.RoomSubscribe(modelRq)
	if err != nil {
		errorRs.Errors = []*proto.Error{ proto.Err(err) }
		return errorRs, nil
	}

	protoRs, err := c.SubscribeResponseProtoFromModel(modelRs)
	if err != nil {
		errorRs.Errors = []*proto.Error{ proto.Err(err) }
		return errorRs, nil
	}

	return protoRs, nil
}

func (s *RoomGrpcService) GetByCriteria(ctx context.Context, rq *proto.GetRoomsByCriteriaRequest) (*proto.GetRoomsByCriteriaResponse, error) {

	errorRs := &proto.GetRoomsByCriteriaResponse{}
	c := &RoomConverter{}
	modelRq, err := c.GetByCriteriaRequestFromProto(rq)
	if err != nil {
		errorRs.Errors = []*proto.Error{ proto.Err(err) }
		return errorRs, nil
	}

	modelRs, err := s.ws.GetRoomsByCriteria(modelRq)
	if err != nil {
		errorRs.Errors = []*proto.Error{ proto.Err(err) }
		return errorRs, nil
	}

	protoRs, err := c.GetByCriteriaResponseProtoFromModel(modelRs)
	if err != nil {
		errorRs.Errors = []*proto.Error{ proto.Err(err) }
		return errorRs, nil
	}

	return protoRs, nil
}

func (s *RoomGrpcService) CloseRoom(ctx context.Context, rq *proto.CloseRoomRequest) (*proto.CloseRoomResponse, error) {

	errorRs := &proto.CloseRoomResponse{}
	c := &RoomConverter{}
	modelRq, err := c.CloseRoomRequestFromProto(rq)
	if err != nil {
		errorRs.Errors = []*proto.Error{ proto.Err(err) }
		return errorRs, nil
	}

	modelRs, err := s.ws.CloseRoom(modelRq)
	if err != nil {
		errorRs.Errors = []*proto.Error{ proto.Err(err) }
		return errorRs, nil
	}

	protoRs, err := c.CloseRoomResponseProtoFromModel(modelRs)
	if err != nil {
		errorRs.Errors = []*proto.Error{ proto.Err(err) }
		return errorRs, nil
	}

	return protoRs, nil
}

func (s *RoomGrpcService) SendChatMessages(ctx context.Context, rq *proto.SendChatMessagesRequest) (*proto.SendChatMessageResponse, error) {

	errorRs := &proto.SendChatMessageResponse{}
	c := &RoomConverter{}
	modelRq, err := c.SendChatMessageRequestFromProto(rq)
	if err != nil {
		errorRs.Errors = []*proto.Error{ proto.Err(err) }
		return errorRs, nil
	}

	modelRs, err := s.ws.SendChatMessages(modelRq)
	if err != nil {
		errorRs.Errors = []*proto.Error{ proto.Err(err) }
		return errorRs, nil
	}

	protoRs, err := c.SendChatMessageResponseProtoFromModel(modelRs)
	if err != nil {
		errorRs.Errors = []*proto.Error{ proto.Err(err) }
		return errorRs, nil
	}

	return protoRs, nil
}

func (s *RoomGrpcService) Unsubscribe(ctx context.Context, rq *proto.RoomUnsubscribeRequest) (*proto.RoomUnsubscribeResponse, error) {

	errorRs := &proto.RoomUnsubscribeResponse{}
	c := &RoomConverter{}
	modelRq, err := c.UnsubscribeRequestFromProto(rq)
	if err != nil {
		errorRs.Errors = []*proto.Error{ proto.Err(err) }
		return errorRs, nil
	}

	modelRs, err := s.ws.RoomUnsubscribe(modelRq)
	if err != nil {
		errorRs.Errors = []*proto.Error{ proto.Err(err) }
		return errorRs, nil
	}

	protoRs, err := c.UnsubscribeResponseProtoFromModel(modelRs)
	if err != nil {
		errorRs.Errors = []*proto.Error{ proto.Err(err) }
		return errorRs, nil
	}

	return protoRs, nil
}
