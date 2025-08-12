module github.com/Costin2000/GoChat---Schwarz-Internship---2025/services/api-rest-gateway

go 1.24.4

require (
	github.com/Costin2000/GoChat---Schwarz-Internship---2025/auth v0.0.0
	github.com/grpc-ecosystem/grpc-gateway/v2 v2.27.1
	google.golang.org/genproto/googleapis/api v0.0.0-20250603155806-513f23925822
	google.golang.org/grpc v1.74.2
	google.golang.org/protobuf v1.36.6
)

require (
	golang.org/x/net v0.40.0 // indirect
	golang.org/x/sys v0.33.0 // indirect
	golang.org/x/text v0.26.0 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20250603155806-513f23925822 // indirect
)

replace github.com/Costin2000/GoChat---Schwarz-Internship---2025/auth => ../auth

replace github.com/Costin2000/GoChat---Schwarz-Internship---2025/api-rest-gateway => ./
