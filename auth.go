package main

import (
	"bytes"
	"context"
	"encoding/base64"
	"fmt"
	"log"
	"net"
	"strings"

	corev3 "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	authv3 "github.com/envoyproxy/go-control-plane/envoy/service/auth/v3"
	typev3 "github.com/envoyproxy/go-control-plane/envoy/type/v3"
	"google.golang.org/genproto/googleapis/rpc/status"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
)

type keepOut struct {
	realm string
	user  string
	pass  string
}

func NewKeepOut(realm, user, pass string) *keepOut {
	return &keepOut{realm, user, pass}
}

func (ko *keepOut) Run(port string, portChan chan int) {
	grpcSrv := grpc.NewServer()

	authv3.RegisterAuthorizationServer(grpcSrv, ko)

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

func (k *keepOut) unauthorizedResponse() *authv3.CheckResponse {
	// log.Print("deny")
	return &authv3.CheckResponse{
		Status: &status.Status{
			Code: int32(codes.PermissionDenied),
		},
		HttpResponse: &authv3.CheckResponse_DeniedResponse{
			DeniedResponse: &authv3.DeniedHttpResponse{
				Status: &typev3.HttpStatus{
					Code: typev3.StatusCode_Unauthorized,
				},
				Headers: []*corev3.HeaderValueOption{
					{
						Header: &corev3.HeaderValue{
							Key:   "WWW-Authenticate",
							Value: fmt.Sprintf(`Basic realm=%q`, k.realm),
						},
					},
				},
			},
		},
		DynamicMetadata: nil,
	}
}

func (k *keepOut) Check(
	_ context.Context,
	request *authv3.CheckRequest,
) (*authv3.CheckResponse, error) {
	log.Print("new request")

	if request == nil {
		log.Println("request is nil")
		return k.unauthorizedResponse(), nil
	}

	attr := request.GetAttributes()
	if attr == nil {
		log.Println("attr is nil")
		return k.unauthorizedResponse(), nil
	}

	req := attr.GetRequest()
	if req == nil {
		log.Println("req is nil")
		return k.unauthorizedResponse(), nil
	}

	http := req.GetHttp()
	if http == nil {
		log.Println("http is nil")
		return k.unauthorizedResponse(), nil
	}

	headers := http.GetHeaders()
	if headers == nil {
		log.Println("headers are nil")
		return k.unauthorizedResponse(), nil
	}

	// log.Printf("headers %+v", headers["authorization"])

	authHeader, ok := headers["authorization"]
	if !ok {
		log.Println("no authorization headers")
		return k.unauthorizedResponse(), nil
	}

	authValues := strings.Split(authHeader, ",")
	basicAuthValues := make([]string, 0, len(authValues))
	for _, authValue := range authValues {
		authValue = strings.TrimSpace(authValue)
		if strings.HasPrefix(strings.ToLower(authValue), "basic ") {
			basicAuthValues = append(basicAuthValues, strings.TrimSpace(authValue[5:]))
		}

		if len(basicAuthValues) == 0 {
			log.Println("no basic auth present")
			return k.unauthorizedResponse(), nil
		}
	}

	for _, encodedAuth := range basicAuthValues {
		auth, err := base64.StdEncoding.DecodeString(encodedAuth)
		if err != nil {
			log.Println("failed to decode basic auth base64")
			return k.unauthorizedResponse(), nil
		}

		if !bytes.ContainsRune(auth, ':') {
			log.Println("no colon in basic auth")
			return k.unauthorizedResponse(), nil
		}

		parts := strings.SplitN(string(auth), ":", 2)
		user, pass := parts[0], parts[1]

		if user != k.user {
			log.Printf("user %q does not match %q", user, k.user)
			return k.unauthorizedResponse(), nil
		}

		if pass != k.pass {
			log.Printf("pass %q does not match %q", pass, k.pass)
			return k.unauthorizedResponse(), nil
		}
	}

	// log.Print("pass")

	return &authv3.CheckResponse{
		Status: &status.Status{
			Code: int32(codes.OK),
		},
	}, nil
}
