// services/auth/main/service.go
package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"

	proto "github.com/Costin2000/GoChat---Schwarz-Internship---2025/services/auth/proto"
	userbasepb "github.com/Costin2000/GoChat---Schwarz-Internship---2025/services/user-base/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type userBaseClient interface {
	GetUser(ctx context.Context, in *userbasepb.GetUserRequest, opts ...grpc.CallOption) (*userbasepb.User, error)
}

var jwtSecret = []byte(os.Getenv("AUTH_JWT_SECRET"))

type authServer struct {
	proto.UnimplementedAuthServiceServer
	userBaseClient userBaseClient
}

func (s *authServer) Ping(ctx context.Context, in *proto.Empty) (*proto.Pong, error) {
	return &proto.Pong{Message: "pong"}, nil
}

func main() {
	userBaseAddr := os.Getenv("USER_BASE_ADDR")
	if userBaseAddr == "" {
		userBaseAddr = "user-base:50051"
	}

	conn, err := grpc.NewClient(userBaseAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("did not connect to user-base service: %v", err)
	}
	defer conn.Close()

	userBaseClient := userbasepb.NewUserServiceClient(conn)

	lis, err := net.Listen("tcp", ":50053")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	grpcServer := grpc.NewServer()
	proto.RegisterAuthServiceServer(grpcServer, &authServer{
		userBaseClient: userBaseClient,
	})

	fmt.Println("Auth gRPC server listening on :50053...")

	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
