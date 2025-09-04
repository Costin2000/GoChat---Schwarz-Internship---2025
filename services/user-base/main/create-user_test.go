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
	"github.com/google/go-cmp/cmp"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/testing/protocmp"

	pb "github.com/Costin2000/GoChat---Schwarz-Internship---2025/services/user-base/proto"
	//Added for hash password
	"golang.org/x/crypto/bcrypt"
)

func fixtureCreateUserRequest(mods ...func(req *pb.CreateUserRequest)) *pb.CreateUserRequest {
	user := &pb.CreateUserRequest{
		User: &pb.User{
			FirstName: "John",
			LastName:  "Walter",
			UserName:  "Test",
			Email:     "johnwalter@yahoo.com",
			Password:  "secretpassword",
		},
	}
	for _, mod := range mods {
		mod(user)
	}
	return user
}

func fixtureCreateUserResponse(mods ...func(req *pb.CreateUserResponse)) *pb.CreateUserResponse {
	user := &pb.CreateUserResponse{
		User: fixtureUser(),
	}
	for _, mod := range mods {
		mod(user)
	}
	return user
}

func Test_CreateUser(t *testing.T) {
	type given struct {
		mockStorageAccess StorageAccess
	}

	tests := []struct {
		name         string
		req          *pb.CreateUserRequest
		given        given
		expectedErr  errchecks.Check
		expectedResp *pb.CreateUserResponse
	}{
		{
			name: "nil user in request",
			req: fixtureCreateUserRequest(func(req *pb.CreateUserRequest) {
				req.User = nil
			}),
			expectedErr: errchecks.All(errchecks.MsgContains("user object is required")),
		},
		{
			name: "missing fields",
			req: fixtureCreateUserRequest(func(req *pb.CreateUserRequest) {
				req.User.FirstName = ""
			}),
			expectedErr: errchecks.MsgContains("all fields are required"),
		},
		{
			name: "storage returns error",
			req:  fixtureCreateUserRequest(),
			given: given{
				mockStorageAccess: newMockStorageAccess(StorageMockOptions{
					createUserFunc: func(ctx context.Context, user *pb.User) (*pb.User, error) {
						return nil, errors.New("db insert failed")
					},
				}),
			},
			expectedErr: errchecks.MsgContains("db insert failed"),
		},
		{
			name:         "successfully creates user",
			req:          fixtureCreateUserRequest(),
			expectedResp: fixtureCreateUserResponse(),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := NewMockService(ServiceMockOptions{
				storageAccess: tt.given.mockStorageAccess,
			})

			resp, err := svc.CreateUser(context.Background(), tt.req)

			errchecks.Assert(t, err, tt.expectedErr)
			if diff := cmp.Diff(tt.expectedResp, resp, protocmp.Transform()); diff != "" {
				t.Errorf("mismatch (-expected +got):\n%s", diff)
			}
		})
	}
}

func TestCreateUser_Integration(t *testing.T) {
	testUser := &pb.User{
		FirstName: "John",
		LastName:  "Doe",
		UserName:  "johndoe",
		Email:     "johndoe@example.com",
		Password:  "password123",
	}

	startDbCmd := exec.Command("bash", "./../scripts/db-start.sh")
	stopDbCmd := exec.Command("bash", "./../scripts/db-stop.sh")

	if err := startDbCmd.Run(); err != nil {
		log.Fatalf("Error starting DB: %v", err)
	}

	// load env
	envPath := "./../../../db/.env"
	if err := loadEnv(envPath); err != nil {
		log.Fatalf("Error loading env: %v", err)
	}

	dbUser := os.Getenv("POSTGRES_USER")
	dbPassword := os.Getenv("POSTGRES_PASSWORD")
	dbName := os.Getenv("POSTGRES_DB")
	dbPort := os.Getenv("DB_PORT")

	connStr := fmt.Sprintf("user=%s password=%s host=localhost port=%s dbname=%s sslmode=disable",
		dbUser, dbPassword, dbPort, dbName)

	db, err := sql.Open("pgx", connStr)
	if err != nil {
		t.Fatalf("Failed to connect to DB: %v", err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		t.Fatalf("Failed to ping DB: %v", err)
	}

	storage := newPostgresAccess(db)
	s := &UserService{storageAccess: storage}

	// Clean table
	db.ExecContext(context.Background(), `DELETE FROM "User"`)

	//The password is stocked in 'plain' variable, because createUser will empty it for the received object
	plain := testUser.Password

	resp, err := s.CreateUser(context.Background(), &pb.CreateUserRequest{User: testUser})
	if err != nil {
		st, _ := status.FromError(err)
		t.Fatalf("Expected success, got error: %v", st.Message())
	}

	if resp.User.Email != testUser.Email {
		t.Errorf("Expected email %s, got %s", testUser.Email, resp.User.Email)
	}

	//The response should not include the password, not even the hash password
	if resp.User.Password != "" {
		t.Errorf("password must not be returned in CreateUser response")
	}

	//In DB the password should be hashed, and should coresponde with the initial password
	stored, err := storage.getUserByEmail(context.Background(), testUser.Email)
	if err != nil {
		t.Fatalf("failed to fetch user from db: %v", err)
	}
	if err := bcrypt.CompareHashAndPassword([]byte(stored.Password), []byte(plain)); err != nil {
		t.Fatalf("stored password is not a valid bcrypt hash for the provided password: %v", err)
	}

	// Stop DB
	if err := stopDbCmd.Run(); err != nil {
		log.Fatalf("Error stopping DB: %v", err)
	}
}
