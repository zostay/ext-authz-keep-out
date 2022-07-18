package main

import (
	"fmt"
	"net"

	authv3 "github.com/envoyproxy/go-control-plane/envoy/service/auth/v3"
	"google.golang.org/grpc"
)

func main() {
	ko := keepOut{
		realm: "Mordac the Preventer",
		user:  "test123",
		pass:  "test321",
	}

	grpcSrv := grpc.NewServer()

	authv3.RegisterAuthorizationServer(grpcSrv, &ko)

	listener, err := net.Listen("tcp", ":8080")
	if err != nil {
		panic(fmt.Sprintf("failed to start listener on :8080: %v", err))
	}

	fmt.Println("Starting GRPC server on :8080")

	grpcSrv.Serve(listener)
}
