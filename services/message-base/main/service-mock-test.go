package main

import (
	"context"

	pb "github.com/Costin2000/GoChat---Schwarz-Internship---2025/services/message-base/proto"
)

type mockStorage struct {
	createMessageFunc func(ctx context.Context, m *pb.Message) (*pb.Message, error)
	listMessagesFunc  func(ctx context.Context, req *pb.ListMessagesRequest) (*pb.ListMessagesResponse, error)
}

func (m *mockStorage) createMessage(ctx context.Context, msg *pb.Message) (*pb.Message, error) {
	if m.createMessageFunc != nil {
		return m.createMessageFunc(ctx, msg)
	}
	return nil, nil
}

func (m *mockStorage) listMessages(ctx context.Context, req *pb.ListMessagesRequest) (*pb.ListMessagesResponse, error) {
	if m.listMessagesFunc != nil {
		return m.listMessagesFunc(ctx, req)
	}
	return nil, nil
}

type StorageMockOptions struct {
	CreateMessageFunc func(ctx context.Context, m *pb.Message) (*pb.Message, error)
	ListMessagesFunc  func(ctx context.Context, req *pb.ListMessagesRequest) (*pb.ListMessagesResponse, error)
}

func newMockStorageAccess(opts StorageMockOptions) StorageAccess {

	mock := &mockStorage{
		createMessageFunc: opts.CreateMessageFunc,
		listMessagesFunc:  opts.ListMessagesFunc,
	}
	return mock
}
