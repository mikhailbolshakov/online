package server

import (
	"chats/system"
	"fmt"
	"log"
	"net"

	"google.golang.org/grpc"
	//"github.com/golang/protobuf/proto"
	pb "chats/proto"
)

func (ws *WsServer) Grpc() {

	lis, err := net.Listen("tcp", fmt.Sprintf("localhost:%s", "50051"))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	var opts []grpc.ServerOption
	grpcServer := grpc.NewServer(opts...)

	registration(ws, grpcServer)

	log.Println("Listening GRPC....")
	err = grpcServer.Serve(lis)
	if err != nil {
		system.ErrHandler.SetError(&system.Error{
			Error: err,
		})
	}

}

func registration(ws *WsServer, s *grpc.Server) {
	pb.RegisterRoomServer(s, &RoomGrpcService{ws: ws})
	pb.RegisterAccountServer(s, &AccountGrpcService{ws: ws})
}
