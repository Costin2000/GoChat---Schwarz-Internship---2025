package main

import (
	"context"
	"testing"

	errchecks "github.com/Costin2000/GoChat---Schwarz-Internship---2025/pkg"
	pb "github.com/Costin2000/GoChat---Schwarz-Internship---2025/services/friend-request-base/proto"
	"github.com/google/go-cmp/cmp"
	"google.golang.org/protobuf/testing/protocmp"
	"google.golang.org/protobuf/types/known/fieldmaskpb"
)

func fixtureUpdateFriendRequest(mods ...func(req *pb.UpdateFriendRequestRequest)) *pb.UpdateFriendRequestRequest {
	req := &pb.UpdateFriendRequestRequest{
		FriendRequest: &pb.FriendRequest{
			Id:         "1",
			SenderId:   "111",
			ReceiverId: "222",
			Status:     pb.RequestStatus_STATUS_ACCEPTED,
		},
		FieldMask: &fieldmaskpb.FieldMask{Paths: []string{"status"}},
	}

	for _, mod := range mods {
		mod(req)
	}

	return req
}

func Test_UpdateFriendRequest(t *testing.T) {
	type given struct {
		mockStorageAccess StorageAccess
	}

	successfulResponse := fixtureUpdateFriendResponse()

	tests := []struct {
		name         string
		req          *pb.UpdateFriendRequestRequest
		given        given
		expecterErr  errchecks.Check
		expectedResp *pb.UpdateFriendRequestResponse
	}{
		{
			name: "Nil friend request object",
			req: fixtureUpdateFriendRequest(func(req *pb.UpdateFriendRequestRequest) {
				req.FriendRequest = nil
			}),
			expecterErr: errchecks.MsgContains("friend request id must be provided"),
		},
		{
			name: "Empty friend request ID",
			req: fixtureUpdateFriendRequest(func(req *pb.UpdateFriendRequestRequest) {
				req.FriendRequest.Id = ""
			}),
			expecterErr: errchecks.MsgContains("friend request id must be provided"),
		},
		{
			name: "happy path - should update friend request successfully",
			req:  fixtureUpdateFriendRequest(),
			given: given{
				mockStorageAccess: newMockStorageAccess(StorageMockOptions{
					updateFriendRequestFunc: func(ctx context.Context, req *pb.UpdateFriendRequestRequest) (*pb.UpdateFriendRequestResponse, error) {
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

			rsp, err := svc.UpdateFriendRequest(context.Background(), tt.req)

			errchecks.Assert(t, err, tt.expecterErr)
			if diff := cmp.Diff(tt.expectedResp, rsp, protocmp.Transform()); diff != "" {
				t.Errorf("mismatch (-expected +got):\n%s", diff)
			}
		})
	}
}
