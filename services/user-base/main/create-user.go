package main

import (
	"context"
	pb "github.com/Costin2000/GoChat---Schwarz-Internship---2025/services/user-base/proto"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"strings"
)

func (svc *UserService) CreateUser(ctx context.Context, req *pb.CreateUserRequest) (*pb.CreateUserResponse, error) {
	user := req.GetUser()
	if user == nil {
		return nil, status.Errorf(codes.InvalidArgument, "user object is required")
	}

	// Basic validation
	if strings.TrimSpace(user.FirstName) == "" ||
		strings.TrimSpace(user.LastName) == "" ||
		strings.TrimSpace(user.UserName) == "" ||
		strings.TrimSpace(user.Email) == "" ||
		strings.TrimSpace(user.Password) == "" {
		return nil, status.Errorf(codes.InvalidArgument, "all fields are required")
	}

	user, err := svc.storageAccess.createUser(ctx, req.User)
	if err != nil {
		return nil, err
	}

	return &pb.CreateUserResponse{
		User: user,
	}, nil
}
