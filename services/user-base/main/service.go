package main

import (
	"log"
	"net"

	proto "github.com/Costin2000/GoChat---Schwarz-Internship---2025/services/user-base/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

const port = ":50051"

type userBaseServer struct {
	proto.UnimplementedUserBaseServer
}

func main() {

	lis, err := net.Listen("tcp", port)
	if err != nil {
		log.Fatalf("Error polling port %s %v", port, err)
	}

	UserBaseServer := grpc.NewServer()
	proto.RegisterUserBaseServer(UserBaseServer, &userBaseServer{})
	reflection.Register(UserBaseServer)

	log.Printf("gRPC polling port %s...", port)

	if err := UserBaseServer.Serve(lis); err != nil {
		log.Fatalf("Error starting server: %v", err)
	}
}
