package main

import (
	"bytes"
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	proto "github.com/Costin2000/GoChat---Schwarz-Internship---2025/services/friend-request-base/proto"
	"github.com/jackc/pgx/v5/pgconn"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

const (
	pgUniqueViolation     = "23505"
	pgForeignKeyViolation = "23503"
	pgCheckViolation      = "23514"
)

type StorageAccess interface {
	requestCreateFriendRequest(ctx context.Context, req *proto.CreateFriendRequestRequest) (*proto.CreateFriendRequestResponse, error)
	listFriendRequests(ctx context.Context, req *proto.ListFriendRequestsRequest) (*proto.ListFriendRequestsResponse, error)
	requestUpdateFriendRequest(ctx context.Context, req *proto.UpdateFriendRequestRequest) (*proto.UpdateFriendRequestResponse, error)
}

type PostgresAccess struct {
	db *sql.DB
}

func newPostgresAccess(db *sql.DB) *PostgresAccess {
	return &PostgresAccess{db: db}
}

func (pa *PostgresAccess) requestCreateFriendRequest(ctx context.Context, req *proto.CreateFriendRequestRequest) (*proto.CreateFriendRequestResponse, error) {
	senderIDStr := req.SenderId
	receiverIDStr := req.ReceiverId

	senderID, err := strconv.ParseInt(senderIDStr, 10, 64)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid sender ID format: %v", err)
	}

	receiverID, err := strconv.ParseInt(receiverIDStr, 10, 64)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid receiver ID format: %v", err)
	}

	query := `
        INSERT INTO "Friend Requests" (sender_id, receiver_id)
        VALUES ($1, $2)
        RETURNING id, created_at;
    `

	var requestID int64
	var createdAt time.Time
	err = pa.db.QueryRowContext(ctx, query, senderID, receiverID).Scan(&requestID, &createdAt)

	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {

			switch pgErr.Code {
			case pgUniqueViolation:
				return nil, status.Error(codes.AlreadyExists, "a friend request between these users already exists")

			case pgForeignKeyViolation:
				return nil, status.Error(codes.NotFound, "one or both users do not exist")

			case pgCheckViolation:
				return nil, status.Error(codes.InvalidArgument, "sender and receiver cannot be the same user")
			}
		}
		return nil, status.Errorf(codes.Internal, "failed to create friend request: %v", err)
	}

	return &proto.CreateFriendRequestResponse{
		Request: &proto.FriendRequest{
			Id:         strconv.FormatInt(requestID, 10),
			SenderId:   senderIDStr,
			ReceiverId: receiverIDStr,
			Status:     proto.RequestStatus_STATUS_PENDING,
			CreatedAt:  timestamppb.New(createdAt),
		},
	}, nil

}

