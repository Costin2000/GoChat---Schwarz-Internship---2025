package main

import (
	"context"
	pb "github.com/Costin2000/GoChat---Schwarz-Internship---2025/services/user-base/proto"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"log"
)

func (svc *UserService) GetUser(ctx context.Context, req *pb.GetUserRequest) (*pb.User, error) {
	log.Printf("GetUser request for email: %s", req.Email)
	email := req.GetEmail()
	if email == "" {
		return nil, status.Errorf(codes.InvalidArgument, "email cannot be empty")
	}

	user, err := svc.storageAccess.getUserByEmail(ctx, req.Email)
	if err != nil {
		return nil, err
	}

	return user, nil
}
