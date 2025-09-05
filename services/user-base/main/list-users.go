package main

import (
	"context"

	pb "github.com/Costin2000/GoChat---Schwarz-Internship---2025/services/user-base/proto"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (svc *UserService) ListUsers(ctx context.Context, req *pb.ListUsersRequest) (*pb.ListUsersResponse, error) {
	if req.GetPageSize() < 0 {
		return nil, status.Error(codes.InvalidArgument, "pageSize cannot be negative")
	}

	return svc.storageAccess.listUsers(ctx, req)
}
