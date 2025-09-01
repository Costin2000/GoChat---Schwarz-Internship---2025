package main

import (
	"context"
	"testing"

	errchecks "github.com/Costin2000/GoChat---Schwarz-Internship---2025/pkg"
	pb "github.com/Costin2000/GoChat---Schwarz-Internship---2025/services/friend-request-base/proto"
	"github.com/google/go-cmp/cmp"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/testing/protocmp"
	"google.golang.org/protobuf/types/known/fieldmaskpb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func fixtureUpdateFriendRequest(mods ...func(req *pb.UpdateFriendRequestRequest)) *pb.UpdateFriendRequestRequest {
	req := &pb.UpdateFriendRequestRequest{
		FriendRequest: &pb.FriendRequest{
			Id:     "1",
			Status: pb.RequestStatus_STATUS_ACCEPTED,
		},
		FieldMask: &fieldmaskpb.FieldMask{
			Paths: []string{"status"},
		},
	}

	for _, mod := range mods {
		mod(req)
	}

	return req
}

func fixtureUpdateFriendResponse(mods ...func(resp *pb.UpdateFriendRequestResponse)) *pb.UpdateFriendRequestResponse {
	resp := &pb.UpdateFriendRequestResponse{
		FriendRequest: &pb.FriendRequest{
			Id:         "1",
			SenderId:   "111",
			ReceiverId: "222",
			Status:     pb.RequestStatus_STATUS_ACCEPTED,
			CreatedAt:  timestamppb.Now(),
		},
	}

	for _, mod := range mods {
		mod(resp)
	}

	return resp
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
			name: "Nil friend request",
			req: fixtureUpdateFriendRequest(func(r *pb.UpdateFriendRequestRequest) {
				r.FriendRequest = nil
			}),
			expecterErr: errchecks.MsgContains("friend request cannot be nil"),
		},
		{
			name: "Missing ID",
			req: fixtureUpdateFriendRequest(func(r *pb.UpdateFriendRequestRequest) {
				r.FriendRequest.Id = ""
			}),
			given: given{
				// simulam comportamentul storage-ului care valideaza ID-ul
				mockStorageAccess: newMockStorageAccess(StorageMockOptions{
					updateFriendRequestFunc: func(ctx context.Context, req *pb.UpdateFriendRequestRequest) (*pb.UpdateFriendRequestResponse, error) {
						return nil, status.Error(codes.InvalidArgument, "friend request ID is required")
					},
				}),
			},
			expecterErr: errchecks.MsgContains("friend request ID is required"),
		},
		{
			name: "Invalid field mask path",
			req: fixtureUpdateFriendRequest(func(r *pb.UpdateFriendRequestRequest) {
				r.FieldMask = &fieldmaskpb.FieldMask{Paths: []string{"sender_id"}}
			}),
			given: given{
				// simulam validarea din storage pentru field_mask
				mockStorageAccess: newMockStorageAccess(StorageMockOptions{
					updateFriendRequestFunc: func(ctx context.Context, req *pb.UpdateFriendRequestRequest) (*pb.UpdateFriendRequestResponse, error) {
						return nil, status.Error(codes.InvalidArgument, "only 'status' can be updated")
					},
				}),
			},
			expecterErr: errchecks.MsgContains("only 'status' can be updated"),
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
