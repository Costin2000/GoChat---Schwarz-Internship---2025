// services/auth/main/login.go
package main

import (
	"context"
	"time"

	userbasepb "github.com/Costin2000/GoChat---Schwarz-Internship---2025/services/user-base/proto"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"

	proto "github.com/Costin2000/GoChat---Schwarz-Internship---2025/services/auth/proto"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s *authServer) Login(ctx context.Context, req *proto.LoginRequest) (*proto.LoginResponse, error) {
	if req.Email == "" || req.Password == "" {
		return nil, status.Error(codes.InvalidArgument, "email and password are required")
	}

	userReq := &userbasepb.GetUserRequest{Email: req.Email}
	user, err := s.userBaseClient.GetUser(ctx, userReq)
	if err != nil {
		if status.Code(err) == codes.NotFound {
			return nil, status.Error(codes.NotFound, "user not found")
		}
		return nil, status.Errorf(codes.Internal, "failed to get user: %v", err)
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password))
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, "invalid credentials")
	}

	tokenString, err := generateJWT(user.Id)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to generate token: %v", err)
	}

	return &proto.LoginResponse{
		UserId: user.Id,
		Token:  tokenString,
	}, nil
}

func generateJWT(userID int64) (string, error) {
	claims := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": userID,
		"exp":     time.Now().Add(time.Hour * 24).Unix(),
	})

	tokenString, err := claims.SignedString(jwtSecret)
	if err != nil {
		return "", err
	}

	return tokenString, nil
}
