package main

import (
	"context"

	frpb "github.com/Costin2000/GoChat---Schwarz-Internship---2025/services/friend-request-base/proto"
	userpb "github.com/Costin2000/GoChat---Schwarz-Internship---2025/services/user-base/proto"
	"google.golang.org/grpc"
)

type ServiceMockOptions struct {
	userClient UserClient
	frClient   FriendRequestClient
}

type userClientMock struct {
	capturedListUsersReq *userpb.ListUsersRequest
	ListUsersFunc        func(ctx context.Context, req *userpb.ListUsersRequest, opts ...grpc.CallOption) (*userpb.ListUsersResponse, error)
}
type frClientMock struct {
	capturedFriendRequestsReq *frpb.ListFriendRequestsRequest
	ListFrFunc                func(ctx context.Context, req *frpb.ListFriendRequestsRequest, opts ...grpc.CallOption) (*frpb.ListFriendRequestsResponse, error)
}

func (client *userClientMock) ListUsers(ctx context.Context, req *userpb.ListUsersRequest, opts ...grpc.CallOption) (*userpb.ListUsersResponse, error) {
	client.capturedListUsersReq = req
	return client.ListUsersFunc(ctx, req)
}

func (client *frClientMock) ListFriendRequests(ctx context.Context, req *frpb.ListFriendRequestsRequest, opts ...grpc.CallOption) (*frpb.ListFriendRequestsResponse, error) {
	client.capturedFriendRequestsReq = req
	return client.ListFrFunc(ctx, req)
}

func NewMockService(opts ServiceMockOptions) *AggregatorService {

	service := &AggregatorService{}

	if opts.userClient == nil {
		service.userBaseClient = &userClientMock{
			ListUsersFunc: func(ctx context.Context, req *userpb.ListUsersRequest, opts ...grpc.CallOption) (*userpb.ListUsersResponse, error) {
				return &userpb.ListUsersResponse{}, nil
			},
		}
	} else {
		service.userBaseClient = opts.userClient
	}

	if opts.frClient == nil {
		service.frClient = &frClientMock{
			ListFrFunc: func(ctx context.Context, req *frpb.ListFriendRequestsRequest, opts ...grpc.CallOption) (*frpb.ListFriendRequestsResponse, error) {
				return &frpb.ListFriendRequestsResponse{}, nil
			},
		}
	} else {
		service.frClient = opts.frClient
	}

	return service
}
