package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	gatewaypb "github.com/Costin2000/GoChat---Schwarz-Internship---2025/services/api-rest-gateway/proto"
	authpb "github.com/Costin2000/GoChat---Schwarz-Internship---2025/services/auth/proto"
	friendrequestpb "github.com/Costin2000/GoChat---Schwarz-Internship---2025/services/friend-request-base/proto"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/protobuf/encoding/protojson"
)

type server struct {
	gatewaypb.UnimplementedGatewayServiceServer
	authClient authpb.AuthServiceClient
	frClient   friendrequestpb.FriendRequestServiceClient
	upstreamTO time.Duration
}

func main() {
	httpAddr := env("GATEWAY_HTTP_ADDR", ":8080")
	authAddr := env("AUTH_ADDR", "auth:50053")
	friendRequestAddr := env("FRIEND_REQUEST_ADDR", "friend-request:50052")
	upstreamTimeout := durEnv("UPSTREAM_REQUEST_TIMEOUT", 5*time.Second)

	authConn, err := grpc.NewClient(authAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	check(err, "dial auth")
	defer authConn.Close()

	frConn, err := grpc.NewClient(friendRequestAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	check(err, "dial friend request service")
	defer frConn.Close()

	s := &server{
		authClient: authpb.NewAuthServiceClient(authConn),
		frClient:   friendrequestpb.NewFriendRequestServiceClient(frConn),
		upstreamTO: upstreamTimeout,
	}

	json := &runtime.JSONPb{
		MarshalOptions: protojson.MarshalOptions{
			UseProtoNames:   true,
			EmitUnpopulated: true,
		},
		UnmarshalOptions: protojson.UnmarshalOptions{
			DiscardUnknown: true,
		},
	}

	mux := runtime.NewServeMux(runtime.WithMarshalerOption(runtime.MIMEWildcard, json))
	if err := gatewaypb.RegisterGatewayServiceHandlerServer(context.Background(), mux, s); err != nil {
		log.Fatalf("register gateway handler: %v", err)
	}

	httpMux := http.NewServeMux()
	httpMux.Handle("/healthz", http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(http.StatusOK) }))
	httpMux.Handle("/readyz", http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(http.StatusOK) }))
	httpMux.Handle("/", withLogging(withCORS(withTimeout(mux, upstreamTimeout))))

	srv := &http.Server{
		Addr:              httpAddr,
		Handler:           httpMux,
		ReadHeaderTimeout: 5 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	go func() {
		log.Printf("api-rest-gateway HTTP on %s (auth=%s)", httpAddr, authAddr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("http serve: %v", err)
		}
	}()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	<-sigCh
	log.Println("shutting down gateway...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_ = srv.Shutdown(ctx)
}

func (s *server) Ping(ctx context.Context, _ *authpb.Empty) (*authpb.Pong, error) {
	c, cancel := context.WithTimeout(ctx, s.upstreamTO)
	defer cancel()
	return s.authClient.Ping(c, &authpb.Empty{})
}

func (s *server) CreateFriendRequest(ctx context.Context, req *friendrequestpb.CreateFriendRequestRequest) (*friendrequestpb.CreateFriendRequestResponse, error) {
	c, cancel := context.WithTimeout(ctx, s.upstreamTO)
	defer cancel()
	return s.frClient.CreateFriendRequest(c, req)
}

func env(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}

func durEnv(k string, def time.Duration) time.Duration {
	if v := os.Getenv(k); v != "" {
		if d, err := time.ParseDuration(v); err == nil {
			return d
		}
	}
	return def
}

func withTimeout(next http.Handler, d time.Duration) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), d)
		defer cancel()
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func withCORS(next http.Handler) http.Handler {
	allowOrigin := env("CORS_ALLOW_ORIGIN", "*")
	allowHeaders := env("CORS_ALLOW_HEADERS", "Content-Type, Authorization")
	allowMethods := env("CORS_ALLOW_METHODS", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", allowOrigin)
		w.Header().Set("Access-Control-Allow-Headers", allowHeaders)
		w.Header().Set("Access-Control-Allow-Methods", allowMethods)
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func withLogging(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, r)
		log.Printf("%s %s %s", r.Method, r.URL.Path, time.Since(start))
	})
}

func check(err error, msg string) {
	if err != nil {
		log.Fatalf("%s: %v", msg, err)
	}
}
