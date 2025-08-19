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

	proto "github.com/Costin2000/GoChat---Schwarz-Internship---2025/services/user-base/proto"
	_ "github.com/jackc/pgx/v5/stdlib"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

const port = ":50051"

type userBaseServer struct {
	proto.UnimplementedUserBaseServer
	proto.UnimplementedUserServiceServer
	db *sql.DB
}

// retrieve db setup from the .env file
func loadEnv(filename string) error {

	file, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()

		// ignore empty lines and comments
		if len(line) == 0 || strings.HasPrefix(line, "#") {
			continue
		}

		envs := strings.SplitN(line, "=", 2)
		if len(envs) != 2 {
			continue // skip bad lines
		}

		key := strings.TrimSpace(envs[0])
		value := strings.TrimSpace(envs[1])

		// Set the environment variable
		if err := os.Setenv(key, value); err != nil {
			log.Printf("Warning: could not set env var %s: %v", key, err)
		}
	}

	return scanner.Err()
}

func main() {

	wd, err := os.Getwd()
	if err != nil {
		log.Fatalf("Could not get current working directory: %v", err)
	}

	envPath := "./db/.env" // db env path from root
	if filepath.Base(wd) == "main" {
		envPath = "./../../../db/.env" // db env path from user-base service directory
	}

	// db connection
	if err := loadEnv(envPath); err != nil {
		log.Fatalf("Error loading .env file: %v", err)
	}

	dbUser := os.Getenv("POSTGRESS_USER")
	dbPassword := os.Getenv("POSTGRESS_PASSWORD")
	dbName := os.Getenv("POSTGRESS_DB")
	dbPort := os.Getenv("DB_PORT")
	dbHost := os.Getenv("DB_HOST")

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

	// network connection
	lis, err := net.Listen("tcp", port)
	if err != nil {
		log.Fatalf("Error polling port %s %v", port, err)
	}

	// server connections
	UserBaseServer := &userBaseServer{
		db: db,
	}

	grpcServer := grpc.NewServer()

	proto.RegisterUserBaseServer(grpcServer, UserBaseServer)
	proto.RegisterUserServiceServer(grpcServer, UserBaseServer)
	reflection.Register(grpcServer)

	log.Printf("gRPC polling on port %s...", port)

	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("Error starting server: %v", err)
	}
}
