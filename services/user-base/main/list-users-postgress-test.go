// Storage/SQL unit tests with sqlmock
package main

import (
	"context"
	"fmt"
	"regexp"
	"testing"
	"time"

	errchecks "github.com/Costin2000/GoChat---Schwarz-Internship---2025/pkg"
	pb "github.com/Costin2000/GoChat---Schwarz-Internship---2025/services/user-base/proto"
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/go-cmp/cmp"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/testing/protocmp"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func q(s string) string { return regexp.QuoteMeta(s) }

func Test_ListUsers_Postgres(t *testing.T) {
	now := time.Now().UTC().Truncate(time.Second)

	type given struct {
		mock func(sqlmock.Sqlmock)
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
			name: "happy path — first page, no filters",
			req:  &pb.ListUsersRequest{},
			given: given{
				mock: func(m sqlmock.Sqlmock) {
					m.ExpectQuery(q(`SELECT id, first_name, last_name, user_name, email, created_at FROM "User" ORDER BY id ASC LIMIT $1`)).
						WithArgs(int64(51)).
						WillReturnRows(
							sqlmock.NewRows([]string{"id", "first_name", "last_name", "user_name", "email", "created_at"}).
								AddRow(int64(1), "Ana", "Ionescu", "ana", "ana@example.com", now).
								AddRow(int64(2), "Ion", "Popescu", "ion", "ion@example.com", now),
						)
				},
			},
			want: want{
				resp: &pb.ListUsersResponse{
					NextPageToken: "",
					Users: []*pb.User{
						{Id: 1, FirstName: "Ana", LastName: "Ionescu", UserName: "ana", Email: "ana@example.com", CreatedAt: timestamppb.New(now)},
						{Id: 2, FirstName: "Ion", LastName: "Popescu", UserName: "ion", Email: "ion@example.com", CreatedAt: timestamppb.New(now)},
					},
				},
				err: nil,
			},
		},
		{
			name: "happy path — filters equals + seek token",
			req: &pb.ListUsersRequest{
				PageSize:      2,
				NextPageToken: "id:10",
				Filters: []*pb.ListUsersFiltersOneOf{
					{Filter: &pb.ListUsersFiltersOneOf_FirstName{FirstName: &pb.FilterByFirstName{Equals: "Ana"}}},
					{Filter: &pb.ListUsersFiltersOneOf_LastName{LastName: &pb.FilterByLastName{Equals: "Ionescu"}}},
				},
			},
			given: given{
				mock: func(m sqlmock.Sqlmock) {
					// LIMIT ps+1 = 3
					m.ExpectQuery(q(`SELECT id, first_name, last_name, user_name, email, created_at FROM "User" WHERE LOWER(first_name) = LOWER($1) AND LOWER(last_name) = LOWER($2) AND "User".id > $3 ORDER BY id ASC LIMIT $4`)).
						WithArgs("Ana", "Ionescu", int64(10), int64(3)).
						WillReturnRows(
							sqlmock.NewRows([]string{"id", "first_name", "last_name", "user_name", "email", "created_at"}).
								AddRow(int64(11), "Ana", "Ionescu", "ana", "ana@example.com", now).
								AddRow(int64(12), "Ana", "Ionescu", "ana2", "ana2@example.com", now).
								AddRow(int64(13), "Ana", "Ionescu", "ana3", "ana3@example.com", now), // extra pt. nextToken
						)
				},
			},
			want: want{
				resp: &pb.ListUsersResponse{
					NextPageToken: "id:12", // The last included from the page
					Users: []*pb.User{
						{Id: 11, FirstName: "Ana", LastName: "Ionescu", UserName: "ana", Email: "ana@example.com", CreatedAt: timestamppb.New(now)},
						{Id: 12, FirstName: "Ana", LastName: "Ionescu", UserName: "ana2", Email: "ana2@example.com", CreatedAt: timestamppb.New(now)},
					},
				},
				err: nil,
			},
		},
		{
			name: "invalid token — returns InvalidArgument, no DB hit",
			req: &pb.ListUsersRequest{
				PageSize:      10,
				NextPageToken: "id:not-a-number",
			},
			given: given{
				mock: func(m sqlmock.Sqlmock) {

				},
			},
			want: want{
				resp: nil,
				err:  errchecks.HasStatusCode(codes.InvalidArgument),
			},
		},
		{
			name: "DB error bubbles up as Internal",
			req:  &pb.ListUsersRequest{},
			given: given{
				mock: func(m sqlmock.Sqlmock) {
					m.ExpectQuery(q(`SELECT id, first_name, last_name, user_name, email, created_at FROM "User" ORDER BY id ASC LIMIT $1`)).
						WithArgs(int64(51)).
						WillReturnError(fmt.Errorf("db down"))
				},
			},
			want: want{
				resp: nil,
				err:  errchecks.HasStatusCode(codes.Internal),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// special case: invalid token test
			if tt.name == "invalid token — returns InvalidArgument, no DB hit" {
				store := &PostgresAccess{db: nil}
				rsp, err := store.listUsers(context.Background(), tt.req)
				errchecks.Assert(t, err, tt.want.err)
				if diff := cmp.Diff(tt.want.resp, rsp, protocmp.Transform()); diff != "" {
					t.Errorf("mismatch (-want +got):\n%s", diff)
				}
				return
			}

			db, mock, err := sqlmock.New()
			if err != nil {
				t.Fatalf("sqlmock: %v", err)
			}
			defer db.Close()

			if tt.given.mock != nil {
				tt.given.mock(mock)
			}

			store := &PostgresAccess{db: db}
			rsp, err := store.listUsers(context.Background(), tt.req)

			errchecks.Assert(t, err, tt.want.err)

			// Passwords should not be expaused in lists
			if rsp != nil {
				for i, u := range rsp.Users {
					if u.Password != "" {
						t.Fatalf("user[%d].Password should be empty in list response", i)
					}
				}
			}

			if diff := cmp.Diff(tt.want.resp, rsp, protocmp.Transform()); diff != "" {
				t.Errorf("mismatch (-want +got):\n%s", diff)
			}

			if err := mock.ExpectationsWereMet(); err != nil {
				if status.Code(err) != codes.OK {
					t.Fatalf("unmet expectations: %v", err)
				}
			}
		})
	}
}
