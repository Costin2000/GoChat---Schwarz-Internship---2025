package main

import (
	"context"

	proto "github.com/Costin2000/GoChat---Schwarz-Internship---2025/services/friend-request-base/proto"
)

const (
	defaultPageSize = int64(10)
	maxPageSize     = 50
)

func (svc *friendRequestService) ListFriendRequests(ctx context.Context, req *proto.ListFriendRequestsRequest) (*proto.ListFriendRequestsResponse, error) {

	if req.PageSize < defaultPageSize {
		req.PageSize = defaultPageSize
	}
	if req.PageSize > maxPageSize {
		req.PageSize = maxPageSize
	}

	return svc.storageAccess.listFriendRequests(ctx, req)
}
