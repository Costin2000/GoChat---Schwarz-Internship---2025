package main

import (
	"context"
	"log"
	"maps"
	"slices"
	"strconv"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	aggrpb "github.com/Costin2000/GoChat---Schwarz-Internship---2025/services/aggregator/proto"
	frpb "github.com/Costin2000/GoChat---Schwarz-Internship---2025/services/friend-request-base/proto"
	userpb "github.com/Costin2000/GoChat---Schwarz-Internship---2025/services/user-base/proto"
)

func (svc *AggregatorService) FetchUserFriends(ctx context.Context, req *aggrpb.FetchUserFriendsRequest) (*aggrpb.FetchUserFriendsResponse, error) {

	if req.UserId == "" {
		return nil, status.Error(codes.InvalidArgument, "UserId cannot be empty")
	}

	reqId := req.UserId

	// Query all accepted friend requests concerning the requesting user
	sentFiltersSender := []*frpb.ListFriendRequestsFiltersOneOf{
		{Filter: &frpb.ListFriendRequestsFiltersOneOf_SenderId{SenderId: reqId}},
		{Filter: &frpb.ListFriendRequestsFiltersOneOf_Status{Status: frpb.RequestStatus_STATUS_ACCEPTED.String()}},
	}
	sentFiltersReceiver := []*frpb.ListFriendRequestsFiltersOneOf{
		{Filter: &frpb.ListFriendRequestsFiltersOneOf_ReceiverId{ReceiverId: reqId}},
		{Filter: &frpb.ListFriendRequestsFiltersOneOf_Status{Status: frpb.RequestStatus_STATUS_ACCEPTED.String()}},
	}

	friendRequests := []*frpb.FriendRequest{}
	var err error
	var friendRequestsRsp *frpb.ListFriendRequestsResponse

	nextPageToken := ""
	for {
		listFrReqSender := &frpb.ListFriendRequestsRequest{
			NextPageToken: nextPageToken,
			PageSize:      1000,
			Filters:       sentFiltersSender,
		}
		friendRequestsRsp, err = svc.frClient.ListFriendRequests(ctx, listFrReqSender)
		if err != nil {
			log.Printf("Error fetching sent friend requests for user %s: %v", reqId, err)
			return nil, status.Error(codes.Internal, "Failed to fetch friend data")
		}
		friendRequests = append(friendRequests, friendRequestsRsp.Requests...)

		nextPageToken = friendRequestsRsp.GetNextPageToken()
		if nextPageToken == "" {
			break
		}
	}

	for {
		listFrReqReceiver := &frpb.ListFriendRequestsRequest{
			NextPageToken: nextPageToken,
			PageSize:      1000,
			Filters:       sentFiltersReceiver,
		}
		friendRequestsRsp, err = svc.frClient.ListFriendRequests(ctx, listFrReqReceiver)
		if err != nil {
			log.Printf("Error fetching received friend requests for user %s: %v", reqId, err)
			return nil, status.Error(codes.Internal, "Failed to fetch friend data")
		}
		friendRequests = append(friendRequests, friendRequestsRsp.Requests...)

		nextPageToken = friendRequestsRsp.GetNextPageToken()
		if nextPageToken == "" {
			break
		}
	}

	// Query users by the ids returned in the accepted friend requests
	friendIDsMap := make(map[int64]struct{})
	for _, fr := range friendRequests {
		var friendIdStr string
		if fr.SenderId == reqId {
			friendIdStr = fr.ReceiverId
		} else {
			friendIdStr = fr.SenderId
		}
		id, err := strconv.ParseInt(friendIdStr, 10, 64)
		if err != nil {
			log.Printf("Could not parse user ID '%s', skipping: %v", friendIdStr, err)
			continue
		}
		friendIDsMap[id] = struct{}{}
	}

	if len(friendIDsMap) == 0 {
		return &aggrpb.FetchUserFriendsResponse{Friends: []*userpb.User{}}, nil
	}

	uniqueIds := slices.Collect(maps.Keys(friendIDsMap))
	userFilters := []*userpb.ListUsersFiltersOneOf{
		{
			Filter: &userpb.ListUsersFiltersOneOf_UserIds{
				UserIds: &userpb.FilterByIdIn{
					UserId: uniqueIds,
				},
			},
		},
	}

	var Friends []*userpb.User
	userNextPageToken := ""
	for {
		listUsersReq := &userpb.ListUsersRequest{
			PageSize:      1000,
			NextPageToken: userNextPageToken,
			Filters:       userFilters,
		}

		userRsp, err := svc.userBaseClient.ListUsers(ctx, listUsersReq)
		if err != nil {
			log.Printf("Failed to list users by IDs: %v", err)
			return nil, status.Error(codes.Internal, "Failed to fetch user friends")
		}

		Friends = append(Friends, userRsp.Users...)

		userNextPageToken = userRsp.GetNextPageToken()
		if userNextPageToken == "" {
			break
		}
	}

	return &aggrpb.FetchUserFriendsResponse{Friends: Friends}, nil
}
