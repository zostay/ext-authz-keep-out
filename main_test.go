package main

import (
	"context"
	"encoding/base64"
	"fmt"
	"io"
	"log"
	"os"
	"strings"
	"testing"

	authv3 "github.com/envoyproxy/go-control-plane/envoy/service/auth/v3"
	typev3 "github.com/envoyproxy/go-control-plane/envoy/type/v3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
)

func TestMain(m *testing.M) {
	log.SetOutput(io.Discard)
	startTestServer()
	code := m.Run()
	stopTestServer()
	os.Exit(code)
}

func encodeAuthToken(user, pass string) string {
	authPair := fmt.Sprintf("%s:%s", user, pass)
	encAuthPair := base64.StdEncoding.EncodeToString([]byte(authPair))
	return fmt.Sprintf("Basic %s", encAuthPair)
}

var (
	testUser = "test123"
	testPass = "test321"

	authClient authv3.AuthorizationClient
)

func startTestServer() {
	grpcPort := make(chan int)
	ko := NewKeepOut("Mordac the Preventer", testUser, testPass)
	go ko.Run("localhost:0", grpcPort)

	portNumber := <-grpcPort
	conn, err := grpc.Dial(
		fmt.Sprintf("localhost:%d", portNumber),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)

	if err != nil {
		panic(fmt.Sprintf("unable to start grpc client: %v", err))
	}

	authClient = authv3.NewAuthorizationClient(conn)
}

func stopTestServer() {}

func TestGoodToken(t *testing.T) {
	goodToken := encodeAuthToken(testUser, testPass)

	ctx := context.Background()
	resp, err := authClient.Check(ctx, &authv3.CheckRequest{
		Attributes: &authv3.AttributeContext{
			Request: &authv3.AttributeContext_Request{
				Http: &authv3.AttributeContext_HttpRequest{
					Host: "localhost",
					Path: "/check",
					Headers: map[string]string{
						"authorization": goodToken,
					},
				},
			},
		},
	})

	require.NoError(t, err, "check good token error")
	assert.Equal(t, int32(codes.OK), resp.GetStatus().GetCode(), "valid response")
}

func TestBadToken(t *testing.T) {
	badToken := encodeAuthToken("not", "good")

	ctx := context.Background()
	resp, err := authClient.Check(ctx, &authv3.CheckRequest{
		Attributes: &authv3.AttributeContext{
			Request: &authv3.AttributeContext_Request{
				Http: &authv3.AttributeContext_HttpRequest{
					Host: "localhost",
					Path: "/check",
					Headers: map[string]string{
						"Authorization": badToken,
					},
				},
			},
		},
	})

	require.NoError(t, err, "check good token error")
	assert.Equal(t, int32(codes.PermissionDenied), resp.GetStatus().GetCode(), "valid response")
	assert.Equal(t, typev3.StatusCode_Unauthorized, resp.GetDeniedResponse().GetStatus().GetCode(), "failed authentication")

	hdr := resp.GetDeniedResponse().GetHeaders()
	require.Len(t, hdr, 1, "returned one header")
	assert.Equal(t, "www-authenticate", strings.ToLower(hdr[0].Header.Key), "header key is WWW-Authenticate")
	assert.Equal(t, hdr[0].Header.Value, `Basic realm="Mordac the Preventer"`, "header value is Mordac the Preventer's realm")
}
