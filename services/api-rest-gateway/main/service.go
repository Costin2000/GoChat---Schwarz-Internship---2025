package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	aggrpb "github.com/Costin2000/GoChat---Schwarz-Internship---2025/services/aggregator/proto"
	gatewaypb "github.com/Costin2000/GoChat---Schwarz-Internship---2025/services/api-rest-gateway/proto"
	authpb "github.com/Costin2000/GoChat---Schwarz-Internship---2025/services/auth/proto"
	conversationpb "github.com/Costin2000/GoChat---Schwarz-Internship---2025/services/conversation-base/proto"
	friendrequestpb "github.com/Costin2000/GoChat---Schwarz-Internship---2025/services/friend-request-base/proto"
	messagepb "github.com/Costin2000/GoChat---Schwarz-Internship---2025/services/message-base/proto"
	userbasepb "github.com/Costin2000/GoChat---Schwarz-Internship---2025/services/user-base/proto"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/protobuf/encoding/protojson"
)

type server struct {
	gatewaypb.UnimplementedGatewayServiceServer
	authClient         authpb.AuthServiceClient
	frClient           friendrequestpb.FriendRequestServiceClient
	userBaseClient     userbasepb.UserServiceClient
	aggrClient         aggrpb.AggregatorServiceClient
	messageClient      messagepb.MessageServiceClient
	conversationClient conversationpb.ConversationServiceClient
	upstreamTO         time.Duration
}

func main() {
	httpAddr := env("GATEWAY_HTTP_ADDR", ":8080")
	authAddr := env("AUTH_ADDR", "auth:50053")
	friendRequestAddr := env("FRIEND_REQUEST_ADDR", "friend-request:50052")
	userBaseAddr := env("USER_BASE_ADDR", "user-base:50051")
	aggrReqAddr := env("AGGR_REQUEST_ADDR", "aggregator:50054")
	upstreamTimeout := durEnv("UPSTREAM_REQUEST_TIMEOUT", 5*time.Second)
	messageAddr := env("MESSAGE_BASE_ADDR", "message-base:50055")
	conversationAddr := env("CONVERSATION_ADDR", "conversation:50056")

	authConn, err := grpc.NewClient(authAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	check(err, "dial auth")
	defer authConn.Close()

	userBaseConn, err := grpc.NewClient(userBaseAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	check(err, "dial user-base")
	defer userBaseConn.Close()

	frConn, err := grpc.NewClient(friendRequestAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	check(err, "dial friend request service")
	defer frConn.Close()

	aggrConn, err := grpc.NewClient(aggrReqAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	check(err, "dial aggregator service")
	defer aggrConn.Close()

	msgConn, err := grpc.NewClient(messageAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	check(err, "dial message-base")
	defer msgConn.Close()

	convConn, err := grpc.NewClient(conversationAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	check(err, "dial conversation service")
	defer convConn.Close()

	s := &server{
		authClient:         authpb.NewAuthServiceClient(authConn),
		frClient:           friendrequestpb.NewFriendRequestServiceClient(frConn),
		userBaseClient:     userbasepb.NewUserServiceClient(userBaseConn),
		aggrClient:         aggrpb.NewAggregatorServiceClient(aggrConn),
		upstreamTO:         upstreamTimeout,
		messageClient:      messagepb.NewMessageServiceClient(msgConn),
		conversationClient: conversationpb.NewConversationServiceClient(convConn),
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
	httpMux.Handle("/", withLogging(withCORS(withAuth(withTimeout(mux, upstreamTimeout)))))

	srv := &http.Server{
		Addr:              httpAddr,
		Handler:           httpMux,
		ReadHeaderTimeout: 5 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	go func() {
		log.Printf("api-rest-gateway HTTP on %s (auth=%s, user-base=%s)", httpAddr, authAddr, userBaseAddr)
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

func (s *server) UpdateFriendRequest(ctx context.Context, req *friendrequestpb.UpdateFriendRequestRequest) (*friendrequestpb.UpdateFriendRequestResponse, error) {
	c, cancel := context.WithTimeout(ctx, s.upstreamTO)
	defer cancel()
	return s.frClient.UpdateFriendRequest(c, req)
}

func (s *server) Login(ctx context.Context, req *authpb.LoginRequest) (*authpb.LoginResponse, error) {
	c, cancel := context.WithTimeout(ctx, s.upstreamTO)
	defer cancel()
	return s.authClient.Login(c, req)
}

func (s *server) CreateUser(ctx context.Context, req *userbasepb.CreateUserRequest) (*userbasepb.CreateUserResponse, error) {
	c, cancel := context.WithTimeout(ctx, s.upstreamTO)
	defer cancel()
	return s.userBaseClient.CreateUser(c, req)
}

func (s *server) ListFriendRequests(ctx context.Context, req *friendrequestpb.ListFriendRequestsRequest) (*friendrequestpb.ListFriendRequestsResponse, error) {
	c, cancel := context.WithTimeout(ctx, s.upstreamTO)
	defer cancel()
	return s.frClient.ListFriendRequests(c, req)
}

func (s *server) ListUsers(ctx context.Context, req *userbasepb.ListUsersRequest) (*userbasepb.ListUsersResponse, error) {
	c, cancel := context.WithTimeout(ctx, s.upstreamTO)
	defer cancel()
	return s.userBaseClient.ListUsers(c, req)
}

func (s *server) FetchUserFriends(ctx context.Context, req *aggrpb.FetchUserFriendsRequest) (*aggrpb.FetchUserFriendsResponse, error) {
	c, cancel := context.WithTimeout(ctx, s.upstreamTO)
	defer cancel()
	return s.aggrClient.FetchUserFriends(c, req)
}

func (s *server) CreateMessage(ctx context.Context, req *messagepb.CreateMessageRequest) (*messagepb.CreateMessageResponse, error) {
	c, cancel := context.WithTimeout(ctx, s.upstreamTO)
	defer cancel()
	return s.messageClient.CreateMessage(c, req)
}

func (s *server) ListMessages(ctx context.Context, req *messagepb.ListMessagesRequest) (*messagepb.ListMessagesResponse, error) {
	c, cancel := context.WithTimeout(ctx, s.upstreamTO)
	defer cancel()
	return s.messageClient.ListMessages(c, req)
}

func (s *server) CreateConversation(ctx context.Context, req *conversationpb.CreateConversationRequest) (*conversationpb.CreateConversationResponse, error) {
	c, cancel := context.WithTimeout(ctx, s.upstreamTO)
	defer cancel()
	return s.conversationClient.CreateConversation(c, req)
}

func (s *server) ListConversations(ctx context.Context, req *conversationpb.ListConversationsRequest) (*conversationpb.ListConversationsResponse, error) {
	c, cancel := context.WithTimeout(ctx, s.upstreamTO)
	defer cancel()
	return s.conversationClient.ListConversations(c, req)
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
