package main

import (
	"context"
	"log"

	proto "github.com/Costin2000/GoChat---Schwarz-Internship---2025/services/user-base/proto"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func (s *userBaseServer) CreateUser(ctx context.Context, req *proto.CreateUserRequest) (*proto.CreateUserResponse, error) {

	log.Printf("Create user request for \"%s\" received.", req.User.UserName)

	newUser := &proto.User{
		Id:        "someID",
		FirstName: req.User.FirstName,
		LastName:  req.User.LastName,
		UserName:  req.User.UserName,
		Email:     req.User.Email,
		CreatedAt: timestamppb.Now(),
	}

	return &proto.CreateUserResponse{User: newUser}, nil
}

func (s *userBaseServer) Ping(ctx context.Context, req *proto.Empty) (*proto.PingResponse, error) {

	return &proto.PingResponse{Response: "Pong"}, nil
}
