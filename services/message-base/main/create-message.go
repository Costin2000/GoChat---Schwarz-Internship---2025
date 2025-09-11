package main

import (
	"context"
	"strings"

	pb "github.com/Costin2000/GoChat---Schwarz-Internship---2025/services/message-base/proto"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const maxContentLen = 4096

func (svc *MessageService) CreateMessage(ctx context.Context, req *pb.CreateMessageRequest) (*pb.CreateMessageResponse, error) {
	m := req.GetMessage()
	if m == nil {
		return nil, status.Error(codes.InvalidArgument, "message is required")
	}
	if m.ConversationId <= 0 || m.SenderId <= 0 {
		return nil, status.Error(codes.InvalidArgument, "conversation_id, sender_id must be positive")
	}
	m.Content = strings.TrimSpace(m.Content)
	if m.Content == "" {
		return nil, status.Error(codes.InvalidArgument, "content cannot be empty")
	}
	if len(m.Content) > maxContentLen {
		return nil, status.Errorf(codes.InvalidArgument, "content too long (max %d chars)", maxContentLen)
	}

	created, err := svc.storageAccess.createMessage(ctx, m)
	if err != nil {
		return nil, err
	}

	return &pb.CreateMessageResponse{Message: created}, nil
}
