package main

import (
	"context"
	"log"
	"net"

	proto "github.com/Costin2000/GoChat---Schwarz-Internship---2025/services/user-base/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type userBaseServer struct {
	proto.UnimplementedUserBaseServer
}

func (s *userBaseServer) CreateUser(ctx context.Context, req *proto.CreateUserRequest) (*proto.CreateUserResponse, error) {

	log.Printf("Create user request for \"%s\" received.", req.GetUserName())

	newUser := &proto.User{
		Id:        "someID",
		FirstName: req.GetFirstName(),
		LastName:  req.GetLastName(),
		UserName:  req.GetUserName(),
		Email:     req.GetEmail(),
		CreatedAt: timestamppb.Now(),
	}

	return &proto.CreateUserResponse{User: newUser}, nil
}

func (s *userBaseServer) Ping(ctx context.Context, req *proto.Empty) (*proto.PingResponse, error) {

	return &proto.PingResponse{Response: "Pong"}, nil
}

func main() {

	port := ":50051"
	lis, err := net.Listen("tcp", port)
	if err != nil {
		log.Fatalf("Error polling port %s.: %v", port[1:], err)
	}

	UserBaseServer := grpc.NewServer()
	proto.RegisterUserBaseServer(UserBaseServer, &userBaseServer{})
	reflection.Register(UserBaseServer)

	log.Printf("gRPC polling port %s...", port[1:])

	if err := UserBaseServer.Serve(lis); err != nil {
		log.Fatalf("Error starting server: %v", err)
	}
}
