package main

import (
	"context"
	pb "github.com/Costin2000/GoChat---Schwarz-Internship---2025/services/user-base/proto"
)

type mockStorage struct {
	createUserFunc     func(ctx context.Context, user *pb.User) (*pb.User, error)
	getUserByEmailFunc func(ctx context.Context, email string) (*pb.User, error)
}

func (m *mockStorage) createUser(ctx context.Context, user *pb.User) (*pb.User, error) {
	return m.createUserFunc(ctx, user)
}

func (m *mockStorage) getUserByEmail(ctx context.Context, email string) (*pb.User, error) {
	return m.getUserByEmailFunc(ctx, email)
}

type StorageMockOptions struct {
	createUserFunc     func(ctx context.Context, user *pb.User) (*pb.User, error)
	getUserByEmailFunc func(ctx context.Context, email string) (*pb.User, error)
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
}

func NewMockService(opts ServiceMockOptions) *UserService {
	storage := newMockStorageAccess(StorageMockOptions{})
	if opts.storageAccess != nil {
		storage = opts.storageAccess
	}

	return &UserService{
		storageAccess: storage,
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
