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

// userBaseClient mock
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

func hashPwd(t *testing.T, plain string) string {
	t.Helper()
	h, err := bcrypt.GenerateFromPassword([]byte(plain), bcrypt.MinCost)
	if err != nil {
		t.Fatalf("failed to hash password: %v", err)
	}
	return string(h)
}

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

// JWT secret for tests
func init() {
	jwtSecret = []byte("test-secret")
}

func TestLogin_InvalidArguments(t *testing.T) {
	s := newAuthServerWithMock(&mockUserBaseClient{})
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	// empty email
	_, err := s.Login(ctx, &authpb.LoginRequest{Email: "", Password: "x"})
	if status.Code(err) != codes.InvalidArgument {
		t.Fatalf("want InvalidArgument, got %v (err=%v)", status.Code(err), err)
	}

	// empty password
	_, err = s.Login(ctx, &authpb.LoginRequest{Email: "a@b.com", Password: ""})
	if status.Code(err) != codes.InvalidArgument {
		t.Fatalf("want InvalidArgument, got %v (err=%v)", status.Code(err), err)
	}
}

func TestLogin_UserNotFound(t *testing.T) {
	mock := &mockUserBaseClient{
		getUserFunc: func(ctx context.Context, in *userbasepb.GetUserRequest, _ ...grpc.CallOption) (*userbasepb.User, error) {
			return nil, status.Error(codes.NotFound, "no such user")
		},
	}
	s := newAuthServerWithMock(mock)

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	_, err := s.Login(ctx, &authpb.LoginRequest{Email: "not@found.com", Password: "anything"})
	if status.Code(err) != codes.NotFound {
		t.Fatalf("want NotFound, got %v (err=%v)", status.Code(err), err)
	}
}

func TestLogin_UserBaseInternalError(t *testing.T) {
	mock := &mockUserBaseClient{
		getUserFunc: func(ctx context.Context, in *userbasepb.GetUserRequest, _ ...grpc.CallOption) (*userbasepb.User, error) {
			return nil, errors.New("db down")
		},
	}
	s := newAuthServerWithMock(mock)

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	_, err := s.Login(ctx, &authpb.LoginRequest{Email: "x@y.com", Password: "x"})
	if status.Code(err) != codes.Internal {
		t.Fatalf("want Internal, got %v (err=%v)", status.Code(err), err)
	}
}

func TestLogin_InvalidPassword(t *testing.T) {
	hashed := hashPwd(t, "correct-password")
	mock := &mockUserBaseClient{
		getUserFunc: func(ctx context.Context, in *userbasepb.GetUserRequest, _ ...grpc.CallOption) (*userbasepb.User, error) {
			return &userbasepb.User{
				Id:       42,
				Email:    in.Email,
				Password: hashed,
			}, nil
		},
	}
	s := newAuthServerWithMock(mock)

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	// checking with a wrong password
	_, err := s.Login(ctx, &authpb.LoginRequest{Email: "u@ex.com", Password: "wrong-password"})
	if status.Code(err) != codes.Unauthenticated {
		t.Fatalf("want Unauthenticated, got %v (err=%v)", status.Code(err), err)
	}
}

func TestLogin_Success(t *testing.T) {
	userID := int64(1234)
	const email = "test@example.com"
	const plainPassword = "p@ss"

	hashed := hashPwd(t, plainPassword)
	mock := &mockUserBaseClient{
		getUserFunc: func(ctx context.Context, in *userbasepb.GetUserRequest, _ ...grpc.CallOption) (*userbasepb.User, error) {
			if in.GetEmail() != email {
				return nil, status.Error(codes.NotFound, "unexpected email in test")
			}
			return &userbasepb.User{
				Id:       userID,
				Email:    email,
				Password: hashed,
			}, nil
		},
	}
	s := newAuthServerWithMock(mock)

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	resp, err := s.Login(ctx, &authpb.LoginRequest{Email: email, Password: plainPassword})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.UserId != userID {
		t.Fatalf("want userID %d, got %d", userID, resp.UserId)
	}
	if resp.Token == "" {
		t.Fatal("expected non-empty token")
	}

	claims := parseJWT(t, resp.Token)

	// user_id
	gotUID, ok := claims["user_id"]
	if !ok {
		t.Fatal("token missing user_id claim")
	}
	switch v := gotUID.(type) {
	case float64:
		if int64(v) != userID {
			t.Fatalf("want user_id %d, got %v", userID, v)
		}
	case int64:
		if v != userID {
			t.Fatalf("want user_id %d, got %v", userID, v)
		}
	default:
		t.Fatalf("unexpected user_id type: %T", v)
	}

	// exp in future
	expVal, ok := claims["exp"].(float64)
	if !ok {
		t.Fatal("token missing exp claim")
	}
	exp := time.Unix(int64(expVal), 0)
	if time.Until(exp) <= 0 {
		t.Fatalf("token exp is not in the future: %v", exp)
	}
}
