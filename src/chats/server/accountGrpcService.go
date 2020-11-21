package server

import (
	"chats/proto"
	"context"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type AccountGrpcService struct {
	proto.UnimplementedAccountServer
}

func (s *AccountGrpcService) Create(ctx context.Context, rq *proto.CreatAccountRequest) (*proto.CreateAccountResponse, error) {

	errorRs := &proto.CreateAccountResponse{}
	c := &AccountConverter{}

	modelRq, err := c.CreateRequestFromProto(rq)
	if err != nil {
		errorRs.Errors = []*proto.Error{proto.Err(err)}
		return errorRs, nil
	}

	modelRs, err := Server.createAccount(modelRq)
	if err != nil {
		errorRs.Errors = []*proto.Error{proto.Err(err)}
		return errorRs, nil
	}

	protoRs, err := c.CreateResponseProtoFromModel(modelRs)
	if err != nil {
		errorRs.Errors = []*proto.Error{proto.Err(err)}
		return errorRs, nil
	}

	return protoRs, nil

}

func (s *AccountGrpcService) Update(ctx context.Context, rq *proto.UpdateAccountRequest) (*proto.UpdateAccountResponse, error) {

	errorRs := &proto.UpdateAccountResponse{}
	c := &AccountConverter{}

	modelRq, err := c.UpdateRequestFromProto(rq)
	if err != nil {
		errorRs.Errors = []*proto.Error{proto.Err(err)}
		return errorRs, nil
	}

	modelRs, err := Server.updateAccount(modelRq)
	if err != nil {
		errorRs.Errors = []*proto.Error{proto.Err(err)}
		return errorRs, nil
	}

	protoRs, err := c.UpdateResponseProtoFromModel(modelRs)
	if err != nil {
		errorRs.Errors = []*proto.Error{proto.Err(err)}
		return errorRs, nil
	}

	return protoRs, nil

}

func (s *AccountGrpcService) Lock(ctx context.Context, rq *proto.LockAccountRequest) (*proto.LockAccountResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Lock not implemented")
}

func (s *AccountGrpcService) GetByCriteria(ctx context.Context, rq *proto.GetAccountsByCriteriaRequest) (*proto.GetAccountsByCriteriaResponse, error) {

	errorRs := &proto.GetAccountsByCriteriaResponse{}
	c := &AccountConverter{}

	modelRq, err := c.GetByCriteriaRequestFromProto(rq)
	if err != nil {
		errorRs.Errors = []*proto.Error{proto.Err(err)}
		return errorRs, nil
	}

	modelRs, err := Server.getAccountsByCriteria(modelRq)
	if err != nil {
		errorRs.Errors = []*proto.Error{proto.Err(err)}
		return errorRs, nil
	}

	protoRs, err := c.GetByCriteriaResponseProtoFromModel(modelRs)
	if err != nil {
		errorRs.Errors = []*proto.Error{proto.Err(err)}
		return errorRs, nil
	}

	return protoRs, nil
}

func (s *AccountGrpcService) SetOnlineStatus(ctx context.Context, rq *proto.SetOnlineStatusRequest) (*proto.SetOnlineStatusResponse, error) {
	errorRs := &proto.SetOnlineStatusResponse{}
	c := &AccountConverter{}

	modelRq, err := c.SetOnlineStatusRequestFromProto(rq)
	if err != nil {
		errorRs.Errors = []*proto.Error{proto.Err(err)}
		return errorRs, nil
	}

	modelRs, err := Server.setOnlineStatus(modelRq)
	if err != nil {
		errorRs.Errors = []*proto.Error{proto.Err(err)}
		return errorRs, nil
	}

	protoRs, err := c.SetOnlineStatusResponseProtoFromModel(modelRs)
	if err != nil {
		errorRs.Errors = []*proto.Error{proto.Err(err)}
		return errorRs, nil
	}

	return protoRs, nil
}

func (s *AccountGrpcService) GetOnlineStatus(ctx context.Context, rq *proto.GetOnlineStatusRequest) (*proto.GetOnlineStatusResponse, error) {
	errorRs := &proto.GetOnlineStatusResponse{}
	c := &AccountConverter{}

	modelRq, err := c.GetOnlineStatusRequestFromProto(rq)
	if err != nil {
		errorRs.Errors = []*proto.Error{proto.Err(err)}
		return errorRs, nil
	}

	modelRs, err := Server.getOnlineStatus(modelRq)
	if err != nil {
		errorRs.Errors = []*proto.Error{proto.Err(err)}
		return errorRs, nil
	}

	protoRs, err := c.GetOnlineStatusResponseProtoFromModel(modelRs)
	if err != nil {
		errorRs.Errors = []*proto.Error{proto.Err(err)}
		return errorRs, nil
	}

	return protoRs, nil
}
