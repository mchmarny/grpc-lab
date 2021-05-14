package client

import (
	"context"
	"fmt"
	"io"
	"time"

	pb "github.com/mchmarny/grpc-lab/pkg/api/v1"
	"github.com/mchmarny/grpc-lab/pkg/id"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc"
)

const (
	timeOutInSec = 5
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
		return nil, errors.Wrap(err, "error dialing")
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

// MakeRequest creates a request from message
func (p *PingClient) MakeRequest(msg string, index int) *pb.PingRequest {
	return &pb.PingRequest{
		Sent: time.Now().UTC().UnixNano(),
		Content: &pb.Content{
			Id:   id.NewID(),
			Data: []byte(msg),
			Metadata: map[string]string{
				"client-id":     p.id,
				"created-on":    time.Now().UTC().Format(time.RFC3339),
				"message-index": fmt.Sprintf("%d", index),
			},
		},
	}
}

// Ping sends messages to the server
func (p *PingClient) Ping(ctx context.Context, msg string) (out string, count int64, err error) {
	req := p.MakeRequest(msg, 0)

	pingCtx, cancel := context.WithTimeout(ctx, timeOutInSec*time.Second)
	defer cancel()

	resp, err := p.client.Ping(pingCtx, req)
	if err != nil {
		return "", 0, errors.Wrap(err, "error on ping")
	}
	return resp.Detail, resp.MessageCount, nil
}

// StreamList streams messages from the client
func (p *PingClient) StreamList(ctx context.Context, list []string) error {
	pingCtx, cancel := context.WithTimeout(ctx, timeOutInSec*time.Second)
	defer cancel()

	stream, err := p.client.Stream(pingCtx)
	if err != nil {
		return errors.Wrap(err, "error creating stream")
	}
	waitResponse := make(chan error)
	go func() {
		for {
			res, resErr := stream.Recv()
			if resErr == io.EOF {
				log.Debug("no more responses")
				waitResponse <- nil
				return
			}
			if resErr != nil {
				waitResponse <- errors.Wrap(resErr, "error receiving stream response")
				return
			}

			log.Debugf("received response: %+v", res)
		}
	}()

	// send messages
	for i, msg := range list {
		req := p.MakeRequest(msg, i)

		sendErr := stream.Send(req)
		if sendErr != nil {
			return errors.Wrapf(sendErr, "error sending stream request: %v", stream.RecvMsg(nil))
		}

		log.Debugf("sent request: %+v", req)
	}

	closeErr := stream.CloseSend()
	if closeErr != nil {
		return errors.Wrap(closeErr, "cannot close stream")
	}

	return <-waitResponse
}

// Close cleans up resources
func (p *PingClient) Close() {
	if p.conn != nil {
		if err := p.conn.Close(); err != nil {
			log.Errorf("error closing connection: %v", err)
		}
	}
}
