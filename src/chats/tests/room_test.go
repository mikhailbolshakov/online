package tests

import (
	pb "chats/proto"
	"chats/system"
	"chats/tests/helper"
	"context"
	"log"
	"testing"
)

func TestCreateRoomWithTwoSubscribers_Success(t *testing.T) {

	conn, err := helper.GrpcConnection()
	if err != nil {
		t.Fatal(err.Error())
	}
	defer conn.Close()

	roomService := pb.NewRoomClient(conn)

	accountIdClient, externalIdClient, err := helper.CreateDefaultAccount(conn)
	if err != nil {
		t.Fatal(err.Error())
	}

	accountIdOperator, externalIdOperator, err := helper.CreateDefaultAccount(conn)
	if err != nil {
		t.Fatal(err.Error())
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	referenceId := system.Uuid().String()
	r, err := roomService.Create(ctx, &pb.CreateRoomRequest{
		ReferenceId: referenceId,
		Chat:        true,
		Video:       false,
		Audio:       false,
		Subscribers: []*pb.SubscriberRequest{
			{
				Account: &pb.AccountIdRequest{
					AccountId: pb.FromUUID(accountIdClient),
				},
				Role: "client",
			},
			{
				Account: &pb.AccountIdRequest{
					ExternalId: externalIdOperator,
				},
				Role: "operator",
			},
		},
	})
	if err != nil {
		t.Fatalf("Error: %v", err)
	}
	log.Printf("Response: %v", r)

	// get by accountIdClient
	getRs, err := roomService.GetByCriteria(ctx, &pb.GetRoomsByCriteriaRequest{
		ReferenceId: "",
		AccountId: &pb.AccountIdRequest{
			AccountId:  pb.FromUUID(accountIdClient),
			ExternalId: "",
		},
		RoomId:          nil,
		WithClosed:      false,
		WithSubscribers: true,
	})
	if err != nil {
		t.Fatalf("Error: %v", err)
	}
	log.Println(getRs)

	// get by accountIdOperator
	getRs, err = roomService.GetByCriteria(ctx, &pb.GetRoomsByCriteriaRequest{
		ReferenceId: "",
		AccountId: &pb.AccountIdRequest{
			AccountId:  pb.FromUUID(accountIdOperator),
			ExternalId: "",
		},
		RoomId:          nil,
		WithClosed:      false,
		WithSubscribers: true,
	})
	if err != nil {
		t.Fatalf("Error: %v", err)
	}
	log.Println(getRs)

	// get by externalIdClient
	getRs, err = roomService.GetByCriteria(ctx, &pb.GetRoomsByCriteriaRequest{
		ReferenceId: "",
		AccountId: &pb.AccountIdRequest{
			AccountId:  nil,
			ExternalId: externalIdClient,
		},
		RoomId:          nil,
		WithClosed:      false,
		WithSubscribers: true,
	})
	if err != nil {
		t.Fatalf("Error: %v", err)
	}
	log.Println(getRs)

	// get by referenceId
	getRs, err = roomService.GetByCriteria(ctx, &pb.GetRoomsByCriteriaRequest{
		ReferenceId:     referenceId,
		AccountId:       nil,
		RoomId:          nil,
		WithClosed:      false,
		WithSubscribers: true,
	})
	if err != nil {
		t.Fatalf("Error: %v", err)
	}
	log.Println(getRs)
}

func TestCreateRoomWithOneSubscriberAndSubscribeThen_Success(t *testing.T) {

	conn, err := helper.GrpcConnection()
	if err != nil {
		t.Fatal(err.Error())
	}
	defer conn.Close()

	roomService := pb.NewRoomClient(conn)

	accountIdClient, _, err := helper.CreateDefaultAccount(conn)
	if err != nil {
		t.Fatal(err.Error())
	}

	accountIdOperator, externalIdOperator, err := helper.CreateDefaultAccount(conn)
	if err != nil {
		t.Fatal(err.Error())
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	referenceId := system.Uuid().String()
	r, err := roomService.Create(ctx, &pb.CreateRoomRequest{
		ReferenceId: referenceId,
		Chat:        true,
		Video:       false,
		Audio:       false,
		Subscribers: []*pb.SubscriberRequest{
			{
				Account: &pb.AccountIdRequest{
					AccountId: pb.FromUUID(accountIdClient),
				},
				Role: "client",
			},
		},
	})
	if err != nil {
		t.Fatalf("Error: %v", err)
	}
	log.Printf("Response: %v", r)

	_, err = roomService.Subscribe(ctx, &pb.RoomSubscribeRequest{
		RoomId:      r.Result.Id,
		ReferenceId: "",
		Subscribers: []*pb.SubscriberRequest{
			{
				Account: &pb.AccountIdRequest{
					ExternalId: externalIdOperator,
				},
				Role: "operator",
			},
		},
	})
	if err != nil {
		t.Fatalf("Error: %v", err)
	}

	// get by referenceId
	getRs, err := roomService.GetByCriteria(ctx, &pb.GetRoomsByCriteriaRequest{
		ReferenceId:     referenceId,
		AccountId:       nil,
		RoomId:          nil,
		WithClosed:      false,
		WithSubscribers: true,
	})
	if err != nil {
		t.Fatalf("Error: %v", err)
	}
	log.Println(getRs)

	// get by referenceId
	getRs, err = roomService.GetByCriteria(ctx, &pb.GetRoomsByCriteriaRequest{
		ReferenceId: "",
		AccountId: &pb.AccountIdRequest{
			AccountId:  pb.FromUUID(accountIdOperator),
			ExternalId: "",
		},
		RoomId:          nil,
		WithClosed:      false,
		WithSubscribers: true,
	})
	if err != nil {
		t.Fatalf("Error: %v", err)
	}
	log.Println(getRs)

	// get by referenceId
	getRs, err = roomService.GetByCriteria(ctx, &pb.GetRoomsByCriteriaRequest{
		ReferenceId: "",
		AccountId: &pb.AccountIdRequest{
			AccountId:  nil,
			ExternalId: externalIdOperator,
		},
		RoomId:          nil,
		WithClosed:      false,
		WithSubscribers: true,
	})
	if err != nil {
		t.Fatalf("Error: %v", err)
	}
	log.Println(getRs)

}

func TestFirstRoomClosedAfterSecondCreated_Success(t *testing.T) {

	sdkService, err := helper.InitSdk()
	if err != nil {
		t.Error(err.Error(), sdkService)
	}

	conn, err := helper.GrpcConnection()
	if err != nil {
		t.Fatal(err.Error())
	}
	defer conn.Close()

	roomService := pb.NewRoomClient(conn)

	accountIdClient, _, err := helper.CreateDefaultAccount(conn)
	if err != nil {
		t.Fatal(err.Error())
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	referenceId := system.Uuid().String()
	r1, err := roomService.Create(ctx, &pb.CreateRoomRequest{
		ReferenceId: referenceId,
		Chat:        true,
		Video:       false,
		Audio:       false,
		Subscribers: []*pb.SubscriberRequest{
			{
				Account: &pb.AccountIdRequest{
					AccountId: pb.FromUUID(accountIdClient),
				},
				Role: "client",
			},
		},
	})
	if err != nil {
		t.Fatalf("Error: %v", err)
	}
	log.Printf("Response: %v", r1)

	r2, err := roomService.Create(ctx, &pb.CreateRoomRequest{
		ReferenceId: referenceId,
		Chat:        true,
		Video:       false,
		Audio:       false,
		Subscribers: []*pb.SubscriberRequest{
			{
				Account: &pb.AccountIdRequest{
					AccountId: pb.FromUUID(accountIdClient),
				},
				Role: "client",
			},
		},
	})
	if err != nil {
		t.Fatalf("Error: %v", err)
	}
	log.Printf("Response: %v", r2)

	// get by referenceId
	getRs, err := roomService.GetByCriteria(ctx, &pb.GetRoomsByCriteriaRequest{
		ReferenceId: referenceId,
		AccountId: nil,
		RoomId:          nil,
		WithClosed:      true,
		WithSubscribers: true,
	})
	if err != nil {
		t.Fatalf("Error: %v", err)
	}
	log.Println(getRs)

}