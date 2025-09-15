package main

import (
	"context"
	"log"

	proto "github.com/Costin2000/GoChat---Schwarz-Internship---2025/services/conversation-base/proto"
)

func (svc *conversationService) ListConversations(ctx context.Context, req *proto.ListConversationsRequest) (*proto.ListConversationsResponse, error) {
	log.Printf("Listing conversations (filter user_id=%s)", req.UserId)
	return svc.storageAccess.listConversations(ctx, req)
}
