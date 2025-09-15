package main

import (
	"context"

	pb "github.com/Costin2000/GoChat---Schwarz-Internship---2025/services/conversation-base/proto"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type mockStorage struct {
	createConversationFunc func(ctx context.Context, req *pb.CreateConversationRequest) (*pb.CreateConversationResponse, error)
	listConversationsFunc  func(ctx context.Context, req *pb.ListConversationsRequest) (*pb.ListConversationsResponse, error)
}

func (m *mockStorage) createConversation(ctx context.Context, req *pb.CreateConversationRequest) (*pb.CreateConversationResponse, error) {
	return m.createConversationFunc(ctx, req)
}

func (m *mockStorage) listConversations(ctx context.Context, req *pb.ListConversationsRequest) (*pb.ListConversationsResponse, error) {
	return m.listConversationsFunc(ctx, req)
}

type StorageMockOptions struct {
	createConversationFunc func(ctx context.Context, req *pb.CreateConversationRequest) (*pb.CreateConversationResponse, error)
	listConversationsFunc  func(ctx context.Context, req *pb.ListConversationsRequest) (*pb.ListConversationsResponse, error)
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
		listConversationsFunc:  opts.listConversationsFunc,
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
