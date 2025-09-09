package main

import (
	"context"
	"log"
	"net"

	aggrpb "github.com/Costin2000/GoChat---Schwarz-Internship---2025/services/aggregator/proto"
	frpb "github.com/Costin2000/GoChat---Schwarz-Internship---2025/services/friend-request-base/proto"
	userpb "github.com/Costin2000/GoChat---Schwarz-Internship---2025/services/user-base/proto"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/reflection"
)

const (
	aggrPort = ":50054"
	frAddr   = "localhost:50052"
	userAddr = "localhost:50051"
)

type FriendRequestClient interface {
	ListFriendRequests(ctx context.Context, req *frpb.ListFriendRequestsRequest, opts ...grpc.CallOption) (*frpb.ListFriendRequestsResponse, error)
}
type UserClient interface {
	ListUsers(ctx context.Context, req *userpb.ListUsersRequest, opts ...grpc.CallOption) (*userpb.ListUsersResponse, error)
}

type AggregatorService struct {
	aggrpb.UnimplementedAggregatorServiceServer
	frClient       FriendRequestClient
	userBaseClient UserClient
}

func main() {

	// network connection
	lis, err := net.Listen("tcp", aggrPort)
	if err != nil {
		log.Fatalf("Error polling aggrPort %s %v", aggrPort, err)
	}

	// user and friend-request client connections
	userConn, err := grpc.NewClient(userAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	check(err, "dial user-base")
	defer userConn.Close()

	frConn, err := grpc.NewClient(frAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	check(err, "dial friend request service")
	defer frConn.Close()
	frClient := frpb.NewFriendRequestServiceClient(frConn)
	userBaseClient := userpb.NewUserServiceClient(userConn)

	aggrSvc := &AggregatorService{
		frClient:       frClient,
		userBaseClient: userBaseClient,
	}

	grpcServer := grpc.NewServer()

	aggrpb.RegisterAggregatorServiceServer(grpcServer, aggrSvc)
	reflection.Register(grpcServer)

	log.Printf("gRPC polling on aggrPort %s...", aggrPort)

	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("Error starting server: %v", err)
	}
}

func check(err error, msg string) {
	if err != nil {
		log.Fatalf("%s: %v", msg, err)
	}
}

func (s *AggregatorService) ListFriendRequests(ctx context.Context, req *frpb.ListFriendRequestsRequest) (*frpb.ListFriendRequestsResponse, error) {
	return s.frClient.ListFriendRequests(ctx, req)
}

func (s *AggregatorService) ListUsers(ctx context.Context, req *userpb.ListUsersRequest) (*userpb.ListUsersResponse, error) {
	return s.userBaseClient.ListUsers(ctx, req)
}
