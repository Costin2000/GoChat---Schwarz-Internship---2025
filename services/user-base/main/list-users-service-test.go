// gRPC handler unit tests
package main

import (
	"context"
	"testing"

	errchecks "github.com/Costin2000/GoChat---Schwarz-Internship---2025/pkg"
	pb "github.com/Costin2000/GoChat---Schwarz-Internship---2025/services/user-base/proto"
	"github.com/google/go-cmp/cmp"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/testing/protocmp"
)

func fixtureListUsersResponse(mods ...func(*pb.ListUsersResponse)) *pb.ListUsersResponse {
	r := &pb.ListUsersResponse{
		NextPageToken: "id:42",
		Users: []*pb.User{
			{Id: 41, FirstName: "Ana", LastName: "Ionescu"},
			{Id: 42, FirstName: "Ion", LastName: "Popescu"},
		},
	}
	for _, m := range mods {
		m(r)
	}
	return r
}

func Test_ListUsers_Service(t *testing.T) {
	type given struct {
		mockStorage StorageAccess
	}

	type want struct {
		resp *pb.ListUsersResponse
		err  errchecks.Check
	}

	tests := []struct {
		name  string
		req   *pb.ListUsersRequest
		given given
		want  want
	}{
		{
			name: "invalid pageSize (<0) -> InvalidArgument",
			req:  &pb.ListUsersRequest{PageSize: -1},
			given: given{
				mockStorage: newMockStorageAccess(StorageMockOptions{
					// this will not be called, the handler will be validated before
				}),
			},
			want: want{
				resp: nil,
				err:  errchecks.HasStatusCode(codes.InvalidArgument),
			},
		},
		{
			name: "happy path — delegates to storage",
			req:  &pb.ListUsersRequest{PageSize: 2},
			given: given{
				mockStorage: newMockStorageAccess(StorageMockOptions{
					listUsersFunc: func(ctx context.Context, req *pb.ListUsersRequest) (*pb.ListUsersResponse, error) {
						return fixtureListUsersResponse(), nil
					},
				}),
			},
			want: want{
				resp: fixtureListUsersResponse(),
				err:  nil,
			},
		},
		{
			name: "storage returns Internal -> bubbled up",
			req:  &pb.ListUsersRequest{PageSize: 2},
			given: given{
				mockStorage: newMockStorageAccess(StorageMockOptions{
					listUsersFunc: func(ctx context.Context, req *pb.ListUsersRequest) (*pb.ListUsersResponse, error) {
						return nil, status.Error(codes.Internal, "boom")
					},
				}),
			},
			want: want{
				resp: nil,
				err:  errchecks.HasStatusCode(codes.Internal),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := NewMockService(ServiceMockOptions{
				storageAccess: tt.given.mockStorage,
			})

			rsp, err := svc.ListUsers(context.Background(), tt.req)

			errchecks.Assert(t, err, tt.want.err)
			if diff := cmp.Diff(tt.want.resp, rsp, protocmp.Transform()); diff != "" {
				t.Errorf("mismatch (-want +got):\n%s", diff)
			}
		})
	}
}
