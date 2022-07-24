package main

import (
	"bytes"
	"context"
	"encoding/base64"
	"fmt"
	"strings"

	corev3 "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	authv3 "github.com/envoyproxy/go-control-plane/envoy/service/auth/v3"
	typev3 "github.com/envoyproxy/go-control-plane/envoy/type/v3"
	"google.golang.org/genproto/googleapis/rpc/status"
	"google.golang.org/grpc/codes"
)

type keepOut struct {
	realm string
	user  string
	pass  string
}

func (k *keepOut) unauthorizedResponse() *authv3.CheckResponse {
	return &authv3.CheckResponse{
		Status: &status.Status{
			Code: int32(codes.OK),
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
	if request == nil {
		return k.unauthorizedResponse(), nil
	}

	attr := request.GetAttributes()
	if attr == nil {
		return k.unauthorizedResponse(), nil
	}

	req := attr.GetRequest()
	if req == nil {
		return k.unauthorizedResponse(), nil
	}

	http := req.GetHttp()
	if http == nil {
		return k.unauthorizedResponse(), nil
	}

	headers := http.GetHeaders()
	if headers == nil {
		return k.unauthorizedResponse(), nil
	}

	authHeader, ok := headers["authorization"]
	if !ok {
		return k.unauthorizedResponse(), nil
	}

	if !strings.HasPrefix(strings.ToLower(authHeader), "basic ") {
		return k.unauthorizedResponse(), nil
	}

	encodedAuth := strings.TrimSpace(authHeader[5:])

	auth, err := base64.StdEncoding.DecodeString(encodedAuth)
	if err != nil {
		return k.unauthorizedResponse(), nil
	}

	if !bytes.ContainsRune(auth, ':') {
		return k.unauthorizedResponse(), nil
	}

	parts := strings.SplitN(string(auth), ":", 2)
	user, pass := parts[0], parts[1]

	if user != k.user {
		return k.unauthorizedResponse(), nil
	}

	if pass != k.pass {
		return k.unauthorizedResponse(), nil
	}

	return &authv3.CheckResponse{
		Status: &status.Status{
			Code: int32(codes.OK),
		},
	}, nil
}
