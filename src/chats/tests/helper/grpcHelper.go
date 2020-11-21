package helper

import (
	"google.golang.org/grpc"
	"log"
)

const (
	address     = "localhost:50051"
)

func GrpcConnection() (*grpc.ClientConn, error) {
	conn, err := grpc.Dial(address, grpc.WithInsecure(), grpc.WithBlock())
	if err != nil {
		log.Fatalf("GRPC connection failed: %v", err)
		return nil, err
	}

	return conn, nil
}
