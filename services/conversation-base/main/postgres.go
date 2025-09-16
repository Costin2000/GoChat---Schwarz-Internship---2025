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
	listConversations(ctx context.Context, req *proto.ListConversationsRequest) (*proto.ListConversationsResponse, error)
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

func (pa *PostgresAccess) listConversations(ctx context.Context, req *proto.ListConversationsRequest) (*proto.ListConversationsResponse, error) {
	if req.UserId == "" {
		return nil, status.Errorf(codes.InvalidArgument, "user_id must be provided")
	}

	userID, convErr := strconv.ParseInt(req.UserId, 10, 64)
	if convErr != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid user_id format: %v", convErr)
	}

	query := `
		SELECT id, user1_id, user2_id, created_at, updated_at
		FROM "Conversation"
		WHERE user1_id = $1 OR user2_id = $1
		ORDER BY updated_at DESC;
	`
	rows, err := pa.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to list conversations: %v", err)
	}
	defer rows.Close()

	var conversations []*proto.Conversation
	for rows.Next() {
		var (
			convID  int64
			user1ID int64
			user2ID int64
			created time.Time
			updated time.Time
		)
		if err := rows.Scan(&convID, &user1ID, &user2ID, &created, &updated); err != nil {
			return nil, status.Errorf(codes.Internal, "error scanning row: %v", err)
		}
		conversations = append(conversations, &proto.Conversation{
			Id:        strconv.FormatInt(convID, 10),
			User1Id:   strconv.FormatInt(user1ID, 10),
			User2Id:   strconv.FormatInt(user2ID, 10),
			CreatedAt: timestamppb.New(created),
			UpdatedAt: timestamppb.New(updated),
		})
	}

	return &proto.ListConversationsResponse{Conversations: conversations}, nil
}
