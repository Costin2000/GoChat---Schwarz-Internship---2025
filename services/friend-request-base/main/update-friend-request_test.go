package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
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

	storage := newPostgresAccess(db)
	s := &friendRequestService{storageAccess: storage}

	// clean tables
	db.ExecContext(context.Background(), `DELETE FROM "Friend Requests"`)
	db.ExecContext(context.Background(), `DELETE FROM "User"`)

	// insert test users
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

	// create friend request
	createResp, err := storage.requestCreateFriendRequest(context.Background(), &pb.CreateFriendRequestRequest{
		SenderId:   fmt.Sprintf("%d", senderID),
		ReceiverId: fmt.Sprintf("%d", receiverID),
	})
	if err != nil {
		t.Fatalf("Failed to create friend request: %v", err)
	}

	// update friend request to "accepted"
	updateReq := &pb.UpdateFriendRequestRequest{
		FriendRequest: &pb.FriendRequest{
			Id:     createResp.Request.Id,
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

	// stop DB
	if err := stopDbCmd.Run(); err != nil {
		log.Fatalf("Error stopping DB: %v", err)
	}
}
