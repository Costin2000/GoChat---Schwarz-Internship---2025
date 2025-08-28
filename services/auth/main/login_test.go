package main

import (
	"context"
	"errors"
	"testing"
	"time"

	authpb "github.com/Costin2000/GoChat---Schwarz-Internship---2025/services/auth/proto"
	userbasepb "github.com/Costin2000/GoChat---Schwarz-Internship---2025/services/user-base/proto"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// gRPC client mock for user-base
type mockUserBaseClient struct {
	getUserFunc func(ctx context.Context, in *userbasepb.GetUserRequest, opts ...grpc.CallOption) (*userbasepb.User, error)
}

func (m *mockUserBaseClient) GetUser(ctx context.Context, in *userbasepb.GetUserRequest, opts ...grpc.CallOption) (*userbasepb.User, error) {
	if m.getUserFunc != nil {
		return m.getUserFunc(ctx, in, opts...)
	}
	return nil, status.Error(codes.NotFound, "not implemented")
}

func newAuthServerWithMock(m *mockUserBaseClient) *authServer {
	return &authServer{userBaseClient: m}
}

// testing helper, returns the hash bcrypt password for the original one
func hashPwd(t *testing.T, plain string) string {
	t.Helper()
	h, err := bcrypt.GenerateFromPassword([]byte(plain), bcrypt.MinCost)
	if err != nil {
		t.Fatalf("failed to hash password: %v", err)
	}
	return string(h)
}

// parsing and validating the JWT token
func parseJWT(t *testing.T, tokenStr string) jwt.MapClaims {
	t.Helper()
	token, err := jwt.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return jwtSecret, nil
	})
	if err != nil || !token.Valid {
		t.Fatalf("invalid token: %v", err)
	}
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		t.Fatalf("unexpected claims type")
	}
	return claims
}

func init() {
	jwtSecret = []byte("test-secret")
}

func TestLogin(t *testing.T) {
	const (
		emailOK  = "test@example.com"
		passOK   = "p@ss"
		userIDOK = int64(1234)
	)

	hashedOK := hashPwd(t, passOK)

	tests := []struct {
		name       string
		req        *authpb.LoginRequest
		mockClient *mockUserBaseClient
		wantCode   codes.Code
		wantUserID int64
		checkToken bool
	}{
		{
			name:       "invalid args - empty email",
			req:        &authpb.LoginRequest{Email: "", Password: "x"},
			mockClient: &mockUserBaseClient{},
			wantCode:   codes.InvalidArgument,
		},
		{
			name:       "invalid args - empty password",
			req:        &authpb.LoginRequest{Email: "a@b.com", Password: ""},
			mockClient: &mockUserBaseClient{},
			wantCode:   codes.InvalidArgument,
		},
		{
			name: "user not found",
			req:  &authpb.LoginRequest{Email: "not@found.com", Password: "anything"},
			mockClient: &mockUserBaseClient{
				getUserFunc: func(ctx context.Context, in *userbasepb.GetUserRequest, _ ...grpc.CallOption) (*userbasepb.User, error) {
					return nil, status.Error(codes.NotFound, "no such user")
				},
			},
			wantCode: codes.NotFound,
		},
		{
			name: "user-base internal error",
			req:  &authpb.LoginRequest{Email: "x@y.com", Password: "x"},
			mockClient: &mockUserBaseClient{
				getUserFunc: func(ctx context.Context, in *userbasepb.GetUserRequest, _ ...grpc.CallOption) (*userbasepb.User, error) {
					return nil, errors.New("db down")
				},
			},
			wantCode: codes.Internal,
		},
		{
			name: "invalid password",
			req:  &authpb.LoginRequest{Email: "u@ex.com", Password: "wrong-password"},
			mockClient: &mockUserBaseClient{
				getUserFunc: func(ctx context.Context, in *userbasepb.GetUserRequest, _ ...grpc.CallOption) (*userbasepb.User, error) {
					return &userbasepb.User{
						Id:       42,
						Email:    in.Email,
						Password: hashedOK,
					}, nil
				},
			},
			wantCode: codes.Unauthenticated,
		},
		{
			name: "success",
			req:  &authpb.LoginRequest{Email: emailOK, Password: passOK},
			mockClient: &mockUserBaseClient{
				getUserFunc: func(ctx context.Context, in *userbasepb.GetUserRequest, _ ...grpc.CallOption) (*userbasepb.User, error) {
					if in.Email != emailOK {
						return nil, status.Error(codes.NotFound, "unexpected email in test")
					}
					return &userbasepb.User{
						Id:       userIDOK,
						Email:    emailOK,
						Password: hashedOK,
					}, nil
				},
			},
			wantCode:   codes.OK,
			wantUserID: userIDOK,
			checkToken: true,
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			s := newAuthServerWithMock(tc.mockClient)

			ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
			defer cancel()

			resp, err := s.Login(ctx, tc.req)
			if tc.wantCode == codes.OK {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				if resp == nil {
					t.Fatal("expected non-nil response")
				}
				if resp.UserId != tc.wantUserID {
					t.Fatalf("want userID %d, got %d", tc.wantUserID, resp.UserId)
				}
				if !tc.checkToken {
					return
				}
				if resp.Token == "" {
					t.Fatal("expected non-empty token")
				}

				claims := parseJWT(t, resp.Token)

				gotUID, ok := claims["user_id"]
				if !ok {
					t.Fatal("token missing user_id claim")
				}
				switch v := gotUID.(type) {
				case float64:
					if int64(v) != tc.wantUserID {
						t.Fatalf("want user_id %d, got %v", tc.wantUserID, v)
					}
				case int64:
					if v != tc.wantUserID {
						t.Fatalf("want user_id %d, got %v", tc.wantUserID, v)
					}
				default:
					t.Fatalf("unexpected user_id type: %T", v)
				}

				expVal, ok := claims["exp"].(float64)
				if !ok {
					t.Fatal("token missing exp claim")
				}
				exp := time.Unix(int64(expVal), 0)
				if time.Until(exp) <= 0 {
					t.Fatalf("token exp is not in the future: %v", exp)
				}
				return
			}

			if err == nil {
				t.Fatalf("expected error code %v, got nil error", tc.wantCode)
			}
			st, _ := status.FromError(err)
			if st.Code() != tc.wantCode {
				t.Fatalf("want code %v, got %v (err=%v)", tc.wantCode, st.Code(), err)
			}
		})
	}
}
