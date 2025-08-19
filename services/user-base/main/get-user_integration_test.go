package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"testing"

	proto "github.com/Costin2000/GoChat---Schwarz-Internship---2025/services/user-base/proto"
)

func TestGetUser_Integration(t *testing.T) {

	// db connection
	wd, err := os.Getwd()
	if err != nil {
		log.Fatalf("Could not get current working directory: %v", err)
	}

	envPath := "./db/.env" // db env path from root
	if filepath.Base(wd) == "main" {
		envPath = "./../../../db/.env" // db env path from user-base service directory
	}

	if err := loadEnv(envPath); err != nil {
		log.Fatalf("Error loading .env file: %v", err)
	}

	dbUser := os.Getenv("POSTGRESS_USER")
	dbPassword := os.Getenv("POSTGRESS_PASSWORD")
	dbName := os.Getenv("POSTGRESS_DB")
	dbPort := os.Getenv("DB_PORT")

	connStr := fmt.Sprintf("user=%s password=%s host=localhost port=%s dbname=%s sslmode=disable", dbUser, dbPassword, dbPort, dbName)

	db, err := sql.Open("pgx", connStr)
	if err != nil {
		t.Fatalf("Failed to connect to database for integration test: %v", err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		log.Fatalf("Failed to ping database: %v", err)
	}

	// server connections
	s := &userBaseServer{db: db}

	// Define the user we expect to find
	dummyEmail := "test@example.com"
	expectedUserName := "testuser"

	req := &proto.GetUserRequest{Email: dummyEmail}

	// send dummy request to the GetUser method
	res, err := s.GetUser(context.Background(), req)
	if err != nil {
		t.Fatalf("GetUser failed with an unexpected error: %v", err)
	}

	// sssert results
	if res == nil || res.Id == 0 {
		t.Fatal("Response or user was nil, but a user was expected.")
	}

	if res.Email != dummyEmail {
		t.Errorf("Expected email %s, but got %s", dummyEmail, res.Email)
	}

	if res.UserName != expectedUserName {
		t.Errorf("Expected user_name %s, but got %s", expectedUserName, res.UserName)
	}

	t.Logf("Successfully retrieved user: %s (%s)", res.UserName, res.Email)
}
