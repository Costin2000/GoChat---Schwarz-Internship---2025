package main

import (
	"context"
	"database/sql"
	"errors"
	"strconv"
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
