package main

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"os/exec"
	"testing"

	errchecks "github.com/Costin2000/GoChat---Schwarz-Internship---2025/pkg"
	pb "github.com/Costin2000/GoChat---Schwarz-Internship---2025/services/conversation-base/proto"
	"github.com/google/go-cmp/cmp"
	"google.golang.org/protobuf/testing/protocmp"
)

func fixtureCreateConversationRequest(mods ...func(*pb.CreateConversationRequest)) *pb.CreateConversationRequest {
	req := &pb.CreateConversationRequest{
		User1Id: "111",
		User2Id: "222",
	}

	for _, mod := range mods {
		mod(req)
	}

	return req
}

func fixtureCreateConversationResp(mods ...func(*pb.CreateConversationResponse)) *pb.CreateConversationResponse {
	resp := fixtureCreateConversationResponse()

	for _, mod := range mods {
		mod(resp)
	}

	return resp
}

func Test_CreateConversation(t *testing.T) {
	type given struct {
		mockStorageAccess StorageAccess
	}

	successfulResponse := fixtureCreateConversationResp()

	tests := []struct {
		name         string
		req          *pb.CreateConversationRequest
		given        given
		expecterErr  errchecks.Check
		expectedResp *pb.CreateConversationResponse
	}{
		{
			name: "Empty user1 ID",
			req: fixtureCreateConversationRequest(func(req *pb.CreateConversationRequest) {
				req.User1Id = ""
			}),
			expecterErr: errchecks.MsgContains("user1 and user2 IDs cannot be empty"),
		},
		{
			name: "Empty user2 ID",
			req: fixtureCreateConversationRequest(func(req *pb.CreateConversationRequest) {
				req.User2Id = ""
			}),
			expecterErr: errchecks.MsgContains("user1 and user2 IDs cannot be empty"),
		},
		{
			name: "Same user for both IDs",
			req: fixtureCreateConversationRequest(func(req *pb.CreateConversationRequest) {
				req.User2Id = req.User1Id
			}),
			expecterErr: errchecks.MsgContains("cannot create conversation with the same user"),
		},
		{
			name: "Happy path - should create conversation successfully",
			req:  fixtureCreateConversationRequest(),
			given: given{
				mockStorageAccess: newMockStorageAccess(StorageMockOptions{
					createConversationFunc: func(ctx context.Context, req *pb.CreateConversationRequest) (*pb.CreateConversationResponse, error) {
						return successfulResponse, nil
					},
				}),
			},
			expecterErr:  nil,
			expectedResp: successfulResponse,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := NewMockService(ServiceMockOptions{
				storageAccess: tt.given.mockStorageAccess,
			})

			rsp, err := svc.CreateConversation(context.Background(), tt.req)

			errchecks.Assert(t, err, tt.expecterErr)
			if diff := cmp.Diff(tt.expectedResp, rsp, protocmp.Transform()); diff != "" {
				t.Errorf("Mismatch (-expected +got):\n%s", diff)
			}
		})
	}
}

func TestCreateConversation_Integration(t *testing.T) {
	startDbCmd := exec.Command("bash", "./../scripts/db-start.sh")
	stopDbCmd := exec.Command("bash", "./../scripts/db-stop.sh")

	if err := startDbCmd.Run(); err != nil {
		t.Fatalf("Error starting DB: %v", err)
	}
	defer func() {
		if err := stopDbCmd.Run(); err != nil {
			t.Fatalf("Error stopping DB: %v", err)
		}
	}()

	// load env
	envPath := "./../../../devtest-db/.env"
	if err := loadEnv(envPath); err != nil {
		t.Fatalf("Error loading env: %v", err)
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

	// cleanup
	db.ExecContext(context.Background(), `DELETE FROM "Conversation"`)
	db.ExecContext(context.Background(), `DELETE FROM "User"`)

	storage := newPostgresAccess(db)
	svc := &conversationService{storageAccess: storage}

	// insert dummy users
	var user1ID, user2ID int64
	err = db.QueryRow(`INSERT INTO "User" (first_name, last_name, user_name, email, password)
		VALUES ('Alice', 'Wonder', 'alice', 'alice@example.com', 'secret') RETURNING id`).Scan(&user1ID)
	if err != nil {
		t.Fatalf("Failed to insert user1: %v", err)
	}

	err = db.QueryRow(`INSERT INTO "User" (first_name, last_name, user_name, email, password)
		VALUES ('Bob', 'Builder', 'bob', 'bob@example.com', 'secret') RETURNING id`).Scan(&user2ID)
	if err != nil {
		t.Fatalf("Failed to insert user2: %v", err)
	}

	// create conversation
	req := &pb.CreateConversationRequest{
		User1Id: fmt.Sprintf("%d", user1ID),
		User2Id: fmt.Sprintf("%d", user2ID),
	}

	resp, err := svc.CreateConversation(context.Background(), req)
	if err != nil {
		t.Fatalf("CreateConversation failed: %v", err)
	}

	if resp.Conversation.User1Id != fmt.Sprintf("%d", user1ID) || resp.Conversation.User2Id != fmt.Sprintf("%d", user2ID) {
		t.Errorf("Expected conversation between %d and %d, got %+v", user1ID, user2ID, resp.Conversation)
	}

	// validate in DB
	var count int
	err = db.QueryRow(`SELECT COUNT(*) FROM "Conversation" WHERE user1_id=$1 AND user2_id=$2`, user1ID, user2ID).Scan(&count)
	if err != nil {
		t.Fatalf("Failed to validate conversation in DB: %v", err)
	}
	if count != 1 {
		t.Errorf("Expected 1 conversation, found %d", count)
	}

	// cleanup
	db.ExecContext(context.Background(), `DELETE FROM "Conversation"`)
	db.ExecContext(context.Background(), `DELETE FROM "User"`)
}
