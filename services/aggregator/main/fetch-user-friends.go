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
	reqIdInt, err := strconv.ParseInt(reqId, 10, 64)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid user ID format: %v", err)
	}

	friendRequests := []*frpb.FriendRequest{}
	involvedUserIDs := make(map[int64]struct{})
	var friendRequestsRsp *frpb.ListFriendRequestsResponse

	if req.ShowFriends { // Query for user's friends

		// Query all accepted friend requests concerning the requesting user
		sentFiltersSender := []*frpb.ListFriendRequestsFiltersOneOf{
			{Filter: &frpb.ListFriendRequestsFiltersOneOf_SenderId{SenderId: reqId}},
			{Filter: &frpb.ListFriendRequestsFiltersOneOf_Status{Status: "accepted"}},
		}
		sentFiltersReceiver := []*frpb.ListFriendRequestsFiltersOneOf{
			{Filter: &frpb.ListFriendRequestsFiltersOneOf_ReceiverId{ReceiverId: reqId}},
			{Filter: &frpb.ListFriendRequestsFiltersOneOf_Status{Status: "accepted"}},
		}

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

		nextPageToken = ""
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
	} else { // Gather IDs of all users who have ANY friend request with the requester.
		involvedUserIDs[reqIdInt] = struct{}{}

		// Filters to get all requests sent BY or TO the user
		filters := [][]*frpb.ListFriendRequestsFiltersOneOf{
			{{Filter: &frpb.ListFriendRequestsFiltersOneOf_SenderId{SenderId: reqId}}},
			{{Filter: &frpb.ListFriendRequestsFiltersOneOf_ReceiverId{ReceiverId: reqId}}},
		}

		for _, filterSet := range filters {
			nextPageToken := ""
			for {
				listReq := &frpb.ListFriendRequestsRequest{PageSize: 1000, Filters: filterSet, NextPageToken: nextPageToken}
				frRsp, err := svc.frClient.ListFriendRequests(ctx, listReq)
				if err != nil {
					log.Printf("Error fetching friend requests to build exclusion list for user %s: %v", reqId, err)
					return nil, status.Error(codes.Internal, "Failed to fetch user relationship data")
				}
				for _, fr := range frRsp.Requests {
					senderID, _ := strconv.ParseInt(fr.SenderId, 10, 64)
					receiverID, _ := strconv.ParseInt(fr.ReceiverId, 10, 64)
					involvedUserIDs[senderID] = struct{}{}
					involvedUserIDs[receiverID] = struct{}{}
				}
				nextPageToken = frRsp.GetNextPageToken()
				if nextPageToken == "" {
					break
				}
			}
		}
	}

	var Users []*userpb.User

	if req.ShowFriends { // Query for friends ids
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
			return &aggrpb.FetchUserFriendsResponse{Users: []*userpb.User{}}, nil
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

			Users = append(Users, userRsp.Users...)

			userNextPageToken = userRsp.GetNextPageToken()
			if userNextPageToken == "" {
				break
			}
		}

	} else { // Fetch all users and filter out those who are already involved.
		allUsers := []*userpb.User{}
		nextPageToken := ""
		for {
			listUsersReq := &userpb.ListUsersRequest{PageSize: 1000, NextPageToken: nextPageToken}
			userRsp, err := svc.userBaseClient.ListUsers(ctx, listUsersReq)
			if err != nil {
				log.Printf("Error fetching all users: %v", err)
				return nil, status.Error(codes.Internal, "Failed to fetch all users list")
			}
			allUsers = append(allUsers, userRsp.Users...)
			nextPageToken = userRsp.GetNextPageToken()
			if nextPageToken == "" {
				break
			}
		}

		// Filter out the excluded users and assign to the final slice.
		for _, user := range allUsers {
			if _, found := involvedUserIDs[user.Id]; !found {
				Users = append(Users, user)
			}
		}
	}

	return &aggrpb.FetchUserFriendsResponse{Users: Users}, nil
}
