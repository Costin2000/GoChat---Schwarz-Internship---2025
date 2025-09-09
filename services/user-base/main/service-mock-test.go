package main

import (
	"context"

	pbauth "github.com/Costin2000/GoChat---Schwarz-Internship---2025/services/auth/proto"
	pb "github.com/Costin2000/GoChat---Schwarz-Internship---2025/services/user-base/proto"
	"google.golang.org/grpc"
)

type mockStorage struct {
	createUserFunc     func(ctx context.Context, user *pb.User) (*pb.User, error)
	getUserByEmailFunc func(ctx context.Context, email string) (*pb.User, error)
	listUsersFunc      func(ctx context.Context, req *pb.ListUsersRequest) (*pb.ListUsersResponse, error)
}

func (m *mockStorage) createUser(ctx context.Context, user *pb.User) (*pb.User, error) {
	return m.createUserFunc(ctx, user)
}

func (m *mockStorage) getUserByEmail(ctx context.Context, email string) (*pb.User, error) {
	return m.getUserByEmailFunc(ctx, email)
}
func (m *mockStorage) listUsers(ctx context.Context, req *pb.ListUsersRequest) (*pb.ListUsersResponse, error) {
	if m.listUsersFunc != nil {
		return m.listUsersFunc(ctx, req)
	}
	return nil, nil
}

type authMock struct {
	loginFunc func(ctx context.Context, req *pbauth.LoginRequest, opts ...grpc.CallOption) (*pbauth.LoginResponse, error)
}

func (m *authMock) Login(ctx context.Context, req *pbauth.LoginRequest, opts ...grpc.CallOption) (*pbauth.LoginResponse, error) {
	return m.loginFunc(ctx, req)
}

type StorageMockOptions struct {
	createUserFunc     func(ctx context.Context, user *pb.User) (*pb.User, error)
	getUserByEmailFunc func(ctx context.Context, email string) (*pb.User, error)
	listUsersFunc      func(ctx context.Context, req *pb.ListUsersRequest) (*pb.ListUsersResponse, error)
}

func newMockStorageAccess(
	opts StorageMockOptions,
) StorageAccess {
	createUserFunc := func(ctx context.Context, user *pb.User) (*pb.User, error) {
		return fixtureUser(), nil
	}
	if opts.createUserFunc != nil {
		createUserFunc = opts.createUserFunc
	}

	getUserByEmailFunc := func(ctx context.Context, email string) (*pb.User, error) {
		return fixtureUser(), nil
	}
	if opts.getUserByEmailFunc != nil {
		getUserByEmailFunc = opts.getUserByEmailFunc
	}

	return &mockStorage{
		createUserFunc:     createUserFunc,
		getUserByEmailFunc: getUserByEmailFunc,
	}
}

type ServiceMockOptions struct {
	storageAccess StorageAccess
	authMock      authClient
}

func NewMockService(opts ServiceMockOptions) *UserService {
	storage := newMockStorageAccess(StorageMockOptions{})
	if opts.storageAccess != nil {
		storage = opts.storageAccess
	}

	var authCl authClient
	if opts.authMock != nil {
		authCl = opts.authMock
	} else {
		authCl = &authMock{
			loginFunc: func(ctx context.Context, req *pbauth.LoginRequest, opts ...grpc.CallOption) (*pbauth.LoginResponse, error) {
				return &pbauth.LoginResponse{Token: "myCustomToken", UserId: 1}, nil
			},
		}
	}

	return &UserService{
		storageAccess: storage,
		authClient:    authCl,
	}
}

func fixtureUser(mods ...func(user *pb.User)) *pb.User {
	user := &pb.User{
		Id:        1,
		FirstName: "John",
		LastName:  "Walter",
		Email:     "johnwalter@yahoo.com",
		Password:  "secretpassword",
	}
	for _, mod := range mods {
		mod(user)
	}
	return user
}
