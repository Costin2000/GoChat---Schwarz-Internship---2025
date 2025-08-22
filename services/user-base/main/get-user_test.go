package main

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"os"
	"os/exec"
	"testing"

	errchecks "github.com/Costin2000/GoChat---Schwarz-Internship---2025/pkg"
	pb "github.com/Costin2000/GoChat---Schwarz-Internship---2025/services/user-base/proto"
	proto "github.com/Costin2000/GoChat---Schwarz-Internship---2025/services/user-base/proto"
	"github.com/google/go-cmp/cmp"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/testing/protocmp"
)

func fixtureGetUserRequest(mods ...func(req *pb.GetUserRequest)) *pb.GetUserRequest {
	user := &pb.GetUserRequest{
		Email: "johnwalter@yahoo.com",
	}
	for _, mod := range mods {
		mod(user)
	}
	return user
}

func Test_GetUser(t *testing.T) {
	type given struct {
		mockStorageAccess StorageAccess
	}

	tests := []struct {
		name         string
		req          *pb.GetUserRequest
		given        given
		expectedErr  errchecks.Check
		expectedResp *pb.User
	}{
		{
			name: "email is empty",
			req: fixtureGetUserRequest(func(req *pb.GetUserRequest) {
				req.Email = ""
			}),
			expectedErr: errchecks.All(errchecks.MsgContains("email cannot be empty")),
		},
		{
			name: "propagates error from getUserByEmail",
			req:  fixtureGetUserRequest(),
			given: given{
				mockStorageAccess: newMockStorageAccess(StorageMockOptions{
					getUserByEmailFunc: func(ctx context.Context, email string) (*proto.User, error) {
						return nil, errors.New("getting user failed")
					},
				}),
			},
			expectedErr: errchecks.All(errchecks.MsgContains("getting user failed")),
		},
		{
			name:         "successfully found user by email",
			req:          fixtureGetUserRequest(),
			expectedResp: fixtureUser(),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := NewMockService(ServiceMockOptions{
				storageAccess: tt.given.mockStorageAccess,
			})

			resp, err := svc.GetUser(context.Background(), tt.req)

			errchecks.Assert(t, err, tt.expectedErr)
			if diff := cmp.Diff(tt.expectedResp, resp, protocmp.Transform()); diff != "" {
				t.Errorf("mismatch (-expected +got):\n%s", diff)
			}
		})
	}
}

func TestGetUser_Integration(t *testing.T) {

	testGoodUser := &proto.User{
		FirstName: "dummy",
		LastName:  "user",
		UserName:  "testuser1",
		Email:     "test@example.com",
		Password:  "password",
	}

	testBadUser := &proto.User{
		FirstName: "dummy",
		LastName:  "user",
		UserName:  "",
		Email:     "test@example.com",
		Password:  "password",
	}

	tests := []struct {
		name         string
		inputUser    *proto.User
		request      *proto.GetUserRequest
		expectedCode codes.Code
	}{
		{
			name:         "Good request",
			inputUser:    testGoodUser,
			request:      &proto.GetUserRequest{Email: "test@example.com"},
			expectedCode: codes.OK,
		},
		{
			name:         "Bad request - no email provided",
			inputUser:    testGoodUser,
			request:      &proto.GetUserRequest{Email: ""},
			expectedCode: codes.InvalidArgument,
		},
		{
			name:         "Bad request - invalid email",
			inputUser:    testBadUser,
			request:      &proto.GetUserRequest{Email: "test@badexample.com"},
			expectedCode: codes.NotFound,
		},
	}

	startDbScript := "./../scripts/db-start.sh"
	startDbCmd := exec.Command("bash", startDbScript)
	stopDbScript := "./../scripts/db-stop.sh"
	stopDbCmd := exec.Command("bash", stopDbScript)

	err := startDbCmd.Run()
	if err != nil {
		log.Fatalf("Error executing script: %v", err)
	}

	// db connection
	envPath := "./../../../db/.env"

	if err := loadEnv(envPath); err != nil {
		log.Fatalf("Error loading .env file: %v", err)
	}

	dbUser := os.Getenv("POSTGRES_USER")
	dbPassword := os.Getenv("POSTGRES_PASSWORD")
	dbName := os.Getenv("POSTGRES_DB")
	dbPort := os.Getenv("DB_PORT")

	connStr := fmt.Sprintf("user=%s password=%s host=localhost port=%s dbname=%s sslmode=disable", dbUser, dbPassword, dbPort, dbName)

	db, err := sql.Open("pgx", connStr)
	if err != nil {
		t.Fatalf("Failed to connect to database for integration test: %v", err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		log.Fatalf("Failed to ping database: %v", err)
	}

	// server connections
	storage := newPostgresAccess(db)
	s := &UserService{storageAccess: storage}

	for _, tt := range tests {

		_, err := db.ExecContext(context.Background(), `DELETE FROM "User"`)
		if err != nil {
			t.Fatalf("Failed to clean User table: %v", err)
		}

		query := `INSERT INTO "User" (first_name, last_name, user_name, email, password) VALUES ($1, $2, $3, $4, $5)`
		_, err = db.ExecContext(context.Background(), query, tt.inputUser.FirstName, tt.inputUser.LastName, tt.inputUser.UserName, tt.inputUser.Email, tt.inputUser.Password)
		if err != nil {
			t.Fatalf("Failed to insert user for test setup: %v", err)
		}

		user, err := s.GetUser(context.Background(), tt.request)

		st, _ := status.FromError(err)
		if st.Code() != tt.expectedCode {
			t.Fatalf("Expected status code %v, got %v", tt.expectedCode, st.Code())
		}

		if tt.expectedCode == codes.OK {
			if user == nil {
				t.Fatal("Expected a user response, but got nil")
			}
			if user.Email != tt.inputUser.Email {
				t.Errorf("Expected email %s, got %s", tt.inputUser.Email, user.Email)
			}
		}

	}

	err = stopDbCmd.Run()
	if err != nil {
		log.Fatalf("Error executing script: %v", err)
	}
}
