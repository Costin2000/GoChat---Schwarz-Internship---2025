package main

import (
	"context"
	"testing"
	"time"

	proto "github.com/Costin2000/GoChat---Schwarz-Internship---2025/services/auth/proto"
	userbasepb "github.com/Costin2000/GoChat---Schwarz-Internship---2025/services/user-base/proto"
	"golang.org/x/crypto/bcrypt"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type mockUserBaseClient struct {
	userToReturn *userbasepb.User
	errToReturn  error
}

func (m *mockUserBaseClient) GetUser(ctx context.Context, in *userbasepb.GetUserRequest, opts ...grpc.CallOption) (*userbasepb.User, error) {
	return m.userToReturn, m.errToReturn
}

func TestLogin(t *testing.T) {
	password := "right-password"
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)

	t.Setenv("AUTH_JWT_SECRET", "secret-test")

	testCases := []struct {
		name          string
		email         string
		password      string
		mockUser      *userbasepb.User
		mockError     error
		expectedError codes.Code
	}{
		{
			name:     "Login succesful!",
			email:    "test@exemple.com",
			password: "right-password",
			mockUser: &userbasepb.User{
				Id:        1,
				Email:     "test@exemple.com",
				Password:  string(hashedPassword),
				CreatedAt: timestamppb.New(time.Now()),
			},
			mockError:     nil,
			expectedError: codes.OK,
		},
		{
			name:          "User not found",
			email:         "usernotfound@exemplu.com",
			password:      "pass",
			mockUser:      nil,
			mockError:     status.Error(codes.NotFound, "user not found"),
			expectedError: codes.NotFound,
		},
		{
			name:     "Wrong password",
			email:    "test@exemple.com",
			password: "wrong-password",
			mockUser: &userbasepb.User{
				Id:       1,
				Email:    "test@exemple.com",
				Password: string(hashedPassword),
			},
			mockError:     nil,
			expectedError: codes.Unauthenticated,
		},
		{
			name:          "Invalid input - empty password",
			email:         "test@exemple.com",
			password:      "",
			mockUser:      nil,
			mockError:     nil,
			expectedError: codes.InvalidArgument,
		},
		{
			name:          "Internal error from user-base",
			email:         "test@exemple.com",
			password:      "any-password",
			mockUser:      nil,
			mockError:     status.Error(codes.Internal, "simulated database error"),
			expectedError: codes.Internal,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockClient := &mockUserBaseClient{
				userToReturn: tc.mockUser,
				errToReturn:  tc.mockError,
			}

			authSrv := &authServer{
				userBaseClient: mockClient,
			}

			req := &proto.LoginRequest{
				Email:    tc.email,
				Password: tc.password,
			}

			res, err := authSrv.Login(context.Background(), req)

			if tc.expectedError == codes.OK {
				if err != nil {
					t.Errorf("Error: %v", err)
				}
				if res == nil {
					t.Fatal("Waiting for response, got nil")
				}
				if res.Token == "" {
					t.Error("JWT tocken should not be empty")
				}
			} else {
				if err == nil {
					t.Errorf("No errors")
				}
				st, _ := status.FromError(err)
				if st.Code() != tc.expectedError {
					t.Errorf("Expected error was %s, but we got %s", tc.expectedError, st.Code())
				}
			}
		})
	}
}
