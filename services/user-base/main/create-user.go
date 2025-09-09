package main

import (
	"context"
	"fmt"
	"strings"

	authpb "github.com/Costin2000/GoChat---Schwarz-Internship---2025/services/auth/proto"
	pb "github.com/Costin2000/GoChat---Schwarz-Internship---2025/services/user-base/proto"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
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

	// Salvam parola originala pentru login ulterior
	plainPassword := user.Password

	// Cream utilizatorul in DB
	createdUser, err := svc.storageAccess.createUser(ctx, req.User)
	if err != nil {
		return nil, err
	}

	// Obtinem token apeland serviciul Auth
	loginResp, err := svc.authClient.Login(ctx, &authpb.LoginRequest{
		Email:    user.Email,
		Password: plainPassword,
	})
	if err != nil {
		return nil, status.Errorf(codes.Internal, "user created but failed to login: %v", err)
	}

	// Trimitem si tokenul in raspuns
	if svc.emailPub != nil {
		_ = svc.emailPub.Publish(ctx, EmailMessage{
			To:      user.Email,
			Subject: "Welcome to GoChat",
			Body:    fmt.Sprintf("Hi %s, \n\nYour account was created successfully. Enjoy the experience!\n\n- GoChat Team", user.FirstName),
		})
	}

	return &pb.CreateUserResponse{
		User:  createdUser,
		Token: loginResp.Token,
	}, nil
}
