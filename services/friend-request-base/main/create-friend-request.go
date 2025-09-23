package main

import (
	"context"
	"fmt"
	"log"
	"strconv"

	proto "github.com/Costin2000/GoChat---Schwarz-Internship---2025/services/friend-request-base/proto"
	pbuser "github.com/Costin2000/GoChat---Schwarz-Internship---2025/services/user-base/proto"
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

	// Trimitem email doar daca avem publisher si clientul user-base
	if svc.emailPub != nil && svc.userClient != nil {
		receiver, err := svc.getUserByID(ctx, receiverID)
		if err != nil {
			log.Printf("WARN: could not fetch receiver user %s: %v", receiverID, err)
		} else {
			sender, sErr := svc.getUserByID(ctx, senderID)
			if sErr != nil {
				log.Printf("WARN: could not fetch sender user %s: %v", senderID, sErr)
			}

			senderName := ""
			senderUsername := ""
			if sender != nil {
				senderName = fmt.Sprintf("%s %s", sender.FirstName, sender.LastName)
				senderUsername = sender.UserName
			}

			subject := "You received a friend request!"
			body := fmt.Sprintf("Hi %s,\n\nYou have a new friend request from %s (@%s).\n\n- GoChat Team",
				receiver.FirstName, senderName, senderUsername,
			)

			if pubErr := svc.emailPub.Publish(ctx, EmailMessage{
				To:      receiver.Email,
				Subject: subject,
				Body:    body,
			}); pubErr != nil {
				log.Printf("WARN: failed to publish email for friend request to %s: %v", receiver.Email, pubErr)
			}
		}
	}

	return friendRequestResp, nil
}

// helper: obtine user din user-base folosind ListUsers + filter by id
func (svc *friendRequestService) getUserByID(ctx context.Context, idStr string) (*pbuser.User, error) {
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		return nil, fmt.Errorf("invalid user id %s: %w", idStr, err)
	}

	resp, err := svc.userClient.ListUsers(ctx, &pbuser.ListUsersRequest{
		PageSize: 1,
		Filters: []*pbuser.ListUsersFiltersOneOf{
			{
				Filter: &pbuser.ListUsersFiltersOneOf_UserIds{
					UserIds: &pbuser.FilterByIdIn{
						UserId: []int64{id},
					},
				},
			},
		},
	})
	if err != nil {
		return nil, err
	}
	if len(resp.Users) == 0 {
		return nil, status.Errorf(codes.NotFound, "user with id %s not found", idStr)
	}
	return resp.Users[0], nil
}
