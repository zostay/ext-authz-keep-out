package main

import (
	"flag"
	"fmt"
	"log"
	"net"

	authv3 "github.com/envoyproxy/go-control-plane/envoy/service/auth/v3"
	"google.golang.org/grpc"
)

var (
	realm, user, pass string
)

func init() {
	flag.StringVar(&realm, "realm", "Invitation Only", "set the name of the Basic realm to present")
	flag.StringVar(&user, "user", "demo", "set the name of the user to accept")
	flag.StringVar(&pass, "pass", "demo", "set the password of the user to accept")
}

func StartServer(port string, portChan chan int) {
	ko := keepOut{realm, user, pass}

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
	log.Printf("Starting GRPC server on %s", actualPort)

	grpcSrv.Serve(listener)
}

func main() {
	flag.Parse()
	StartServer("0.0.0.0:8080", nil)
}
