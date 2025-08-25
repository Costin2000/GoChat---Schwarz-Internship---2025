package main

import (
	"context"
	"database/sql"
	"errors"
	"log"
	"strings"
	"time"

	pb "github.com/Costin2000/GoChat---Schwarz-Internship---2025/services/user-base/proto"
	"golang.org/x/crypto/bcrypt"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type StorageAccess interface {
	getUserByEmail(ctx context.Context, email string) (*pb.User, error)
	createUser(ctx context.Context, user *pb.User) (*pb.User, error)
}

type PostgresAccess struct {
	db *sql.DB
}

func newPostgresAccess(db *sql.DB) *PostgresAccess {
	return &PostgresAccess{db: db}
}

func (pa *PostgresAccess) createUser(ctx context.Context, user *pb.User) (*pb.User, error) {
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
	err = pa.db.QueryRowContext(ctx, query,
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
	return &pb.User{
		Id:        id,
		FirstName: user.FirstName,
		LastName:  user.LastName,
		UserName:  user.UserName,
		Email:     user.Email,
		Password:  string(hashedPassword), // return hashed
		CreatedAt: timestamppb.New(createdAt),
	}, nil
}

func (pa *PostgresAccess) getUserByEmail(ctx context.Context, email string) (*pb.User, error) {
	var user pb.User
	var createdAt time.Time

	query := `
        SELECT id, first_name, last_name, user_name, email, password, created_at
        FROM "User"
        WHERE email = $1;
    `

	row := pa.db.QueryRowContext(ctx, query, email)

	err := row.Scan(
		&user.Id,
		&user.FirstName,
		&user.LastName,
		&user.UserName,
		&user.Email,
		&user.Password,
		&createdAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, status.Errorf(codes.NotFound, "user with email %s not found", email)
		}

		log.Printf("Database error on GetUser: %v", err)
		return nil, status.Errorf(codes.Internal, "failed to retrieve user")
	}

	user.CreatedAt = timestamppb.New(createdAt)

	return &user, nil
}