func (pa *PostgresAccess) listFriendRequests(ctx context.Context, req *proto.ListFriendRequestsRequest) (*proto.ListFriendRequestsResponse, error) {

	// Creating the db query via formatted string - argument list pair to ensure protection against SQL Injection
	var args []any
	var query bytes.Buffer
	argCounter := 1

	query.WriteString(`SELECT id, sender_id, receiver_id, status, created_at FROM "Friend Requests" WHERE 1=1`)

	// Iterate through the fields; frontend should always set the receiver id field to the requesting user's id to ensure not showing all requests in the system
	for _, fil := range req.Filters {
		switch filterTyped := fil.Filter.(type) {
		case *proto.ListFriendRequestsFiltersOneOf_ReceiverId:
			receiverId, err := strconv.ParseInt(filterTyped.ReceiverId, 10, 64)
			if err != nil {
				return nil, status.Errorf(codes.InvalidArgument, "invalid receiver_id filter format: %v", err)
			}
			query.WriteString(fmt.Sprintf(" AND receiver_id = $%d", argCounter))
			args = append(args, receiverId)
			argCounter++
		case *proto.ListFriendRequestsFiltersOneOf_SenderId:
			senderId, err := strconv.ParseInt(filterTyped.SenderId, 10, 64)
			if err != nil {
				return nil, status.Errorf(codes.InvalidArgument, "invalid sender_id filter format: %v", err)
			}
			query.WriteString(fmt.Sprintf(" AND sender_id = $%d", argCounter))
			args = append(args, senderId)
			argCounter++
		case *proto.ListFriendRequestsFiltersOneOf_Status:
			query.WriteString(fmt.Sprintf(" AND status = $%d", argCounter))
			args = append(args, strings.ToLower(filterTyped.Status))
			argCounter++
		}
	}

	// If the user is requesting the FR's for the first time in the session, the frontend should send an empty nextPageToken. Afterwards the frontend should return the last received nextPageToken back to the server
	if req.NextPageToken != "" {
		nextPageTokenId, err := strconv.ParseInt(req.NextPageToken, 10, 64)
		if err != nil {
			return nil, status.Errorf(codes.InvalidArgument, "invalid nextPageToken format: %v", err)
		}
		query.WriteString(fmt.Sprintf(" AND id > $%d", argCounter))
		args = append(args, nextPageTokenId)
		argCounter++
	}

	// Always sort id's in ascending order to ensure the same request is never showed again on different pages
	query.WriteString(fmt.Sprintf(" ORDER BY id ASC LIMIT $%d", argCounter))
	args = append(args, req.PageSize)

	log.Printf("Executing query: %s with args: %v", query.String(), args)
	rows, err := pa.db.QueryContext(ctx, query.String(), args...)
	if err != nil {
		log.Printf("DATABASE ERROR: Query '%s' with args %v failed: %v", query.String(), args, err)
		return nil, status.Errorf(codes.Internal, "database query failed: %v", err)
	}

	var lastReqId int64
	var requests []*proto.FriendRequest
	for rows.Next() {
		var fr proto.FriendRequest
		var senderID, receiverID int64
		var createdAt time.Time
		var statusStr string

		if err := rows.Scan(&lastReqId, &senderID, &receiverID, &statusStr, &createdAt); err != nil {
			return nil, status.Errorf(codes.Internal, "failed to scan row: %v", err)
		}

		fr.Id = strconv.FormatInt(lastReqId, 10)
		fr.SenderId = strconv.FormatInt(senderID, 10)
		fr.ReceiverId = strconv.FormatInt(receiverID, 10)
		fr.CreatedAt = timestamppb.New(createdAt)

		enumKey := fmt.Sprintf("STATUS_%s", strings.ToUpper(statusStr))
		if val, ok := proto.RequestStatus_value[enumKey]; ok {
			fr.Status = proto.RequestStatus(val)
		} else {
			fr.Status = proto.RequestStatus_STATUS_UNKNOWN // fallback
		}

		requests = append(requests, &fr)
	}

	var nextPageToken string = ""
	if len(requests) == int(req.PageSize) {
		nextPageToken = strconv.FormatInt(lastReqId, 10)
	}

	return &proto.ListFriendRequestsResponse{
		NextPageToken: nextPageToken,
		Requests:      requests,
	}, nil
}

func (pa *PostgresAccess) requestUpdateFriendRequest(ctx context.Context, req *proto.UpdateFriendRequestRequest) (*proto.UpdateFriendRequestResponse, error) {
	if req.FriendRequest == nil {
		return nil, status.Error(codes.InvalidArgument, "friend request must be provided")
	}

	// validam field_mask
	allowed := map[string]bool{"status": true}
	for _, path := range req.FieldMask.Paths {
		if !allowed[path] {
			return nil, status.Errorf(codes.InvalidArgument, "field %s cannot be updated", path)
		}
	}

	id, err := strconv.ParseInt(req.FriendRequest.Id, 10, 64)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid friend request ID: %v", err)
	}

	// mapam enum-ul protobuf la enum-ul postgres
	var statusStr string
	switch req.FriendRequest.Status {
	case proto.RequestStatus_STATUS_PENDING:
		statusStr = "pending"
	case proto.RequestStatus_STATUS_ACCEPTED:
		statusStr = "accepted"
	case proto.RequestStatus_STATUS_REJECTED:
		statusStr = "rejected"
	default:
		return nil, status.Errorf(codes.InvalidArgument, "unsupported status value")
	}

	query := `
        UPDATE "Friend Requests"
        SET status = $1
        WHERE id = $2
        RETURNING sender_id, receiver_id, status, created_at;
    `

	var senderID, receiverID int64
	var statusDB string
	var createdAt time.Time

	err = pa.db.QueryRowContext(ctx, query, statusStr, id).Scan(&senderID, &receiverID, &statusDB, &createdAt)
	if err == sql.ErrNoRows {
		return nil, status.Error(codes.NotFound, "friend request not found")
	}
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to update friend request: %v", err)
	}

	// convertim statusul din DB in enum-ul protobuf
	statusEnum := proto.RequestStatus_STATUS_UNKNOWN
	switch strings.ToLower(statusDB) {
	case "pending":
		statusEnum = proto.RequestStatus_STATUS_PENDING
	case "accepted":
		statusEnum = proto.RequestStatus_STATUS_ACCEPTED
	case "rejected":
		statusEnum = proto.RequestStatus_STATUS_REJECTED
	case "blocked":
		statusEnum = proto.RequestStatus_STATUS_REJECTED // sau STATUS_UNKNOWN
	}

	return &proto.UpdateFriendRequestResponse{
		FriendRequest: &proto.FriendRequest{
			Id:         strconv.FormatInt(id, 10),
			SenderId:   strconv.FormatInt(senderID, 10),
			ReceiverId: strconv.FormatInt(receiverID, 10),
			Status:     statusEnum,
			CreatedAt:  timestamppb.New(createdAt),
		},
	}, nil
}
