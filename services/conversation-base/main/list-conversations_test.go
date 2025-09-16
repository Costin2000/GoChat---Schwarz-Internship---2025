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
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/testing/protocmp"
)

func fixtureListConversationsResponse() *pb.ListConversationsResponse {
	return &pb.ListConversationsResponse{
		Conversations: []*pb.Conversation{
			{Id: "1", User1Id: "111", User2Id: "222"},
			{Id: "2", User1Id: "333", User2Id: "444"},
		},
	}
}

func Test_ListConversations(t *testing.T) {
	successResp := fixtureListConversationsResponse()

	tests := []struct {
		name        string
		req         *pb.ListConversationsRequest
		mockFunc    func(ctx context.Context, req *pb.ListConversationsRequest) (*pb.ListConversationsResponse, error)
		expecterErr errchecks.Check
		expected    *pb.ListConversationsResponse
	}{
		{
			name: "Error - missing userId",
			req:  &pb.ListConversationsRequest{},
			mockFunc: func(ctx context.Context, req *pb.ListConversationsRequest) (*pb.ListConversationsResponse, error) {
				return nil, status.Errorf(codes.InvalidArgument, "user_id must be provided")
			},
			expecterErr: errchecks.IsInvalidArgument(nil),
			expected:    nil,
		},
		{
			name: "Happy path - filtered by user",
			req:  &pb.ListConversationsRequest{UserId: "111"},
			mockFunc: func(ctx context.Context, req *pb.ListConversationsRequest) (*pb.ListConversationsResponse, error) {
				return &pb.ListConversationsResponse{
					Conversations: []*pb.Conversation{successResp.Conversations[0]},
				}, nil
			},
			expected: &pb.ListConversationsResponse{
				Conversations: []*pb.Conversation{successResp.Conversations[0]},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := NewMockService(ServiceMockOptions{
				storageAccess: newMockStorageAccess(StorageMockOptions{
					listConversationsFunc: tt.mockFunc,
				}),
			})

			resp, err := svc.ListConversations(context.Background(), tt.req)
			errchecks.Assert(t, err, tt.expecterErr)

			if diff := cmp.Diff(tt.expected, resp, protocmp.Transform()); diff != "" {
				t.Errorf("Mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestListConversations_Integration(t *testing.T) {
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
	var user1ID, user2ID, user3ID int64
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

	err = db.QueryRow(`INSERT INTO "User" (first_name, last_name, user_name, email, password)
		VALUES ('Charlie', 'Brown', 'charlie', 'charlie@example.com', 'secret') RETURNING id`).Scan(&user3ID)
	if err != nil {
		t.Fatalf("Failed to insert user3: %v", err)
	}

	// create conversations
	_, err = svc.CreateConversation(context.Background(), &pb.CreateConversationRequest{
		User1Id: fmt.Sprintf("%d", user1ID),
		User2Id: fmt.Sprintf("%d", user2ID),
	})
	if err != nil {
		t.Fatalf("Failed to create conversation 1: %v", err)
	}

	_, err = svc.CreateConversation(context.Background(), &pb.CreateConversationRequest{
		User1Id: fmt.Sprintf("%d", user2ID),
		User2Id: fmt.Sprintf("%d", user3ID),
	})
	if err != nil {
		t.Fatalf("Failed to create conversation 2: %v", err)
	}

	// Test 1: Filter by user2 (Bob)
	filterResp, err := svc.ListConversations(context.Background(), &pb.ListConversationsRequest{
		UserId: fmt.Sprintf("%d", user2ID),
	})
	if err != nil {
		t.Fatalf("ListConversations (filter) failed: %v", err)
	}
	if len(filterResp.Conversations) != 2 {
		t.Errorf("Expected 2 conversations for user %d, got %d", user2ID, len(filterResp.Conversations))
	}

	// Test 2: Filter by user1 (Alice)
	aliceResp, err := svc.ListConversations(context.Background(), &pb.ListConversationsRequest{
		UserId: fmt.Sprintf("%d", user1ID),
	})
	if err != nil {
		t.Fatalf("ListConversations (Alice) failed: %v", err)
	}
	if len(aliceResp.Conversations) != 1 {
		t.Errorf("Expected 1 conversation for Alice, got %d", len(aliceResp.Conversations))
	}

	// cleanup
	db.ExecContext(context.Background(), `DELETE FROM "Conversation"`)
	db.ExecContext(context.Background(), `DELETE FROM "User"`)
}
