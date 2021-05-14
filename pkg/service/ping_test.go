package service

import (
	"context"
	"log"
	"testing"
	"time"

	pb "github.com/mchmarny/grpc-lab/pkg/api/v1"

	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc/test/bufconn"
)

func TestPing(t *testing.T) {
	ctx := context.Background()
	srv := startTestServer(ctx)
	assert.NotNil(t, srv)

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
		assert.Exactly(t, req.Content.Id, resp.MessageID)
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
		assert.True(t, resp2.MessageCount > resp1.MessageCount)
	})
}

func getTestRequest() *pb.PingRequest {
	return &pb.PingRequest{
		Sent: time.Now().UTC().UnixNano(),
		Content: &pb.Content{
			Id:   "test-id",
			Data: []byte("test"),
			Metadata: map[string]string{
				"client-id": "test",
			},
		},
	}
}

func startTestServer(ctx context.Context) *PingService {
	list := bufconn.Listen(1024 * 1024)
	defer list.Close()
	srv := NewPingService(list)

	go func() {
		if err := srv.Start(ctx); err != nil && err.Error() != "closed" {
			log.Fatalf("error starting server: %v", err)
		}
	}()

	return srv
}
