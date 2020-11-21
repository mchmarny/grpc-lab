package main

import (
	"context"
	"testing"
	"time"

	"github.com/mchmarny/grpc-lab/pkg/config"
	pb "github.com/mchmarny/grpc-lab/pkg/proto/v1"

	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc/test/bufconn"
)

func TestPing(t *testing.T) {
	ctx := context.Background()
	srv := getTestServer()
	assert.NotNil(t, srv)
	startTestServer(ctx, t, srv)
	defer stopTestServer(t, srv)

	t.Run("ping sans args", func(t *testing.T) {
		if _, err := srv.Ping(ctx, nil); err == nil {
			t.FailNow()
		}
	})

	t.Run("ping with args", func(t *testing.T) {
		req := getTestRequest()
		pingCtx, cancel := context.WithTimeout(ctx, 3*time.Second)
		defer cancel()
		resp, err := srv.Ping(pingCtx, req)
		if err != nil {
			t.Errorf("error on ping: %v", err)
		}
		assert.NotNil(t, resp)
		assert.Exactly(t, "test", resp.Message)
		assert.Exactly(t, "tset", resp.Reversed)
	})
	t.Run("ping count", func(t *testing.T) {
		req := getTestRequest()
		pingCtx, cancel := context.WithTimeout(ctx, 3*time.Second)
		defer cancel()

		resp1, err := srv.Ping(pingCtx, req)
		if err != nil {
			t.Errorf("error on ping: %v", err)
		}
		assert.NotNil(t, resp1)

		resp2, err := srv.Ping(pingCtx, req)
		if err != nil {
			t.Errorf("error on ping: %v", err)
		}
		assert.NotNil(t, resp2)
		assert.True(t, resp2.Count > resp1.Count)
	})
}

func getTestRequest() *pb.PingRequest {
	return &pb.PingRequest{
		Id:      "test-id",
		Message: "test",
		Metadata: map[string]string{
			"client-id":  "test",
			"created-on": time.Now().UTC().Format(time.RFC3339),
		},
	}
}

func getTestServer() *PingServer {
	c := &config.Config{}
	return &PingServer{
		listener: bufconn.Listen(1024 * 1024),
		config:   c,
	}
}

func startTestServer(ctx context.Context, t *testing.T, srv *PingServer) {
	go func() {
		if err := srv.Start(ctx); err != nil && err.Error() != "closed" {
			t.Fatalf("error starting server: %v", err)
		}
	}()
}

func stopTestServer(t *testing.T, srv *PingServer) {
	assert.NotNil(t, srv)
	srv.Close()
}
