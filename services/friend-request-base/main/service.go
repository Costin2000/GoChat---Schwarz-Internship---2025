package main

import (
    "context"
    "fmt"
    "log"
    "net"

    proto "github.com/Costin2000/GoChat---Schwarz-Internship---2025/services/friend-request-base/proto"
    "google.golang.org/grpc"
)

type friendRequestServer struct {
    proto.UnimplementedFriendRequestServiceServer
}

func (s *friendRequestServer) Ping(ctx context.Context, in *proto.Empty) (*proto.Pong, error) {
    return &proto.Pong{Message: "pong from friend-request-base"}, nil
}

func main() {
    lis, err := net.Listen("tcp", ":50052")
    if err != nil {
        log.Fatalf("failed to listen: %v", err)
    }

    grpcServer := grpc.NewServer()
    proto.RegisterFriendRequestServiceServer(grpcServer, &friendRequestServer{})

    fmt.Println("Friend Request gRPC server listening on :50052...")

    if err := grpcServer.Serve(lis); err != nil {
        log.Fatalf("failed to serve: %v", err)
    }
}