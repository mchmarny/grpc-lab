package client

import (
	"context"
	"time"

	"github.com/mchmarny/grpc-lab/pkg/id"
	pb "github.com/mchmarny/grpc-lab/pkg/proto/v1"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc"
)

// NewPingClient creates a new instance of the ping client
func NewPingClient(ctx context.Context, target, clientID string) (client *PingClient, err error) {
	if target == "" {
		return nil, errors.New("target required")
	}
	// dialing options
	opt := grpc.WithInsecure()
	log.Infof("dialing: %s...)", target)
	conn, err := grpc.Dial(target, opt)
	if err != nil {
		return nil, errors.Wrap(err, "error dialling")
	}

	client = &PingClient{
		conn:   conn,
		client: pb.NewServiceClient(conn),
		target: target,
		id:     clientID,
	}
	return
}

// PingClient represents local version of the gRPC client
type PingClient struct {
	conn   *grpc.ClientConn
	client pb.ServiceClient
	target string
	id     string
}

// Ping sends messages to the server
func (p *PingClient) Ping(ctx context.Context, msg string) (out string, count int64, err error) {
	req := &pb.PingRequest{
		Id:      id.NewID(),
		Message: msg,
		Metadata: map[string]string{
			"client-id":  p.id,
			"created-on": time.Now().UTC().Format(time.RFC3339),
		},
	}

	pingCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	resp, err := p.client.Ping(pingCtx, req)
	if err != nil {
		return "", 0, errors.Wrap(err, "error on ping")
	}
	return resp.Reversed, resp.Count, nil
}

// Close cleans up resources
func (p *PingClient) Close() {
	if p.conn != nil {
		if err := p.conn.Close(); err != nil {
			log.Errorf("error closing connection: %v", err)
		}
	}
}
