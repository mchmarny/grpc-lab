package main

import (
	"context"
	"flag"
	"net"
	"os"
	"os/signal"
	"sync"

	"github.com/mchmarny/grpc-lab/pkg/creds"
	"github.com/mchmarny/grpc-lab/pkg/id"
	pb "github.com/mchmarny/grpc-lab/pkg/proto/v1"
	"github.com/mchmarny/grpc-lab/pkg/string"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

var (
	caPath   = flag.String("ca", "", "Path to file containing the CA root cert file")
	certPath = flag.String("cert", "", "Path to TLS cert file")
	keyPath  = flag.String("key", "", "Path to TLS key file")
	address  = flag.String("address", ":50505", "The server address")
	debug    = flag.Bool("debug", false, "Verbose logging")
)

// PingServer represents the server that responds to pings
type PingServer struct {
	pb.UnimplementedServiceServer
	messageCount uint64
	lock         sync.Mutex
	listener     net.Listener
	config       *creds.Config
}

// Start starts the ping server
func (s *PingServer) Start(ctx context.Context) error {
	opts := []grpc.ServerOption{}
	if s.config.HasCerts() {
		creds, err := creds.GetServerCredentials(s.config)
		if err != nil {
			return errors.Wrapf(err, "error getting credentials (%+v): %v", s.config, err)
		}
		opts = append(opts, grpc.Creds(creds))
	}

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

// Ping performs ping
func (s *PingServer) Ping(ctx context.Context, req *pb.PingRequest) (res *pb.PingResponse, err error) {
	if req == nil {
		return nil, errors.New("nil request")
	}
	log.Infof("%+v", req)

	s.lock.Lock()
	s.messageCount++
	s.lock.Unlock()

	res = &pb.PingResponse{
		Id:       id.NewID(),
		Message:  req.Message,
		Reversed: string.ReverseString(req.Message),
		Count:    s.messageCount,
	}
	return
}

func main() {
	flag.Parse()
	log.SetFormatter(&log.JSONFormatter{})
	log.SetOutput(os.Stdout)
	log.SetLevel(log.WarnLevel)
	if *debug {
		log.SetLevel(log.TraceLevel)
	}

	c := &creds.Config{
		CA:   *caPath,
		Cert: *certPath,
		Key:  *keyPath,
		Host: *address,
	}
	if !c.HasHost() {
		log.Fatal("host required")
	}

	lis, err := net.Listen("tcp", c.Host)
	if err != nil {
		log.Fatalf("error creating listener on %s: %v", c.Host, err)
	}

	srv := &PingServer{
		listener: lis,
		config:   c,
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

	if err := srv.Start(ctx); err != nil {
		log.Fatal(err)
	}
}
