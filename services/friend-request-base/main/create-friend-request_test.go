package main

import (
	"context"
	"testing"

	errchecks "github.com/Costin2000/GoChat---Schwarz-Internship---2025/pkg"
	pb "github.com/Costin2000/GoChat---Schwarz-Internship---2025/services/friend-request-base/proto"
	"github.com/google/go-cmp/cmp"
	"google.golang.org/protobuf/testing/protocmp"
)

func fixtureCreateFriendRequest(mods ...func(req *pb.CreateFriendRequestRequest)) *pb.CreateFriendRequestRequest {
	friendReqReq := &pb.CreateFriendRequestRequest{
		SenderId:   "111",
		ReceiverId: "222",
	}

	for _, mod := range mods {
		mod(friendReqReq)
	}

	return friendReqReq
}

func fixtureCreateFriendResponse(mods ...func(req *pb.CreateFriendRequestResponse)) *pb.CreateFriendRequestResponse {
	friendReqRsp := &pb.CreateFriendRequestResponse{
		Request: fixtureFriendRequest().Request,
	}

	for _, mod := range mods {
		mod(friendReqRsp)
	}

	return friendReqRsp
}

func Test_CreateFriendRequestRequest(t *testing.T) {

	type given struct {
		mockStorageAccess StorageAccess
	}

	successfulResponse := fixtureCreateFriendResponse()

	tests := []struct {
		name         string
		req          *pb.CreateFriendRequestRequest
		given        given
		expecterErr  errchecks.Check
		expectedResp *pb.CreateFriendRequestResponse
	}{
		{
			name: "Empty sender id",
			req: fixtureCreateFriendRequest(func(req *pb.CreateFriendRequestRequest) {
				req.SenderId = ""
			}),
			expecterErr: errchecks.MsgContains("sender and receiver IDs cannot be empty"),
		},
		{
			name: "Empty receiver id",
			req: fixtureCreateFriendRequest(func(req *pb.CreateFriendRequestRequest) {
				req.ReceiverId = ""
			}),
			expecterErr: errchecks.MsgContains("sender and receiver IDs cannot be empty"),
		},
		{
			name: "Sender id same as receiver id",
			req: fixtureCreateFriendRequest(func(req *pb.CreateFriendRequestRequest) {
				req.ReceiverId = req.SenderId
			}),
			expecterErr:  errchecks.MsgContains("sender and receiver cannot be the same user"),
			expectedResp: nil,
		},
		{
			name: "happy path - should create friend request successfully",
			req:  fixtureCreateFriendRequest(), // Standard valid request
			given: given{
				mockStorageAccess: newMockStorageAccess(StorageMockOptions{
					// Configure the mock to return the successful response
					createFriendRequestFunc: func(ctx context.Context, req *pb.CreateFriendRequestRequest) (*pb.CreateFriendRequestResponse, error) {
						return successfulResponse, nil
					},
				}),
			},
			expecterErr:  nil,
			expectedResp: successfulResponse,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := NewMockService(ServiceMockOptions{
				storageAccess: tt.given.mockStorageAccess,
			})

			rsp, err := svc.CreateFriendRequest(context.Background(), tt.req)

			errchecks.Assert(t, err, tt.expecterErr)
			if diff := cmp.Diff(tt.expectedResp, rsp, protocmp.Transform()); diff != "" {
				t.Errorf("mismatch (-expected +got):\n%s", diff)
			}
		})
	}

}
