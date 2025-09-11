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

	pb "github.com/Costin2000/GoChat---Schwarz-Internship---2025/services/message-base/proto"
	_ "github.com/jackc/pgx/v5/stdlib"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

const port = ":50055"

type MessageService struct {
	storageAccess StorageAccess
	pb.UnimplementedMessageServiceServer
}

func loadEnv(filename string) error {
	f, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer f.Close()

	sc := bufio.NewScanner(f)
	for sc.Scan() {
		line := sc.Text()
		if len(line) == 0 || strings.HasPrefix(line, "#") {
			continue
		}
		kv := strings.SplitN(line, "=", 2)
		if len(kv) != 2 {
			continue
		}
		if err := os.Setenv(strings.TrimSpace(kv[0]), strings.TrimSpace(kv[1])); err != nil {
			log.Printf("warn: set env %s: %v", kv[0], err)
		}
	}
	return sc.Err()
}

func main() {
	wd, err := os.Getwd()
	if err != nil {
		log.Fatalf("cwd: %v", err)
	}

	envPath := "./db/.env"
	if filepath.Base(wd) == "message-base" {
		envPath = "./../../db/.env"
	}
	if err := loadEnv(envPath); err != nil {
		log.Fatalf("load env: %v", err)
	}

	dbUser := os.Getenv("POSTGRES_USER")
	dbPass := os.Getenv("POSTGRES_PASSWORD")
	dbName := os.Getenv("POSTGRES_DB")
	dbPort := os.Getenv("DB_PORT")

	var dbHost string
	if os.Getenv("ENV") == "docker" {
		dbHost = "postgres-db"
	} else {
		dbHost = os.Getenv("DB_HOST")
	}

	dsn := fmt.Sprintf("user=%s password=%s host=%s port=%s dbname=%s sslmode=disable", dbUser, dbPass, dbHost, dbPort, dbName)
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		log.Fatalf("open db: %v", err)
	}
	defer db.Close()
	if err := db.Ping(); err != nil {
		log.Fatalf("ping db: %v", err)
	}
	log.Println("MessageBase: connected to PostgreSQL")

	lis, err := net.Listen("tcp", port)
	if err != nil {
		log.Fatalf("listen %s: %v", port, err)
	}

	s := grpc.NewServer()
	pb.RegisterMessageServiceServer(s, &MessageService{storageAccess: newPostgresAccess(db)})
	reflection.Register(s)

	log.Printf("MessageBase gRPC listening on %s", port)
	if err := s.Serve(lis); err != nil {
		log.Fatalf("serve: %v", err)
	}
}
