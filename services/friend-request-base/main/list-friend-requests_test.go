package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"
	"os/exec"
	"testing"
	"time"

	errchecks "github.com/Costin2000/GoChat---Schwarz-Internship---2025/pkg"
	frproto "github.com/Costin2000/GoChat---Schwarz-Internship---2025/services/friend-request-base/proto"
	"github.com/google/go-cmp/cmp"
	"golang.org/x/crypto/bcrypt"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/testing/protocmp"
)

const (
	maxIntTestUsers  = 4
	maxUnitTestUsers = 20
)

func fixtureListFriendRequestsRequest(mods ...func(req *frproto.ListFriendRequestsRequest)) *frproto.ListFriendRequestsRequest {
	req := &frproto.ListFriendRequestsRequest{
		PageSize: 20,
		Filters: []*frproto.ListFriendRequestsFiltersOneOf{
			{
				Filter: &frproto.ListFriendRequestsFiltersOneOf_ReceiverId{ReceiverId: "222"},
			},
		},
	}
	for _, mod := range mods {
		mod(req)
	}
	return req
}

func fixtureListFriendRequestsResponse(requests []*frproto.FriendRequest, nextPageToken string) *frproto.ListFriendRequestsResponse {
	return &frproto.ListFriendRequestsResponse{
		Requests:      requests,
		NextPageToken: nextPageToken,
	}
}

func TestListFriendRequests_Integration(t *testing.T) {
	startDbCmd := exec.Command("bash", "./../scripts/db-start.sh")
	stopDbCmd := exec.Command("bash", "./../scripts/db-stop.sh")

	defer stopDbCmd.Run()

	if err := startDbCmd.Run(); err != nil {
		log.Fatalf("Error starting DB: %v", err)
	}

	time.Sleep(2 * time.Second)

	// load env
	envPath := "./../../../devtest-db/.env"
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

	frStorage := newPostgresAccess(db)
	frSvc := &friendRequestService{storageAccess: frStorage}

	createdUserIDs := make(map[string]string)
	dummyPassword, _ := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)
	query := `
			INSERT INTO "User" (first_name, last_name, user_name, email, password)
			VALUES ($1, $2, $3, $4, $5)
			RETURNING id;
			`
	for i := 1; i <= maxIntTestUsers; i++ {
		key := fmt.Sprintf("user%d", i)
		var userId string
		err := db.QueryRowContext(context.Background(), query,
			"Dummy",
			fmt.Sprintf("Test%d", i),
			fmt.Sprintf("User%d", i),
			fmt.Sprintf("dummy%d@test.com", i),
			string(dummyPassword),
		).Scan(&userId)

		if err != nil {
			t.Fatalf("Failed to create %s: %v", key, err)
		}
		createdUserIDs[key] = userId
	}
	for i := 1; i <= maxIntTestUsers; i++ {
		for j := i + 1; j <= maxIntTestUsers; j++ {
			frSvc.CreateFriendRequest(context.Background(), &frproto.CreateFriendRequestRequest{
				SenderId:   createdUserIDs[fmt.Sprintf("user%d", i)],
				ReceiverId: createdUserIDs[fmt.Sprintf("user%d", j)],
			})
		}
	}

	senderIDToTest := createdUserIDs["user1"]
	expectedCount := maxIntTestUsers - 1

	listFrReq := &frproto.ListFriendRequestsRequest{
		NextPageToken: "",
		PageSize:      defaultPageSize,
		Filters:       []*frproto.ListFriendRequestsFiltersOneOf{{Filter: &frproto.ListFriendRequestsFiltersOneOf_SenderId{SenderId: senderIDToTest}}},
	}

	listFrRsp, err := frSvc.ListFriendRequests(context.Background(), listFrReq)

	if err != nil {
		t.Fatalf("ListFriendRequests returned an unexpected error: %v", err)
	}

	if listFrRsp == nil {
		t.Fatal("Expected a response but got nil")
	}

	if len(listFrRsp.Requests) != expectedCount {
		t.Errorf("Expected %d friend requests for sender %s, but got %d", expectedCount, senderIDToTest, len(listFrRsp.Requests))
	}

	validReceiverIDs := make(map[string]struct{})
	for key, id := range createdUserIDs {
		if key != "user1" {
			validReceiverIDs[id] = struct{}{}
		}
	}

	for _, req := range listFrRsp.Requests {
		if req.SenderId != senderIDToTest {
			t.Errorf("Expected sender ID to be %s, but got %s", senderIDToTest, req.SenderId)
		}

		if _, ok := validReceiverIDs[req.ReceiverId]; !ok {
			t.Errorf("Received an unexpected receiver ID: %s. Valid receiver IDs are: %v", req.ReceiverId, validReceiverIDs)
		}
	}
}

