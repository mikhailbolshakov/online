package server

import (
	"chats/app"
	"chats/system"
	"fmt"
	"log"
	"net"
	"google.golang.org/grpc"
	pb "chats/proto"
)

func (ws *WsServer) Grpc() {

	lis, err := net.Listen("tcp", fmt.Sprintf("localhost:%s", "50051"))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	var opts []grpc.ServerOption
	ws.grpcServer = grpc.NewServer(opts...)

	registration(ws, ws.grpcServer)

	app.L().Debug("Listening GRPC....")
	err = ws.grpcServer.Serve(lis)
	if err != nil {
		app.E().SetError(&system.Error{
			Error: err,
		})
	}

}

func registration(ws *WsServer, s *grpc.Server) {
	pb.RegisterRoomServer(s, &RoomGrpcService{ws: ws})
	pb.RegisterAccountServer(s, &AccountGrpcService{ws: ws})
}
