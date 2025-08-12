package main

import (
	"context"
	"fmt"
	"log"
	"net"

	proto "github.com/Costin2000/GoChat---Schwarz-Internship---2025/services/auth/proto"
	"google.golang.org/grpc"
)

// Implements the AuthService gRPC server
type authServer struct {
	proto.UnimplementedAuthServiceServer
}

// RPC method that returns a "pong" message to test the connectivity
func (s *authServer) Ping(ctx context.Context, in *proto.Empty) (*proto.Pong, error) {
	return &proto.Pong{Message: "pong "}, nil
}

func main() {
	lis, err := net.Listen("tcp", ":50053")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	// Create new gRPC server
	grpcServer := grpc.NewServer()
	proto.RegisterAuthServiceServer(grpcServer, &authServer{})

	fmt.Println("Auth gRPC server listening on :50053...")

	//  Waiting for incoming gRPC requests
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