func TestListFriendRequests_Unit(t *testing.T) {
	//  Test Setup
	sampleReq1 := fixtureCreateFriendRequestResponse().Request
	sampleReq2 := fixtureCreateFriendRequestResponse(func(fr *frproto.FriendRequest) {
		fr.Id = "2"
		fr.SenderId = "333"
		fr.ReceiverId = "222"
	}).Request

	successfulResponse := fixtureListFriendRequestsResponse([]*frproto.FriendRequest{sampleReq1, sampleReq2}, "2")
	storageError := status.Error(codes.Internal, "database is down")

	// Test Scenarios
	type given struct {
		mockStorageAccess StorageAccess
	}
	type expected struct {
		resp            *frproto.ListFriendRequestsResponse
		err             errchecks.Check
		storagePageSize int64
	}

	tests := []struct {
		name     string
		req      *frproto.ListFriendRequestsRequest
		given    given
		expected expected
	}{
		{
			name: "happy path - should list friend requests successfully",
			req:  fixtureListFriendRequestsRequest(), // uses default PageSize of 20
			given: given{
				mockStorageAccess: newMockStorageAccess(StorageMockOptions{
					listFriendRequestsFunc: func(ctx context.Context, req *frproto.ListFriendRequestsRequest) (*frproto.ListFriendRequestsResponse, error) {
						return successfulResponse, nil
					},
				}),
			},
			expected: expected{
				resp:            successfulResponse,
				err:             nil,
				storagePageSize: 20, // should remain unchanged
			},
		},
		{
			name: "page size less than default - should return InvalidArgument",
			req: fixtureListFriendRequestsRequest(func(req *frproto.ListFriendRequestsRequest) {
				req.PageSize = 5 // less than defaultPageSize
			}),
			expected: expected{
				resp:            nil,
				err:             errchecks.HasStatusCode(codes.InvalidArgument),
				storagePageSize: defaultPageSize, // expect adjustment to 10
			},
		},
		{
			name: "bad page size",
			req: fixtureListFriendRequestsRequest(func(req *frproto.ListFriendRequestsRequest) {
				req.PageSize = -1 // Invalid page size
			}),
			expected: expected{
				resp:            nil,
				err:             errchecks.HasStatusCode(codes.InvalidArgument),
				storagePageSize: defaultPageSize, // expect adjustment to 10
			},
		},
		{
			name: "page size greater than max",
			req: fixtureListFriendRequestsRequest(func(req *frproto.ListFriendRequestsRequest) {
				req.PageSize = 200 // greater than maxPageSize
			}),
			given: given{
				mockStorageAccess: newMockStorageAccess(StorageMockOptions{
					listFriendRequestsFunc: func(ctx context.Context, req *frproto.ListFriendRequestsRequest) (*frproto.ListFriendRequestsResponse, error) {
						return successfulResponse, nil
					},
				}),
			},
			expected: expected{
				resp:            successfulResponse,
				err:             nil,
				storagePageSize: maxPageSize, // expect adjustment to max
			},
		},
		{
			name: "storage layer returns an error - should propagate the error",
			req:  fixtureListFriendRequestsRequest(),
			given: given{
				mockStorageAccess: newMockStorageAccess(StorageMockOptions{
					listFriendRequestsFunc: func(ctx context.Context, req *frproto.ListFriendRequestsRequest) (*frproto.ListFriendRequestsResponse, error) {
						return nil, storageError
					},
				}),
			},
			expected: expected{
				resp:            nil,
				err:             errchecks.Is(storageError),
				storagePageSize: 20, // Page size is adjusted before the call
			},
		},
	}

	// Test Execution
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			svc := NewMockService(ServiceMockOptions{
				storageAccess: tt.given.mockStorageAccess,
			})

			resp, err := svc.ListFriendRequests(context.Background(), tt.req)

			errchecks.Assert(t, err, tt.expected.err)

			if diff := cmp.Diff(tt.expected.resp, resp, protocmp.Transform()); diff != "" {
				t.Errorf("ListFriendRequests response mismatch (-expected +got):\n%s", diff)
			}

		})
	}
}
