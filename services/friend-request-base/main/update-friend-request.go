package main

import (
	"context"
	"fmt"
	"log"

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

	// actualizeaza in DB
	resp, err := svc.storageAccess.requestUpdateFriendRequest(ctx, req)
	if err != nil {
		return nil, err
	}

	// daca statusul este ACCEPTED, trimitem email catre SENDER
	if resp != nil && resp.FriendRequest != nil && resp.FriendRequest.Status == proto.RequestStatus_STATUS_ACCEPTED {
		log.Printf("A intrat in conditia care trebuia, nice")

		// verificam ca avem publisher si clientul user-base
		if svc.emailPub != nil && svc.userClient != nil {
			senderID := resp.FriendRequest.SenderId
			receiverID := resp.FriendRequest.ReceiverId

			// obtinem informatiile sender
			sender, sErr := svc.getUserByID(ctx, senderID)
			if sErr != nil {
				log.Printf("WARN: could not fetch sender user %s: %v", senderID, sErr)
			}

			// obtinem informatiile receiver (cel care a acceptat) ca sa includem numele in mail
			receiver, rErr := svc.getUserByID(ctx, receiverID)
			if rErr != nil {
				log.Printf("WARN: could not fetch receiver user %s: %v", receiverID, rErr)
			}

			// construim corpul email-ului; folosim fallback-uri daca nu avem date
			senderEmail := ""
			if sender != nil {
				senderEmail = sender.Email
			}
			receiverName := ""
			receiverUsername := ""
			if receiver != nil {
				receiverName = fmt.Sprintf("%s %s", receiver.FirstName, receiver.LastName)
				receiverUsername = receiver.UserName
			}

			if senderEmail == "" {
				log.Printf("WARN: sender %s has no email, skipping email notification", senderID)
			} else {
				subject := "Friend request accepted on GoChat"
				body := fmt.Sprintf("Hi %s,\n\nGood news! Your friend request to %s (@%s) was accepted.\n\nYou can now start chatting!\n\n- GoChat Team",
					// preferam numele senderului in salut; daca nu exista, folosim email-ul
					func() string {
						if sender != nil && sender.FirstName != "" {
							return sender.FirstName
						}
						return senderEmail
					}(),
					receiverName, receiverUsername,
				)

				if pubErr := svc.emailPub.Publish(ctx, EmailMessage{
					To:      senderEmail,
					Subject: subject,
					Body:    body,
				}); pubErr != nil {
					log.Printf("WARN: failed to publish accepted-friend-request email to %s: %v", senderEmail, pubErr)
				} else {
					log.Printf("Info: published accepted-friend-request email to %s", senderEmail)
				}
			}
		} else {
			// lipseste emailPub sau userClient — logam pentru debugging
			if svc.emailPub == nil {
				log.Printf("WARN: email publisher not configured; cannot send accepted-friend-request email")
			}
			if svc.userClient == nil {
				log.Printf("WARN: user service client not configured; cannot fetch sender email to notify")
			}
		}
	}

	return resp, nil
}
