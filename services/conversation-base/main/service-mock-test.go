package main

import (
	"context"

	pb "github.com/Costin2000/GoChat---Schwarz-Internship---2025/services/conversation-base/proto"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type mockStorage struct {
	createConversationFunc func(ctx context.Context, req *pb.CreateConversationRequest) (*pb.CreateConversationResponse, error)
}

func (m *mockStorage) createConversation(ctx context.Context, req *pb.CreateConversationRequest) (*pb.CreateConversationResponse, error) {
	return m.createConversationFunc(ctx, req)
}

type StorageMockOptions struct {
	createConversationFunc func(ctx context.Context, req *pb.CreateConversationRequest) (*pb.CreateConversationResponse, error)
}

func newMockStorageAccess(opts StorageMockOptions) StorageAccess {
	createConversationFunc := func(ctx context.Context, req *pb.CreateConversationRequest) (*pb.CreateConversationResponse, error) {
		return fixtureCreateConversationResponse(), nil
	}

	if opts.createConversationFunc != nil {
		createConversationFunc = opts.createConversationFunc
	}

	return &mockStorage{
		createConversationFunc: createConversationFunc,
	}
}

type ServiceMockOptions struct {
	storageAccess StorageAccess
}

func NewMockService(opts ServiceMockOptions) *conversationService {
	storage := newMockStorageAccess(StorageMockOptions{})
	if opts.storageAccess != nil {
		storage = opts.storageAccess
	}

	return &conversationService{
		storageAccess: storage,
	}
}

func fixtureCreateConversationResponse(mods ...func(*pb.Conversation)) *pb.CreateConversationResponse {
	conv := &pb.Conversation{
		Id:        "1",
		User1Id:   "111",
		User2Id:   "222",
		CreatedAt: timestamppb.Now(),
		UpdatedAt: timestamppb.Now(),
	}

	for _, mod := range mods {
		mod(conv)
	}

	return &pb.CreateConversationResponse{
		Conversation: conv,
	}
}
