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
	"time"

	pb "github.com/Costin2000/GoChat---Schwarz-Internship---2025/services/user-base/proto"
	_ "github.com/jackc/pgx/v5/stdlib"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	amqp "github.com/rabbitmq/amqp091-go"
)

const port = ":50051"

type UserService struct {
	storageAccess StorageAccess
	pb.UnimplementedUserServiceServer
	emailPub EmailPublisher
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
	if filepath.Base(wd) == "user-base" {
		envPath = "./../../db/.env" // db env path from user-base service directory
	}

	// db connection
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

	rmqAddr := os.Getenv("RABBITMQ_ADDR")
	var emailPub EmailPublisher
	if rmqAddr != "" {
		conn, err := connectToRabbitMQWithRetries(rmqAddr)
		if err != nil {
			log.Printf("WARN: cannot connect to RabbitMQ (%s): %v (CreatUser will work, but no emails published)", rmqAddr, err)
		} else {
			defer conn.Close()
			pub, err := newAmqpEmailPublisher(conn)
			if err != nil {
				log.Printf("WARN: cannot create email publisher: %v", err)
			} else {
				defer pub.Close()
				emailPub = pub
				log.Println("Connected to RabbitMQ and email publisher ready!")
			}
		}
	} else {
		log.Println("WARN: RABBITMQ_ADDR not set; emails will not be published")
	}

	// network connection
	lis, err := net.Listen("tcp", port)
	if err != nil {
		log.Fatalf("Error polling port %s %v", port, err)
	}

	// server connections
	storage := newPostgresAccess(db)
	UserBaseServer := &UserService{
		storageAccess: storage,
		emailPub:      emailPub,
	}

	grpcServer := grpc.NewServer()
	pb.RegisterUserServiceServer(grpcServer, UserBaseServer)
	reflection.Register(grpcServer)

	log.Printf("gRPC polling on port %s...", port)

	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("Error starting server: %v", err)
	}
}

func connectToRabbitMQWithRetries(addr string) (*amqp.Connection, error) {
	var conn *amqp.Connection
	var err error
	maxRetries := 10
	backoff := 3 * time.Second

	for i := 1; i <= maxRetries; i++ {
		conn, err = amqp.Dial(addr)
		if err == nil {
			log.Println("user-base: connected to RabbitMQ")
			return conn, nil
		}
		log.Printf("user-base: cannot connect to RabbitMQ, retrying in %v... (%d/%d)", backoff, i, maxRetries)
		time.Sleep(backoff)
	}
	return nil, err
}
