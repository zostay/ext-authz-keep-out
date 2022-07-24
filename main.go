package main

import (
	"fmt"
	"net"

	authv3 "github.com/envoyproxy/go-control-plane/envoy/service/auth/v3"
	"google.golang.org/grpc"
)

func StartServer(port string, portChan chan int) {
	ko := keepOut{
		realm: "Mordac the Preventer",
		user:  "test123",
		pass:  "test321",
	}

	grpcSrv := grpc.NewServer()

	authv3.RegisterAuthorizationServer(grpcSrv, &ko)

	listener, err := net.Listen("tcp", port)
	if err != nil {
		panic(fmt.Sprintf("failed to start listener on :8080: %v", err))
	}

	if portChan != nil {
		portChan <- listener.Addr().(*net.TCPAddr).Port
	}

	actualPort := listener.Addr().String()
	fmt.Printf("Starting GRPC server on %s\n", actualPort)

	grpcSrv.Serve(listener)
}

func main() {
	StartServer("0.0.0.0:8080", nil)
}
