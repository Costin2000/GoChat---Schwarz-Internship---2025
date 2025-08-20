package main

import (
	"context"
	"database/sql"
	"log"
	"time"

	proto "github.com/Costin2000/GoChat---Schwarz-Internship---2025/services/user-base/proto"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func (s *userBaseServer) GetUser(ctx context.Context, req *proto.GetUserRequest) (*proto.User, error) {

	log.Printf("GetUser request for email: %s", req.Email)
	email := req.GetEmail()
	if email == "" {
		return nil, status.Errorf(codes.InvalidArgument, "email cannot be empty")
	}

	var user proto.User
	var createdAt time.Time // temporary variable for correctly typing the timestamp

	query := `
		SELECT id, first_name, last_name, user_name, email, created_at
		FROM "User"
		WHERE email = $1;
	`

	row := s.db.QueryRowContext(ctx, query, email)

	err := row.Scan(
		&user.Id,
		&user.FirstName,
		&user.LastName,
		&user.UserName,
		&user.Email,
		&createdAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, status.Errorf(codes.NotFound, "user with email %s not found", email)
		}

		log.Printf("Database error on GetUser: %v", err)
		return nil, status.Errorf(codes.Internal, "failed to retrieve user")
	}

	user.CreatedAt = timestamppb.New(createdAt) // cast the retrived createdAt time to the timestamp type

	return &user, nil
}
