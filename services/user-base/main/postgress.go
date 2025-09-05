package main

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"strconv"
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
	listUsers(ctx context.Context, req *pb.ListUsersRequest) (*pb.ListUsersResponse, error)
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

func (pa *PostgresAccess) listUsers(ctx context.Context, req *pb.ListUsersRequest) (*pb.ListUsersResponse, error) {
	const (
		defaultPageSize = int64(50)
		maxPageSize     = int64(1000)
	)
	ps := req.GetPageSize()
	if ps <= 0 {
		ps = defaultPageSize
	}
	if ps > maxPageSize {
		ps = maxPageSize
	}

	var lastID int64
	if tok := strings.TrimSpace(req.GetNextPageToken()); tok != "" {
		v, err := strconv.ParseInt(strings.TrimPrefix(tok, "id:"), 10, 64)
		if err != nil || v < 0 {
			return nil, status.Error(codes.InvalidArgument, "invalid nextPageToken")
		}
		lastID = v
	}

	where := []string{}
	args := []any{}

	for _, f := range req.GetFilters() {
		switch x := f.Filter.(type) {
		case *pb.ListUsersFiltersOneOf_FirstName:
			if v := strings.TrimSpace(x.FirstName.GetEquals()); v != "" {
				where = append(where, fmt.Sprintf("LOWER(first_name) = LOWER($%d)", len(args)+1))
				args = append(args, v)
			}
		case *pb.ListUsersFiltersOneOf_LastName:
			if v := strings.TrimSpace(x.LastName.GetEquals()); v != "" {
				where = append(where, fmt.Sprintf("LOWER(last_name) = LOWER($%d)", len(args)+1))
				args = append(args, v)
			}
		}
	}

	if lastID > 0 {
		where = append(where, fmt.Sprintf(`"User".id > $%d`, len(args)+1))
		args = append(args, lastID)
	}

	baseQuery := `SELECT id, first_name, last_name, user_name, email, created_at FROM "User"`

	if len(where) > 0 {
		baseQuery += " WHERE " + strings.Join(where, " AND ")
	}
	baseQuery += fmt.Sprintf(" ORDER BY id ASC LIMIT $%d", len(args)+1)
	args = append(args, ps+1)

	rows, err := pa.db.QueryContext(ctx, baseQuery, args...)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "query error: %v", err)
	}
	defer rows.Close()

	var users []*pb.User
	for rows.Next() {
		var user pb.User
		var createdAt time.Time
		if err := rows.Scan(&user.Id, &user.FirstName, &user.LastName, &user.UserName, &user.Email, &createdAt); err != nil {
			return nil, status.Errorf(codes.Internal, "scan error: %v", err)
		}
		user.CreatedAt = timestamppb.New(createdAt)
		user.Password = ""
		users = append(users, &user)
	}
	if err := rows.Err(); err != nil {
		return nil, status.Errorf(codes.Internal, "rows error: %v", err)
	}

	nextToken := ""
	if int64(len(users)) > ps {
		nextToken = fmt.Sprintf("id:%d", users[ps-1].Id)
		users = users[:ps]
	}

	return &pb.ListUsersResponse{
		NextPageToken: nextToken,
		Users:         users,
	}, nil
}
