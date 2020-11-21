package main

import (
	"context"
	"flag"
	"io"
	"net"
	"os"
	"os/signal"
	"sync"
	"time"

	"github.com/mchmarny/grpc-lab/pkg/config"
	"github.com/mchmarny/grpc-lab/pkg/format"
	pb "github.com/mchmarny/grpc-lab/pkg/proto/v1"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

var (
	address = config.GetEnvVar("ADDRESS", ":50505")
	debug   = config.GetEnvBoolVar("DEBUG", false)
)

// PingServer represents the server that responds to pings
type PingServer struct {
	pb.UnimplementedServiceServer
	messageCount int64
	lock         sync.Mutex
	listener     net.Listener
}

// Start starts the ping server
func (s *PingServer) Start(ctx context.Context) error {
	opts := []grpc.ServerOption{}
	grpcServer := grpc.NewServer(opts...)
	reflection.Register(grpcServer)
	pb.RegisterServiceServer(grpcServer, s)

	log.Infof("starting gRPC server: %s", s.listener.Addr().String())
	return grpcServer.Serve(s.listener)
}

// Close closes ping server
func (s *PingServer) Close() {
	if err := s.listener.Close(); err != nil {
		log.Warnf("error closing ping server: %v", err)
	}
}

// Stream stream messages
func (s *PingServer) Stream(stream pb.Service_StreamServer) error {
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
func (s *PingServer) Ping(ctx context.Context, req *pb.PingRequest) (res *pb.PingResponse, err error) {
	if req == nil {
		return nil, errors.New("nil request")
	}
	res = s.processReq(req)
	return
}

func (s *PingServer) processReq(req *pb.PingRequest) *pb.PingResponse {
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

func main() {
	flag.Parse()
	log.SetFormatter(&log.JSONFormatter{})
	log.SetOutput(os.Stdout)
	log.SetLevel(log.WarnLevel)
	if debug {
		log.SetLevel(log.TraceLevel)
	}

	lis, err := net.Listen("tcp", address)
	if err != nil {
		log.Fatalf("error creating listener on %s: %v", address, err)
	}

	srv := &PingServer{
		listener: lis,
	}

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt)
	ctx, cancel := context.WithCancel(context.Background())

	go func() {
		<-sigCh
		cancel()
		srv.Close()
		os.Exit(0)
	}()

	if err := srv.Start(ctx); err != nil && err.Error() != "closed" {
		log.Fatal(err)
	}
}
