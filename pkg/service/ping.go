package service

import (
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	"sync"
	"time"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	pb "github.com/mchmarny/grpc-lab/pkg/api/v1"
	"github.com/mchmarny/grpc-lab/pkg/format"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

// NewPingService creates an instance of the PingService
func NewPingService(list net.Listener) *PingService {
	return &PingService{
		grpcListener: list,
	}
}

// PingService represents the server that responds to pings
type PingService struct {
	pb.UnimplementedServiceServer
	messageCount int64
	lock         sync.Mutex
	grpcListener net.Listener
}

// Start starts the ping service as a gRPC server
func (s *PingService) Start(ctx context.Context) error {
	opts := []grpc.ServerOption{}
	grpcServer := grpc.NewServer(opts...)
	reflection.Register(grpcServer)
	pb.RegisterServiceServer(grpcServer, s)

	log.Infof("starting gRPC server at: %s", s.grpcListener.Addr().String())
	return grpcServer.Serve(s.grpcListener)
}

// StartHTTP starts the ping service as a HTTP server
func (s *PingService) StartHTTP(ctx context.Context, port string) error {
	addr := fmt.Sprintf("0.0.0.0:%s", port)
	lis, err := net.Listen("tcp", addr)
	if err != nil {
		return errors.Wrapf(err, "error creating listener on %s: %v", addr, err)
	}
	defer lis.Close()

	mux := runtime.NewServeMux()
	opts := []grpc.DialOption{grpc.WithInsecure()}

	cancelCtx, cancel := context.WithCancel(ctx)
	defer cancel()

	endpoint := s.grpcListener.Addr().String()
	if err := pb.RegisterServiceHandlerFromEndpoint(cancelCtx, mux, endpoint, opts); err != nil {
		return errors.Wrap(err, "error registering HTTP handler")
	}

	log.Infof("starting REST server at %s", lis.Addr().String())
	return http.Serve(lis, mux)
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
			"address": s.grpcListener.Addr().String(),
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
