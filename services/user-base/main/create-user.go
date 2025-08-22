package main

import (
	"context"
	"database/sql"
	"errors"
	"log"
	"strings"
	"time"

	proto "github.com/Costin2000/GoChat---Schwarz-Internship---2025/services/user-base/proto"
	"golang.org/x/crypto/bcrypt"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func (s *userBaseServer) CreateUser(ctx context.Context, req *proto.CreateUserRequest) (*proto.CreateUserResponse, error) {
	user := req.GetUser()
	if user == nil {
		return nil, status.Errorf(codes.InvalidArgument, "user object is required")
	}

	// Basic validation
	if strings.TrimSpace(user.FirstName) == "" ||
		strings.TrimSpace(user.LastName) == "" ||
		strings.TrimSpace(user.UserName) == "" ||
		strings.TrimSpace(user.Email) == "" ||
		strings.TrimSpace(user.Password) == "" {
		return nil, status.Errorf(codes.InvalidArgument, "all fields are required")
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		log.Printf("Error hashing password: %v", err)
		return nil, status.Errorf(codes.Internal, "failed to hash password")
	}

	// Insert into DB
	query := `
		INSERT INTO "User" (first_name, last_name, user_name, email, password)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, created_at;
	`

	var id int64
	var createdAt time.Time
	err = s.db.QueryRowContext(ctx, query,
		user.FirstName, user.LastName, user.UserName, user.Email, string(hashedPassword),
	).Scan(&id, &createdAt)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, status.Errorf(codes.Internal, "user creation failed")
		}
		if strings.Contains(err.Error(), "duplicate key") {
			return nil, status.Errorf(codes.AlreadyExists, "user with this email or username already exists")
		}
		log.Printf("Database error: %v", err)
		return nil, status.Errorf(codes.Internal, "failed to create user")
	}

	// Prepare response
	newUser := &proto.User{
		Id:        id,
		FirstName: user.FirstName,
		LastName:  user.LastName,
		UserName:  user.UserName,
		Email:     user.Email,
		Password:  string(hashedPassword), // return hashed
		CreatedAt: timestamppb.New(createdAt),
	}

	return &proto.CreateUserResponse{
		User: newUser,
	}, nil
}
