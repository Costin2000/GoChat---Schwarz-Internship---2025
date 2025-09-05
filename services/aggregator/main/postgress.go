package main

import (
	"database/sql"

	"context"

	userpb "github.com/Costin2000/GoChat---Schwarz-Internship---2025/services/user-base/proto"
)

type StorageAccess interface {
	fetchFriends(ctx context.Context, userId string) ([]*userpb.User, error)
}

type PostgresAccess struct {
	db *sql.DB
}

func newPostgresAccess(db *sql.DB) *PostgresAccess {
	return &PostgresAccess{db: db}
}

func (pa *PostgresAccess) fetchFriends(ctx context.Context, userId string) ([]*userpb.User, error) {
	// unimplemented endpoint
	return nil, nil
}
