package main

import (
	"context"

	pb "github.com/Costin2000/GoChat---Schwarz-Internship---2025/services/friend-request-base/proto"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type mockStorage struct {
	createFriendRequestFunc func(ctx context.Context, req *pb.CreateFriendRequestRequest) (*pb.CreateFriendRequestResponse, error)
	listFriendRequestsFunc  func(ctx context.Context, req *pb.ListFriendRequestsRequest) (*pb.ListFriendRequestsResponse, error)
	updateFriendRequestFunc func(ctx context.Context, req *pb.UpdateFriendRequestRequest) (*pb.UpdateFriendRequestResponse, error)
}

func (m *mockStorage) requestCreateFriendRequest(ctx context.Context, req *pb.CreateFriendRequestRequest) (*pb.CreateFriendRequestResponse, error) {
	return m.createFriendRequestFunc(ctx, req)
}

func (m *mockStorage) listFriendRequests(ctx context.Context, req *pb.ListFriendRequestsRequest) (*pb.ListFriendRequestsResponse, error) {
	return m.listFriendRequestsFunc(ctx, req)
}

func (m *mockStorage) requestUpdateFriendRequest(ctx context.Context, req *pb.UpdateFriendRequestRequest) (*pb.UpdateFriendRequestResponse, error) {
	return m.updateFriendRequestFunc(ctx, req)
}

type StorageMockOptions struct {
	createFriendRequestFunc func(ctx context.Context, req *pb.CreateFriendRequestRequest) (*pb.CreateFriendRequestResponse, error)
	listFriendRequestsFunc  func(ctx context.Context, req *pb.ListFriendRequestsRequest) (*pb.ListFriendRequestsResponse, error)
	updateFriendRequestFunc func(ctx context.Context, req *pb.UpdateFriendRequestRequest) (*pb.UpdateFriendRequestResponse, error)
}

func newMockStorageAccess(opts StorageMockOptions) StorageAccess {

	createFriendRequestFunc := func(ctx context.Context, req *pb.CreateFriendRequestRequest) (*pb.CreateFriendRequestResponse, error) {
		return fixtureCreateFriendRequestResponse(), nil
	}
	listFriendRequestsFunc := func(ctx context.Context, req *pb.ListFriendRequestsRequest) (*pb.ListFriendRequestsResponse, error) {
		return &pb.ListFriendRequestsResponse{ // Implement this in the actual test
			NextPageToken: "",
			Requests:      []*pb.FriendRequest{},
		}, nil
	}

	if opts.createFriendRequestFunc != nil {
		createFriendRequestFunc = opts.createFriendRequestFunc
	}

	if opts.listFriendRequestsFunc != nil {
		listFriendRequestsFunc = opts.listFriendRequestsFunc
	}

	return &mockStorage{
		createFriendRequestFunc: createFriendRequestFunc,
		listFriendRequestsFunc:  listFriendRequestsFunc,
		updateFriendRequestFunc: opts.updateFriendRequestFunc,
	}
}

type ServiceMockOptions struct {
	storageAccess StorageAccess
}

func NewMockService(opts ServiceMockOptions) *friendRequestService {
	storage := newMockStorageAccess(StorageMockOptions{})
	if opts.storageAccess != nil {
		storage = opts.storageAccess
	}

	return &friendRequestService{
		storageAccess: storage,
	}
}

func fixtureCreateFriendRequestResponse(mods ...func(*pb.FriendRequest)) *pb.CreateFriendRequestResponse {

	friendReq := &pb.FriendRequest{
		Id:         "1",
		SenderId:   "111",
		ReceiverId: "222",
		Status:     pb.RequestStatus_STATUS_PENDING,
		CreatedAt:  timestamppb.Now(),
	}

	for _, mod := range mods {
		mod(friendReq)
	}

	return &pb.CreateFriendRequestResponse{
		Request: friendReq,
	}
}

func fixtureUpdateFriendResponse(mods ...func(*pb.UpdateFriendRequestResponse)) *pb.UpdateFriendRequestResponse {
	friendReq := &pb.FriendRequest{
		Id:         "1",
		SenderId:   "111",
		ReceiverId: "222",
		Status:     pb.RequestStatus_STATUS_ACCEPTED,
		CreatedAt:  timestamppb.Now(),
	}

	resp := &pb.UpdateFriendRequestResponse{
		FriendRequest: friendReq,
	}

	for _, mod := range mods {
		mod(resp)
	}

	return resp
}
