package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"
	"os/exec"
	"testing"

	proto "github.com/Costin2000/GoChat---Schwarz-Internship---2025/services/user-base/proto"
	"google.golang.org/grpc/status"
)

func TestCreateUser_Integration(t *testing.T) {
	testUser := &proto.User{
		FirstName: "John",
		LastName:  "Doe",
		UserName:  "johndoe",
		Email:     "johndoe@example.com",
		Password:  "password123",
	}

	startDbCmd := exec.Command("bash", "./../scripts/db-start.sh")
	stopDbCmd := exec.Command("bash", "./../scripts/db-stop.sh")

	if err := startDbCmd.Run(); err != nil {
		log.Fatalf("Error starting DB: %v", err)
	}

	// load env
	envPath := "./../../../db/.env"
	if err := loadEnv(envPath); err != nil {
		log.Fatalf("Error loading env: %v", err)
	}

	dbUser := os.Getenv("POSTGRES_USER")
	dbPassword := os.Getenv("POSTGRES_PASSWORD")
	dbName := os.Getenv("POSTGRES_DB")
	dbPort := os.Getenv("DB_PORT")

	connStr := fmt.Sprintf("user=%s password=%s host=localhost port=%s dbname=%s sslmode=disable",
		dbUser, dbPassword, dbPort, dbName)

	db, err := sql.Open("pgx", connStr)
	if err != nil {
		t.Fatalf("Failed to connect to DB: %v", err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		t.Fatalf("Failed to ping DB: %v", err)
	}

	s := &userBaseServer{db: db}

	// Clean table
	db.ExecContext(context.Background(), `DELETE FROM "User"`)

	resp, err := s.CreateUser(context.Background(), &proto.CreateUserRequest{User: testUser})
	if err != nil {
		st, _ := status.FromError(err)
		t.Fatalf("Expected success, got error: %v", st.Message())
	}

	if resp.User.Email != testUser.Email {
		t.Errorf("Expected email %s, got %s", testUser.Email, resp.User.Email)
	}

	if resp.User.Password == testUser.Password {
		t.Errorf("Password should be hashed, but got plain password")
	}

	// Stop DB
	if err := stopDbCmd.Run(); err != nil {
		log.Fatalf("Error stopping DB: %v", err)
	}
}
