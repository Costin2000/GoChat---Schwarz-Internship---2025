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
		return nil, errors.New("sender and receiver IDs cannot be empty")
	}
	if senderIDStr == receiverIDStr {
		return nil, errors.New("sender and receiver cannot be the same user")
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
        RETURNING created_at;
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
