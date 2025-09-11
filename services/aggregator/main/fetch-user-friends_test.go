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

func hasSenderID(filters []*frpb.ListFriendRequestsFiltersOneOf, id string) bool {
	for _, f := range filters {
		if s, ok := f.Filter.(*frpb.ListFriendRequestsFiltersOneOf_SenderId); ok && s.SenderId == id {
			return true
		}
	}
	return false
}

func hasReceiverID(filters []*frpb.ListFriendRequestsFiltersOneOf, id string) bool {
	for _, f := range filters {
		if r, ok := f.Filter.(*frpb.ListFriendRequestsFiltersOneOf_ReceiverId); ok && r.ReceiverId == id {
			return true
		}
	}
	return false
}

func Test_FetchFriendsUnit(t *testing.T) {

	user1 := &userpb.User{Id: 1, UserName: "requester"}
	user2 := &userpb.User{Id: 2, UserName: "friend_one"}
	user3 := &userpb.User{Id: 3, UserName: "friend_two"}
	user4 := &userpb.User{Id: 4, UserName: "non_approached_one"}
	user5 := &userpb.User{Id: 5, UserName: "non_approached_two"}

	type Given struct {
		userClient UserClient
		frClient   FriendRequestClient
	}

	tests := []struct {
		name          string
		expectedUsers []*userpb.User
		req           *aggrpb.FetchUserFriendsRequest
		expectedErr   errchecks.Check
		given         Given
	}{
		{
			name:        "Error: empty user id",
			req:         &aggrpb.FetchUserFriendsRequest{UserId: ""},
			expectedErr: errchecks.MsgContains("UserId cannot be empty"),
		},
		{
			name: "Success: User has friends",
			req:  &aggrpb.FetchUserFriendsRequest{UserId: "1", ShowFriends: true},
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
			expectedUsers: []*userpb.User{user2, user3},
		},
		{
			name: "Success: No friends found",
			req:  &aggrpb.FetchUserFriendsRequest{UserId: "1", ShowFriends: true},
			given: Given{
				frClient: &frClientMock{
					ListFrFunc: func(ctx context.Context, req *frpb.ListFriendRequestsRequest, opts ...grpc.CallOption) (*frpb.ListFriendRequestsResponse, error) {
						return &frpb.ListFriendRequestsResponse{Requests: []*frpb.FriendRequest{}}, nil
					},
				},
			},
			expectedUsers: []*userpb.User{},
		},
		{
			name: "Error: ListFriendRequests call fails",
			req:  &aggrpb.FetchUserFriendsRequest{UserId: "1", ShowFriends: true},
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
			name: "Error: ListUsers call fails",
			req:  &aggrpb.FetchUserFriendsRequest{UserId: "1", ShowFriends: true},
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
		{
			name: "Success: Find non-approached users (ShowFriends=false)",
			req:  &aggrpb.FetchUserFriendsRequest{UserId: "1", ShowFriends: false},
			given: Given{
				frClient: &frClientMock{
					ListFrFunc: func(ctx context.Context, req *frpb.ListFriendRequestsRequest, opts ...grpc.CallOption) (*frpb.ListFriendRequestsResponse, error) {
						if hasSenderID(req.Filters, "1") {
							return &frpb.ListFriendRequestsResponse{Requests: []*frpb.FriendRequest{{SenderId: "1", ReceiverId: "2"}}}, nil
						}
						if hasReceiverID(req.Filters, "1") {
							return &frpb.ListFriendRequestsResponse{Requests: []*frpb.FriendRequest{{SenderId: "3", ReceiverId: "1"}}}, nil
						}
						return &frpb.ListFriendRequestsResponse{}, nil
					},
				},
				userClient: &userClientMock{
					ListUsersFunc: func(ctx context.Context, req *userpb.ListUsersRequest, opts ...grpc.CallOption) (*userpb.ListUsersResponse, error) {
						return &userpb.ListUsersResponse{Users: []*userpb.User{user1, user2, user3, user4, user5}}, nil
					},
				},
			},
			expectedUsers: []*userpb.User{user4, user5},
		},
		{
			name: "Success: No non-approached users found (ShowFriends=false)",
			req:  &aggrpb.FetchUserFriendsRequest{UserId: "1", ShowFriends: false},
			given: Given{
				frClient: &frClientMock{
					ListFrFunc: func(ctx context.Context, req *frpb.ListFriendRequestsRequest, opts ...grpc.CallOption) (*frpb.ListFriendRequestsResponse, error) {
						if hasSenderID(req.Filters, "1") {
							return &frpb.ListFriendRequestsResponse{Requests: []*frpb.FriendRequest{{SenderId: "1", ReceiverId: "2"}}}, nil
						}
						return &frpb.ListFriendRequestsResponse{}, nil
					},
				},
				userClient: &userClientMock{
					ListUsersFunc: func(ctx context.Context, req *userpb.ListUsersRequest, opts ...grpc.CallOption) (*userpb.ListUsersResponse, error) {
						return &userpb.ListUsersResponse{Users: []*userpb.User{user1, user2}}, nil
					},
				},
			},
			expectedUsers: []*userpb.User{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := NewMockService(ServiceMockOptions{
				userClient: tt.given.userClient,
				frClient:   tt.given.frClient,
			})

			resp, err := svc.FetchUserFriends(context.Background(), tt.req)
			errchecks.Assert(t, err, tt.expectedErr)
			if tt.expectedErr == nil {
				expectedRsp := &aggrpb.FetchUserFriendsResponse{
					Users: tt.expectedUsers,
				}
				if diff := cmp.Diff(expectedRsp, resp, protocmp.Transform()); diff != "" {
					t.Errorf("FetchUserFriends response mismatch (-want +got):\n%s", diff)
				}
			}
		})
	}
}
