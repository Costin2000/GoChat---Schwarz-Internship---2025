package main

import (
	"bufio"
	"database/sql"
	"fmt"
	"log"
	"net"
	"os"
	"path/filepath"
	"strings"

	proto "github.com/Costin2000/GoChat---Schwarz-Internship---2025/services/conversation-base/proto"
	_ "github.com/jackc/pgx/v5/stdlib"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

const port = ":50056"

type conversationService struct {
	proto.UnimplementedConversationServiceServer
	storageAccess StorageAccess
}

// env loader
func loadEnv(filename string) error {
	file, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		if len(line) == 0 || strings.HasPrefix(line, "#") {
			continue
		}
		envs := strings.SplitN(line, "=", 2)
		if len(envs) != 2 {
			continue
		}
		key := strings.TrimSpace(envs[0])
		value := strings.TrimSpace(envs[1])
		_ = os.Setenv(key, value)
	}
	return scanner.Err()
}

func main() {
	wd, err := os.Getwd()
	if err != nil {
		log.Fatalf("Could not get current working directory: %v", err)
	}

	envPath := "./db/.env"
	if filepath.Base(wd) == "conversation-base" {
		envPath = "./../../db/.env"
	}

	if err := loadEnv(envPath); err != nil {
		log.Fatalf("Error loading .env file: %v", err)
	}

	dbUser := os.Getenv("POSTGRES_USER")
	dbPassword := os.Getenv("POSTGRES_PASSWORD")
	dbName := os.Getenv("POSTGRES_DB")
	dbPort := os.Getenv("DB_PORT")

	var dbHost string
	if os.Getenv("ENV") == "docker" {
		dbHost = "postgres-db"
	} else {
		dbHost = os.Getenv("DB_HOST")
	}

	connStr := fmt.Sprintf("user=%s password=%s host=%s port=%s dbname=%s sslmode=disable", dbUser, dbPassword, dbHost, dbPort, dbName)
	db, err := sql.Open("pgx", connStr)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		log.Fatalf("Failed to ping database: %v", err)
	}
	log.Println("Successfully connected to PostgreSQL database.")

	lis, err := net.Listen("tcp", port)
	if err != nil {
		log.Fatalf("Error polling port %s %v", port, err)
	}

	storage := newPostgresAccess(db)
	ConversationServer := &conversationService{
		storageAccess: storage,
	}

	grpcServer := grpc.NewServer()
	proto.RegisterConversationServiceServer(grpcServer, ConversationServer)
	reflection.Register(grpcServer)

	fmt.Println("Conversation gRPC server listening on :50056...")

	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
