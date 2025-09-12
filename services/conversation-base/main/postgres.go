package main

import (
	"context"
	"database/sql"
	"errors"
	"strconv"
	"time"

	proto "github.com/Costin2000/GoChat---Schwarz-Internship---2025/services/conversation-base/proto"
	"github.com/jackc/pgx/v5/pgconn"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

const (
	pgUniqueViolation     = "23505"
	pgForeignKeyViolation = "23503"
)

type StorageAccess interface {
	createConversation(ctx context.Context, req *proto.CreateConversationRequest) (*proto.CreateConversationResponse, error)
}

type PostgresAccess struct {
	db *sql.DB
}

func newPostgresAccess(db *sql.DB) *PostgresAccess {
	return &PostgresAccess{db: db}
}

func (pa *PostgresAccess) createConversation(ctx context.Context, req *proto.CreateConversationRequest) (*proto.CreateConversationResponse, error) {
	user1ID, err := strconv.ParseInt(req.User1Id, 10, 64)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid user1_id format: %v", err)
	}

	user2ID, err := strconv.ParseInt(req.User2Id, 10, 64)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid user2_id format: %v", err)
	}

	query := `
		INSERT INTO "Conversation" (user1_id, user2_id)
		VALUES ($1, $2)
		RETURNING id, created_at, updated_at;
	`

	var convID int64
	var createdAt, updatedAt time.Time

	err = pa.db.QueryRowContext(ctx, query, user1ID, user2ID).Scan(&convID, &createdAt, &updatedAt)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			switch pgErr.Code {
			case pgUniqueViolation:
				return nil, status.Error(codes.AlreadyExists, "conversation between these users already exists")
			case pgForeignKeyViolation:
				return nil, status.Error(codes.NotFound, "one or both users do not exist")
			}
		}
		return nil, status.Errorf(codes.Internal, "failed to create conversation: %v", err)
	}

	return &proto.CreateConversationResponse{
		Conversation: &proto.Conversation{
			Id:        strconv.FormatInt(convID, 10),
			User1Id:   req.User1Id,
			User2Id:   req.User2Id,
			CreatedAt: timestamppb.New(createdAt),
			UpdatedAt: timestamppb.New(updatedAt),
		},
	}, nil
}
