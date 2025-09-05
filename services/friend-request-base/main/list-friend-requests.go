package main

import (
	"context"

	proto "github.com/Costin2000/GoChat---Schwarz-Internship---2025/services/friend-request-base/proto"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const (
	defaultPageSize = int64(10)
	maxPageSize     = 1000
)

func (svc *friendRequestService) ListFriendRequests(ctx context.Context, req *proto.ListFriendRequestsRequest) (*proto.ListFriendRequestsResponse, error) {

	if req.PageSize <= 0 {
		return nil, status.Error(codes.InvalidArgument, "page size must be a positive number")
	}
	if req.PageSize < defaultPageSize {
		return nil, status.Errorf(codes.InvalidArgument, "page size cannot be less than %d", defaultPageSize)
	}
	if req.PageSize > maxPageSize {
		req.PageSize = maxPageSize
	}

	return svc.storageAccess.listFriendRequests(ctx, req)
}
