package main

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	pb "github.com/Costin2000/GoChat---Schwarz-Internship---2025/services/message-base/proto"
	"github.com/jackc/pgx/v5/pgconn"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type StorageAccess interface {
	createMessage(ctx context.Context, m *pb.Message) (*pb.Message, error)
}

type PostgresAccess struct{ db *sql.DB }

func newPostgresAccess(db *sql.DB) *PostgresAccess { return &PostgresAccess{db: db} }

func (pa *PostgresAccess) createMessage(ctx context.Context, m *pb.Message) (*pb.Message, error) {
	query := `
	  INSERT INTO "Message"(conversation_id, sender_id, content)
	  VALUES ($1, $2, $3)
	  RETURNING id, created_at;
	`
	var (
		id        int64
		createdAt time.Time
	)
	err := pa.db.QueryRowContext(ctx, query, m.ConversationId, m.SenderId, m.Content).Scan(&id, &createdAt)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			if pgErr.Code == "23503" {
				return nil, status.Error(codes.NotFound, "conversation or sender not found")
			}
		}
		return nil, status.Errorf(codes.Internal, "failed to create message: %v", err)
	}

	return &pb.Message{
		Id:             id,
		ConversationId: m.ConversationId,
		SenderId:       m.SenderId,
		Content:        m.Content,
		CreatedAt:      timestamppb.New(createdAt),
	}, nil
}

func (pa *PostgresAccess) checkExists(ctx context.Context, table string, id int64) (bool, error) {
	q := fmt.Sprintf(`SELECT 1 FROM "%s" WHERE id=$1`, table)
	var one int
	if err := pa.db.QueryRowContext(ctx, q, id).Scan(&one); err != nil {
		if err == sql.ErrNoRows {
			return false, nil
		}
		return false, err
	}
	return true, nil
}
