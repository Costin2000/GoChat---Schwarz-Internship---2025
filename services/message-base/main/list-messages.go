package main

import (
	"context"

	pb "github.com/Costin2000/GoChat---Schwarz-Internship---2025/services/message-base/proto"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (svc *MessageService) ListMessages(ctx context.Context, req *pb.ListMessagesRequest) (*pb.ListMessagesResponse, error) {
	if req.Filter == nil || req.Filter.ConversationId < 1 {
		return nil, status.Errorf(codes.InvalidArgument, "Invalid Conversation ID")
	}

	return svc.storageAccess.listMessages(ctx, req)
}
