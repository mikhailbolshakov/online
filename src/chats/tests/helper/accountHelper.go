package helper

import (
	pb "chats/proto"
	"chats/system"
	"context"
	"errors"
	uuid "github.com/satori/go.uuid"
	"google.golang.org/grpc"
	"log"
)

func defaultAccount() *pb.CreatAccountRequest {
	return &pb.CreatAccountRequest{
		Account:    "testAccount",
		Type:       "user",
		ExternalId: system.Uuid().String(),
		FirstName:  "Иванов",
		MiddleName: "Иванович",
		LastName:   "Иванов",
		Email:      "ivanov@gmail.com",
		Phone:      "+79107895632",
		AvatarUrl:  "https://s3.adacta.ru/ivanov",
	}
}

func CreateDefaultAccount(conn *grpc.ClientConn) (id uuid.UUID, externalId string, err error) {

	accountService := pb.NewAccountClient(conn)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	accountRq := defaultAccount()

	rs, err := accountService.Create(ctx, accountRq)
	if err != nil {
		return uuid.Nil, "", err
	}

	log.Printf("Account created. id: %s, externalId: %s \n", rs.Account.Id.Value, externalId)

	return rs.Account.Id.ToUUID(), accountRq.ExternalId, nil

}

func CreateBotAccount(conn *grpc.ClientConn) (id uuid.UUID, externalId string, err error) {

	accountService := pb.NewAccountClient(conn)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	accountRq := 	&pb.CreatAccountRequest{
		Account:    "bot",
		Type:       "bot",
		ExternalId: system.Uuid().String(),
	}

	rs, err := accountService.Create(ctx, accountRq)
	if err != nil {
		return uuid.Nil, "", err
	}

	log.Printf("Bot account created. id: %s, externalId: %s \n", rs.Account.Id.Value, externalId)

	return rs.Account.Id.ToUUID(), accountRq.ExternalId, nil

}


func UpdateAccount(conn *grpc.ClientConn, rq *pb.UpdateAccountRequest) error {

	accountService := pb.NewAccountClient(conn)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	rs, err := accountService.Update(ctx, rq)
	if err != nil {
		return err
	}

	if len(rs.Errors) > 0 {
		for _, e := range rs.Errors{
			log.Printf("Error: %d %s \n", e.Code, e.Message)
		}
		return errors.New("errors")
	}

	log.Println("Account updated. request: % \n", rq)

	return nil

}

func GetAccountsByCriteria(conn *grpc.ClientConn, rq *pb.GetAccountsByCriteriaRequest) ([]*pb.AccountItem, error) {

	accountService := pb.NewAccountClient(conn)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	rs, err := accountService.GetByCriteria(ctx, rq)
	if err != nil {
		return nil, err
	}

	if len(rs.Errors) > 0 {
		for _, e := range rs.Errors{
			log.Printf("Error: %d %s \n", e.Code, e.Message)
		}
		return nil, errors.New("errors")
	}

	log.Printf("Found %d accounts. \n Criteria: %v", len(rs.Accounts), rq)
	if len(rs.Accounts) > 0 {
		for _, a := range rs.Accounts{
			log.Printf("%v \n", a)
		}
		return rs.Accounts, nil
	}

	return nil, nil
}

func GetAccountById(conn *grpc.ClientConn, id uuid.UUID) (*pb.AccountItem, error) {

	r, err := GetAccountsByCriteria(conn, &pb.GetAccountsByCriteriaRequest{
		AccountId: &pb.AccountIdRequest{
			AccountId: pb.FromUUID(id),
		},
	})
	if err != nil {
		return nil, err
	}

	if len(r) > 0 {
		return r[0], nil
	} else {
		return nil, nil
	}

}

func SetAccountOnlineStatus(conn *grpc.ClientConn, accountId uuid.UUID, status string)  error {

	accountService := pb.NewAccountClient(conn)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	rs, err := accountService.SetOnlineStatus(ctx, &pb.SetOnlineStatusRequest{
		Status:    status,
		AccountId: &pb.AccountIdRequest{AccountId: pb.FromUUID(accountId)},
	})
	if err != nil {
		return err
	}

	if len(rs.Errors) > 0 {
		for _, e := range rs.Errors {
			log.Printf("Error: %d %s \n", e.Code, e.Message)
		}
		return errors.New("errors")
	}

	log.Printf("Online status changed. AccountId: %s, status: %s \n", accountId.String(), status)

	return nil
}

func GetAccountOnlineStatus(conn *grpc.ClientConn, accountId uuid.UUID) (string, error) {

	accountService := pb.NewAccountClient(conn)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	rs, err := accountService.GetOnlineStatus(ctx, &pb.GetOnlineStatusRequest{
		AccountId: &pb.AccountIdRequest{AccountId: pb.FromUUID(accountId)},
	})
	if err != nil {
		return "", err
	}

	if len(rs.Errors) > 0 {
		for _, e := range rs.Errors {
			log.Printf("Error: %d %s \n", e.Code, e.Message)
		}
		return "", errors.New("errors")
	}

	log.Printf("Online status for %s is %s \n", accountId.String(), rs.Status)

	return rs.Status, nil
}