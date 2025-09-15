package main

import (
	"context"
	"testing"

	errchecks "github.com/Costin2000/GoChat---Schwarz-Internship---2025/pkg"
	pb "github.com/Costin2000/GoChat---Schwarz-Internship---2025/services/message-base/proto"
	"github.com/google/go-cmp/cmp"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/testing/protocmp"
)

func Test_listMessages_Unit(t *testing.T) {

	type given struct {
		mockStorage StorageAccess
	}
	type want struct {
		rsp *pb.ListMessagesResponse
		err errchecks.Check
	}

	tests := []struct {
		name  string
		req   *pb.ListMessagesRequest
		given given
		want  want
	}{
		{
			name: "Succes, happy path: returns messages",
			req: &pb.ListMessagesRequest{
				Filter: &pb.ListMessagesFilter{ConversationId: 1},
			},
			given: given{
				mockStorage: newMockStorageAccess(StorageMockOptions{
					ListMessagesFunc: func(ctx context.Context, req *pb.ListMessagesRequest) (*pb.ListMessagesResponse, error) {
						return &pb.ListMessagesResponse{Messages: []*pb.Message{
							{Id: 1, ConversationId: 1, SenderId: 1, Content: "This is some message"},
							{Id: 2, ConversationId: 1, SenderId: 2, Content: "This is another message"},
						}}, nil
					},
				}),
			},
			want: want{
				rsp: &pb.ListMessagesResponse{Messages: []*pb.Message{
					{Id: 1, ConversationId: 1, SenderId: 1, Content: "This is some message"},
					{Id: 2, ConversationId: 1, SenderId: 2, Content: "This is another message"},
				}},
				err: errchecks.IsNil,
			},
		},
		{
			name: "Succes: no messages or conversation doesn't exist",
			req: &pb.ListMessagesRequest{
				Filter: &pb.ListMessagesFilter{ConversationId: 2},
			},
			given: given{
				mockStorage: newMockStorageAccess(StorageMockOptions{
					ListMessagesFunc: func(ctx context.Context, req *pb.ListMessagesRequest) (*pb.ListMessagesResponse, error) {
						return &pb.ListMessagesResponse{Messages: []*pb.Message{}}, nil
					},
				}),
			},
			want: want{
				rsp: &pb.ListMessagesResponse{Messages: []*pb.Message{}},
				err: errchecks.IsNil,
			},
		},
		{
			name: "Failure: invalid conversation ID",
			req: &pb.ListMessagesRequest{
				Filter: &pb.ListMessagesFilter{ConversationId: 0},
			},
			given: given{
				mockStorage: newMockStorageAccess(StorageMockOptions{
					ListMessagesFunc: func(ctx context.Context, req *pb.ListMessagesRequest) (*pb.ListMessagesResponse, error) {
						// this should never be called
						t.Fatal("storageAccess.listMessages should not be called for invalid input")
						return nil, nil
					},
				}),
			},
			want: want{
				rsp: nil,
				err: errchecks.MsgContains("Invalid Conversation ID"),
			},
		},
		{
			name: "Failure: storage layer returns an error",
			req: &pb.ListMessagesRequest{
				Filter: &pb.ListMessagesFilter{ConversationId: 1},
			},
			given: given{
				mockStorage: newMockStorageAccess(StorageMockOptions{
					ListMessagesFunc: func(ctx context.Context, req *pb.ListMessagesRequest) (*pb.ListMessagesResponse, error) {
						return nil, status.Error(codes.Internal, "database connection failed")
					},
				}),
			},
			want: want{
				rsp: nil,
				err: errchecks.Is(status.Error(codes.Internal, "database connection failed")),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := &MessageService{
				storageAccess: tt.given.mockStorage,
			}
			gotRsp, err := svc.ListMessages(context.Background(), tt.req)

			errchecks.Assert(t, err, tt.want.err)
			if diff := cmp.Diff(tt.want.rsp, gotRsp, protocmp.Transform()); diff != "" {
				t.Errorf("ListMessages() response mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

// func Test_listMessages_Integration(t *testing.T) {
// 	startDbCmd := exec.Command("bash", "./../scripts/db-start.sh")
// 	stopDbCmd := exec.Command("bash", "./../scripts/db-stop.sh")

// 	defer stopDbCmd.Run()

// 	if err := startDbCmd.Run(); err != nil {
// 		log.Fatalf("Error starting DB: %v", err)
// 	}

// 	time.Sleep(2 * time.Second)

// 	// load env
// 	envPath := "./../../../devtest-db/.env"
// 	if err := loadEnv(envPath); err != nil {
// 		log.Fatalf("Error loading env: %v", err)
// 	}

// 	dbUser := os.Getenv("POSTGRES_USER")
// 	dbPassword := os.Getenv("POSTGRES_PASSWORD")
// 	dbName := os.Getenv("POSTGRES_DB")
// 	dbPort := os.Getenv("DB_PORT")

// 	connStr := fmt.Sprintf("user=%s password=%s host=localhost port=%s dbname=%s sslmode=disable",
// 		dbUser, dbPassword, dbPort, dbName)

// 	db, err := sql.Open("pgx", connStr)
// 	if err != nil {
// 		t.Fatalf("Failed to connect to DB: %v", err)
// 	}
// 	defer db.Close()

// 	if err := db.Ping(); err != nil {
// 		t.Fatalf("Failed to ping DB: %v", err)
// 	}

// 	msgStorage := newPostgresAccess(db)
// 	msgSvc := &MessageService{storageAccess: msgStorage}

// 	query := `
// 			INSERT INTO "
// 		`
// }
