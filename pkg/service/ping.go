package service

import (
	"context"
	"io"
	"net"
	"sync"
	"time"

	pb "github.com/mchmarny/grpc-lab/pkg/api/v1"
	"github.com/mchmarny/grpc-lab/pkg/format"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

// NewPingService creates an instance of the PingService
func NewPingService(lis net.Listener) *PingService {
	return &PingService{
		listener: lis,
	}
}

// PingService represents the server that responds to pings
type PingService struct {
	pb.UnimplementedServiceServer
	messageCount int64
	lock         sync.Mutex
	listener     net.Listener
}

// Start starts the ping server
func (s *PingService) Start(ctx context.Context) error {
	opts := []grpc.ServerOption{}
	grpcServer := grpc.NewServer(opts...)
	reflection.Register(grpcServer)
	pb.RegisterServiceServer(grpcServer, s)

	log.Infof("starting gRPC server: %s", s.listener.Addr().String())
	return grpcServer.Serve(s.listener)
}

// Close closes ping server
func (s *PingService) Close() {
	if err := s.listener.Close(); err != nil {
		log.Warnf("error closing ping server: %v", err)
	}
}

// Stream stream messages
func (s *PingService) Stream(stream pb.Service_StreamServer) error {
	for {
		err := contextError(stream.Context())
		if err != nil {
			return err
		}

		req, err := stream.Recv()
		if err == io.EOF {
			log.Debug("no more data")
			break
		}
		if err != nil {
			return errors.Wrap(err, "error receiving stream")
		}

		res := s.processReq(req)

		err = stream.Send(res)
		if err != nil {
			return errors.Wrap(err, "error sending stream response")
		}
	}
	return nil
}

// Ping performs ping
func (s *PingService) Ping(ctx context.Context, req *pb.PingRequest) (res *pb.PingResponse, err error) {
	if req == nil {
		return nil, errors.New("nil request")
	}
	res = s.processReq(req)
	return
}

func (s *PingService) processReq(req *pb.PingRequest) *pb.PingResponse {
	log.Infof("%+v", req)

	s.lock.Lock()
	s.messageCount++
	s.lock.Unlock()

	return &pb.PingResponse{
		Id:       req.Id,
		Message:  req.Message,
		Reversed: format.ReverseString(req.Message),
		Count:    s.messageCount,
		Created:  time.Now().UTC().UnixNano(),
		Metadata: map[string]string{
			"address": s.listener.Addr().String(),
		},
	}
}

func contextError(ctx context.Context) error {
	switch ctx.Err() {
	case context.Canceled:
		return errors.New("request is canceled")
	case context.DeadlineExceeded:
		return errors.New("deadline is exceeded")
	default:
		return nil
	}
}
