package main

import (
	"context"
	"database/sql"
	"errors"
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
	requestUpdateFriendRequest(ctx context.Context, req *proto.UpdateFriendRequestRequest) (*proto.UpdateFriendRequestResponse, error)
}

type PostgresAccess struct {
	db *sql.DB
}

func newPostgresAccess(db *sql.DB) *PostgresAccess {
	return &PostgresAccess{db: db}
}

func (pa *PostgresAccess) requestCreateFriendRequest(ctx context.Context, req *proto.CreateFriendRequestRequest) (*proto.CreateFriendRequestResponse, error) {
	senderIDStr := req.GetSenderId()
	receiverIDStr := req.GetReceiverId()

	if senderIDStr == "" || receiverIDStr == "" {
		return nil, status.Error(codes.InvalidArgument, "sender and receiver IDs cannot be empty")
	}
	if senderIDStr == receiverIDStr {
		return nil, status.Error(codes.InvalidArgument, "sender and receiver cannot be the same user")
	}

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

func (pa *PostgresAccess) requestUpdateFriendRequest(ctx context.Context, req *proto.UpdateFriendRequestRequest) (*proto.UpdateFriendRequestResponse, error) {
	if req.GetFriendRequest() == nil {
		return nil, status.Error(codes.InvalidArgument, "friend request cannot be nil")
	}
	if req.FieldMask == nil || len(req.FieldMask.GetPaths()) == 0 {
		return nil, status.Error(codes.InvalidArgument, "field_mask must contain at least 'status'")
	}
	for _, path := range req.FieldMask.GetPaths() {
		if path != "status" {
			return nil, status.Errorf(codes.InvalidArgument, "only 'status' can be updated, got: %s", path)
		}
	}

	idStr := req.FriendRequest.GetId()
	if idStr == "" {
		return nil, status.Error(codes.InvalidArgument, "friend request ID is required")
	}
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid id format: %v", err)
	}

	// mapăm enum-ul proto -> text enum în DB
	newStatusProto := req.FriendRequest.GetStatus()
	var dbStatus string
	switch newStatusProto {
	case proto.RequestStatus_STATUS_PENDING:
		dbStatus = "pending"
	case proto.RequestStatus_STATUS_ACCEPTED:
		dbStatus = "accepted"
	case proto.RequestStatus_STATUS_REJECTED:
		dbStatus = "rejected"
	default:
		return nil, status.Error(codes.InvalidArgument, "invalid status value")
	}

	query := `
		UPDATE "Friend Requests"
		SET status = $1
		WHERE id = $2
		RETURNING id, sender_id, receiver_id, status, created_at;
	`

	var retID, senderID, receiverID int64
	var statusStr string
	var createdAt time.Time

	err = pa.db.QueryRowContext(ctx, query, dbStatus, id).Scan(
		&retID,
		&senderID,
		&receiverID,
		&statusStr,
		&createdAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, status.Error(codes.NotFound, "friend request not found")
		}
		return nil, status.Errorf(codes.Internal, "failed to update friend request: %v", err)
	}

	// mapăm status-ul DB -> enum proto
	var protoStatus proto.RequestStatus
	switch strings.ToLower(statusStr) {
	case "pending":
		protoStatus = proto.RequestStatus_STATUS_PENDING
	case "accepted":
		protoStatus = proto.RequestStatus_STATUS_ACCEPTED
	case "rejected":
		protoStatus = proto.RequestStatus_STATUS_REJECTED
	default:
		protoStatus = proto.RequestStatus_STATUS_UNKNOWN
	}

	return &proto.UpdateFriendRequestResponse{
		FriendRequest: &proto.FriendRequest{
			Id:         strconv.FormatInt(retID, 10),
			SenderId:   strconv.FormatInt(senderID, 10),
			ReceiverId: strconv.FormatInt(receiverID, 10),
			Status:     protoStatus,
			CreatedAt:  timestamppb.New(createdAt),
		},
	}, nil
}
