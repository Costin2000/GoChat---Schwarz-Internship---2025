package main

import (
	"context"

	proto "github.com/Costin2000/GoChat---Schwarz-Internship---2025/services/friend-request-base/proto"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (svc *friendRequestService) UpdateFriendRequest(ctx context.Context, req *proto.UpdateFriendRequestRequest) (*proto.UpdateFriendRequestResponse, error) {
	if req.FriendRequest == nil || req.FriendRequest.Id == "" {
		return nil, status.Errorf(codes.InvalidArgument, "friend request id must be provided")
	}

	if req.FieldMask == nil || len(req.FieldMask.Paths) == 0 {
		return nil, status.Errorf(codes.InvalidArgument, "at least one field must be specified in field mask")
	}

	return svc.storageAccess.requestUpdateFriendRequest(ctx, req)
}
