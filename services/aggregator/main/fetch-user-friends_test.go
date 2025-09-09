package main

import (
	"context"
	"errors"
	"testing"

	errchecks "github.com/Costin2000/GoChat---Schwarz-Internship---2025/pkg"
	aggrpb "github.com/Costin2000/GoChat---Schwarz-Internship---2025/services/aggregator/proto"
	frpb "github.com/Costin2000/GoChat---Schwarz-Internship---2025/services/friend-request-base/proto"
	userpb "github.com/Costin2000/GoChat---Schwarz-Internship---2025/services/user-base/proto"
	"github.com/google/go-cmp/cmp"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/testing/protocmp"
)

func Test_FetchFriendsUnit(t *testing.T) {

	user2 := &userpb.User{Id: 2, UserName: "friend_one"}
	user3 := &userpb.User{Id: 3, UserName: "friend_two"}

	type Given struct {
		userClient UserClient
		frClient   FriendRequestClient
	}

	tests := []struct {
		name            string
		userId          string
		expectedFriends []*userpb.User
		listUserReq     *userpb.ListUsersRequest
		listFrReq       *frpb.ListFriendRequestsRequest
		expectedErr     errchecks.Check
		given           Given
	}{
		{
			name:        "Error: empty user id",
			userId:      "",
			expectedErr: errchecks.MsgContains("UserId cannot be empty"),
		},
		{
			name:   "Success: User has friends",
			userId: "1",
			listUserReq: &userpb.ListUsersRequest{
				PageSize: 10,
			},
			listFrReq: &frpb.ListFriendRequestsRequest{
				PageSize: 10,
			},
			given: Given{
				userClient: &userClientMock{
					ListUsersFunc: func(ctx context.Context, req *userpb.ListUsersRequest, opts ...grpc.CallOption) (*userpb.ListUsersResponse, error) {
						return &userpb.ListUsersResponse{Users: []*userpb.User{user2, user3}}, nil
					},
				},
				frClient: &frClientMock{

					ListFrFunc: func(ctx context.Context, req *frpb.ListFriendRequestsRequest, opts ...grpc.CallOption) (*frpb.ListFriendRequestsResponse, error) {
						if req.Filters[0].GetSenderId() == "1" {
							return &frpb.ListFriendRequestsResponse{
								Requests: []*frpb.FriendRequest{
									{SenderId: "1", ReceiverId: "2", Status: frpb.RequestStatus_STATUS_ACCEPTED},
								},
							}, nil
						}
						if req.Filters[0].GetReceiverId() == "1" {
							return &frpb.ListFriendRequestsResponse{
								Requests: []*frpb.FriendRequest{
									{SenderId: "3", ReceiverId: "1", Status: frpb.RequestStatus_STATUS_ACCEPTED},
								},
							}, nil
						}
						return &frpb.ListFriendRequestsResponse{}, nil
					},
				},
			},
			expectedFriends: []*userpb.User{user2, user3},
			expectedErr:     nil,
		},
		{
			name:   "Success: No friends found",
			userId: "1",
			listUserReq: &userpb.ListUsersRequest{
				PageSize: 10,
			},
			listFrReq: &frpb.ListFriendRequestsRequest{
				PageSize: 10,
			},
			given: Given{
				frClient: &frClientMock{
					ListFrFunc: func(ctx context.Context, req *frpb.ListFriendRequestsRequest, opts ...grpc.CallOption) (*frpb.ListFriendRequestsResponse, error) {
						return &frpb.ListFriendRequestsResponse{Requests: []*frpb.FriendRequest{}}, nil
					},
				},
			},
			expectedFriends: []*userpb.User{},
		},
		{
			name:   "Error: ListFriendRequests call fails",
			userId: "1",
			given: Given{
				frClient: &frClientMock{
					ListFrFunc: func(ctx context.Context, req *frpb.ListFriendRequestsRequest, opts ...grpc.CallOption) (*frpb.ListFriendRequestsResponse, error) {
						return nil, errors.New("ListFriendRequests endpoint unavailable")
					},
				},
			},
			expectedErr: errchecks.MsgContains("Failed to fetch friend data"),
		},
		{
			name:   "Error: ListUsers call fails",
			userId: "1",
			given: Given{
				frClient: &frClientMock{
					ListFrFunc: func(ctx context.Context, req *frpb.ListFriendRequestsRequest, opts ...grpc.CallOption) (*frpb.ListFriendRequestsResponse, error) {
						if req.Filters[0].GetSenderId() == "1" {
							return &frpb.ListFriendRequestsResponse{
								Requests: []*frpb.FriendRequest{{SenderId: "1", ReceiverId: "2"}},
							}, nil
						}
						return &frpb.ListFriendRequestsResponse{}, nil
					},
				},
				userClient: &userClientMock{
					ListUsersFunc: func(ctx context.Context, req *userpb.ListUsersRequest, opts ...grpc.CallOption) (*userpb.ListUsersResponse, error) {
						return nil, errors.New("ListUser endpoint unavailable")
					},
				},
			},
			expectedErr: errchecks.MsgContains("Failed to fetch user friends"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := NewMockService(ServiceMockOptions{
				userClient: tt.given.userClient,
				frClient:   tt.given.frClient,
			})
			req := &aggrpb.FetchUserFriendsRequest{
				UserId: tt.userId,
			}

			resp, err := svc.FetchUserFriends(context.Background(), req)

			errchecks.Assert(t, err, tt.expectedErr)
			if tt.expectedErr == nil {
				expectedRsp := &aggrpb.FetchUserFriendsResponse{
					Friends: tt.expectedFriends,
				}
				if diff := cmp.Diff(expectedRsp, resp, protocmp.Transform()); diff != "" {
					t.Errorf("FetchUserFriends response mismatch (-want +got):\n%s", diff)
				}
			}
		})
	}
}
