package main

import (
	"context"
	"log"

	proto "github.com/Costin2000/GoChat---Schwarz-Internship---2025/services/conversation-base/proto"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (svc *conversationService) CreateConversation(ctx context.Context, req *proto.CreateConversationRequest) (*proto.CreateConversationResponse, error) {
	user1ID := req.User1Id
	user2ID := req.User2Id

	if user1ID == "" || user2ID == "" {
		return nil, status.Errorf(codes.InvalidArgument, "user1 and user2 IDs cannot be empty")
	}

	if user1ID == user2ID {
		return nil, status.Errorf(codes.InvalidArgument, "cannot create conversation with the same user")
	}

	log.Printf("Creating conversation between user %s and user %s", user1ID, user2ID)

	resp, err := svc.storageAccess.createConversation(ctx, req)
	if err != nil {
		return nil, err
	}

	return resp, nil
}
