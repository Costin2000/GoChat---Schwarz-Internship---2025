package main

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	pb "github.com/Costin2000/GoChat---Schwarz-Internship---2025/services/user-base/proto"
	"github.com/jackc/pgx/v5/pgconn"
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
		return nil, status.Errorf(codes.Internal, "failed to hash password: %v", err)
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
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" { // unique_violation
			return nil, status.Error(codes.AlreadyExists, "user with this email or username already exists")
		}
		return nil, status.Errorf(codes.Internal, "failed to create user: %v", err)
	}

	// Building the response
	user.Id = id
	user.CreatedAt = timestamppb.New(createdAt)
	user.Password = "" // do not return the password even if it is hashed
	return user, nil
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
	// Scan into type-safe fields (id,names,email,password hash,created_at)
	if err := row.Scan(
		&user.Id,
		&user.FirstName,
		&user.LastName,
		&user.UserName,
		&user.Email,
		&user.Password, // hash from DB
		&createdAt,
	); err != nil {
		if err == sql.ErrNoRows {
			return nil, status.Errorf(codes.NotFound, "user with email %s not found", email)
		}
		return nil, status.Errorf(codes.Internal, "failed to retrieve user: %v", err)
	}

	user.CreatedAt = timestamppb.New(createdAt)
	return &user, nil
}

/* ListUsers (seek pagination)  */

func (pa *PostgresAccess) listUsers(ctx context.Context, req *pb.ListUsersRequest) (*pb.ListUsersResponse, error) {
	const (
		defaultPageSize = int64(50)
		maxPageSize     = int64(1000)
	)

	// Bounds for pageSize
	ps := req.GetPageSize()
	if ps <= 0 {
		ps = defaultPageSize
	}
	if ps > maxPageSize {
		ps = maxPageSize
	}

	// Seek cursor, token: "id:<n>" or "<n>"
	var lastID int64
	tok := strings.TrimSpace(req.GetNextPageToken())
	tok = strings.TrimPrefix(tok, "id:")
	if tok != "" {
		v, err := strconv.ParseInt(tok, 10, 64)
		if err != nil || v < 0 {
			return nil, status.Error(codes.InvalidArgument, "invalid nextPageToken")
		}
		lastID = v
	}

	// WHERE from filters (CASE-INSENSITIVE) — just equals
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

	// Seek: id > lastID
	if lastID > 0 {
		where = append(where, fmt.Sprintf(`"User".id > $%d`, len(args)+1))
		args = append(args, lastID)
	}

	query := `SELECT id, first_name, last_name, user_name, email, created_at FROM "User"`
	if len(where) > 0 {
		query += " WHERE " + strings.Join(where, " AND ")
	}
	// LIMIT ps+1 to know if there is one more page
	query += fmt.Sprintf(" ORDER BY id ASC LIMIT $%d", len(args)+1)
	args = append(args, ps+1)

	rows, err := pa.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "query error: %v", err)
	}
	defer rows.Close()

	var users []*pb.User
	for rows.Next() {
		var u pb.User
		var createdAt time.Time
		if err := rows.Scan(&u.Id, &u.FirstName, &u.LastName, &u.UserName, &u.Email, &createdAt); err != nil {
			return nil, status.Errorf(codes.Internal, "scan error: %v", err)
		}
		u.CreatedAt = timestamppb.New(createdAt)
		u.Password = "" // dont expose password in list
		users = append(users, &u)
	}
	if err := rows.Err(); err != nil {
		return nil, status.Errorf(codes.Internal, "rows error: %v", err)
	}

	// 6) nextPageToken based on ps
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
