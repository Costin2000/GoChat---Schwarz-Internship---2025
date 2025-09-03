package main

import (
	"context"
	"log"

	proto "github.com/Costin2000/GoChat---Schwarz-Internship---2025/services/friend-request-base/proto"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (svc *friendRequestService) CreateFriendRequest(ctx context.Context, req *proto.CreateFriendRequestRequest) (*proto.CreateFriendRequestResponse, error) {

	senderID := req.SenderId
	receiverID := req.ReceiverId

	if senderID == "" || receiverID == "" {
		return nil, status.Errorf(codes.InvalidArgument, "sender and receiver IDs cannot be empty")
	}

	log.Printf("Create Friend Request from user: %s to user: %s", req.SenderId, req.ReceiverId)

	if senderID == receiverID {
		return nil, status.Errorf(codes.InvalidArgument, "sender and receiver cannot be the same user")
	}

	friendRequestResp, err := svc.storageAccess.requestCreateFriendRequest(ctx, req)

	if err != nil {
		return nil, err
	}

	return friendRequestResp, nil
}

func (svc *friendRequestService) UpdateFriendRequest(ctx context.Context, req *proto.UpdateFriendRequestRequest) (*proto.UpdateFriendRequestResponse, error) {
	if req.FriendRequest == nil || req.FriendRequest.Id == "" {
		return nil, status.Errorf(codes.InvalidArgument, "friend request id must be provided")
	}
	return svc.storageAccess.requestUpdateFriendRequest(ctx, req)
}
