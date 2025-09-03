package main

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"os/exec"
	"testing"

	errchecks "github.com/Costin2000/GoChat---Schwarz-Internship---2025/pkg"
	pb "github.com/Costin2000/GoChat---Schwarz-Internship---2025/services/friend-request-base/proto"
	"github.com/google/go-cmp/cmp"
	"google.golang.org/protobuf/testing/protocmp"
	"google.golang.org/protobuf/types/known/fieldmaskpb"
)

func fixtureUpdateFriendRequest(mods ...func(req *pb.UpdateFriendRequestRequest)) *pb.UpdateFriendRequestRequest {
	req := &pb.UpdateFriendRequestRequest{
		FriendRequest: &pb.FriendRequest{
			Id:         "1",
			SenderId:   "111",
			ReceiverId: "222",
			Status:     pb.RequestStatus_STATUS_ACCEPTED,
		},
		FieldMask: &fieldmaskpb.FieldMask{Paths: []string{"status"}},
	}

	for _, mod := range mods {
		mod(req)
	}

	return req
}

func Test_UpdateFriendRequest(t *testing.T) {
	type given struct {
		mockStorageAccess StorageAccess
	}

	successfulResponse := fixtureUpdateFriendResponse()

	tests := []struct {
		name         string
		req          *pb.UpdateFriendRequestRequest
		given        given
		expecterErr  errchecks.Check
		expectedResp *pb.UpdateFriendRequestResponse
	}{
		{
			name: "Nil friend request object",
			req: fixtureUpdateFriendRequest(func(req *pb.UpdateFriendRequestRequest) {
				req.FriendRequest = nil
			}),
			expecterErr: errchecks.MsgContains("friend request id must be provided"),
		},
		{
			name: "Empty friend request ID",
			req: fixtureUpdateFriendRequest(func(req *pb.UpdateFriendRequestRequest) {
				req.FriendRequest.Id = ""
			}),
			expecterErr: errchecks.MsgContains("friend request id must be provided"),
		},
		{
			name: "Empty field mask",
			req: fixtureUpdateFriendRequest(func(req *pb.UpdateFriendRequestRequest) {
				req.FieldMask.Paths = []string{}
			}),
			expecterErr: errchecks.MsgContains("at least one field must be specified in field mask"),
		},
		{
			name: "happy path - should update friend request successfully",
			req:  fixtureUpdateFriendRequest(),
			given: given{
				mockStorageAccess: newMockStorageAccess(StorageMockOptions{
					updateFriendRequestFunc: func(ctx context.Context, req *pb.UpdateFriendRequestRequest) (*pb.UpdateFriendRequestResponse, error) {
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

			rsp, err := svc.UpdateFriendRequest(context.Background(), tt.req)

			errchecks.Assert(t, err, tt.expecterErr)
			if diff := cmp.Diff(tt.expectedResp, rsp, protocmp.Transform()); diff != "" {
				t.Errorf("mismatch (-expected +got):\n%s", diff)
			}
		})
	}
}

func TestUpdateFriendRequest_Integration(t *testing.T) {
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
	envPath := "./../../../db/.env"
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

	// cleanup before test
	db.ExecContext(context.Background(), `DELETE FROM "Friend Requests"`)
	db.ExecContext(context.Background(), `DELETE FROM "User"`)

	storage := newPostgresAccess(db)
	s := &friendRequestService{storageAccess: storage}

	// insert dummy users
	var senderID, receiverID int64
	err = db.QueryRow(`INSERT INTO "User" (first_name, last_name, user_name, email, password) 
		VALUES ('John', 'Doe', 'johndoe', 'johndoe@example.com', 'secret') RETURNING id`).Scan(&senderID)
	if err != nil {
		t.Fatalf("Failed to insert sender: %v", err)
	}

	err = db.QueryRow(`INSERT INTO "User" (first_name, last_name, user_name, email, password) 
		VALUES ('Jane', 'Smith', 'janesmith', 'jane@example.com', 'secret') RETURNING id`).Scan(&receiverID)
	if err != nil {
		t.Fatalf("Failed to insert receiver: %v", err)
	}

	// insert friend request directly in DB
	var friendRequestID int64
	err = db.QueryRow(`INSERT INTO "Friend Requests" (sender_id, receiver_id, status, created_at) 
		VALUES ($1, $2, $3, NOW()) RETURNING id`, senderID, receiverID, "pending").Scan(&friendRequestID)
	if err != nil {
		t.Fatalf("Failed to insert friend request: %v", err)
	}

	// call UpdateFriendRequest to update status to ACCEPTED
	updateReq := &pb.UpdateFriendRequestRequest{
		FriendRequest: &pb.FriendRequest{
			Id:     fmt.Sprintf("%d", friendRequestID),
			Status: pb.RequestStatus_STATUS_ACCEPTED,
		},
		FieldMask: &fieldmaskpb.FieldMask{Paths: []string{"status"}},
	}

	updateResp, err := s.UpdateFriendRequest(context.Background(), updateReq)
	if err != nil {
		t.Fatalf("UpdateFriendRequest failed: %v", err)
	}

	if updateResp.FriendRequest.Status != pb.RequestStatus_STATUS_ACCEPTED {
		t.Errorf("Expected status 'ACCEPTED', got %v", updateResp.FriendRequest.Status)
	}

	// validate in DB
	var status string
	err = db.QueryRow(`SELECT status FROM "Friend Requests" WHERE id = $1`, friendRequestID).Scan(&status)
	if err != nil {
		t.Fatalf("Failed to query updated friend request: %v", err)
	}
	if status != "accepted" {
		t.Errorf("Expected DB status 'accepted', got %s", status)
	}

	// cleanup after test
	db.ExecContext(context.Background(), `DELETE FROM "Friend Requests"`)
	db.ExecContext(context.Background(), `DELETE FROM "User"`)
}
