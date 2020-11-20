package main

import (
	"context"
	"flag"
	"net"
	"os"

	"github.com/mchmarny/grpc-lab/pkg/id"
	pb "github.com/mchmarny/grpc-lab/pkg/proto/v1/service"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

var (
	certFile = flag.String("cert", "", "Path to TLS cert file")
	keyFile  = flag.String("key", "", "Path to TLS key file")
	address  = flag.String("address", ":50505", "The server address")
	debug    = flag.Bool("debug", false, "Verbose logging")
)

type pingServer struct {
	pb.UnimplementedServiceServer
}

func (s *pingServer) Ping(ctx context.Context, req *pb.PingRequest) (res *pb.PingResponse, err error) {
	if req == nil {
		return nil, errors.New("nil request")
	}
	log.Infof("%+v", req)
	res = &pb.PingResponse{
		Id:       id.NewID(),
		Message:  req.Message,
		Reversed: reverse(req.Message),
	}
	return
}

func reverse(s string) string {
	runes := []rune(s)
	for i, j := 0, len(runes)-1; i < j; i, j = i+1, j-1 {
		runes[i], runes[j] = runes[j], runes[i]
	}
	return string(runes)
}

func main() {
	flag.Parse()
	log.SetFormatter(&log.JSONFormatter{})
	log.SetOutput(os.Stdout)
	log.SetLevel(log.WarnLevel)
	if *debug {
		log.SetLevel(log.TraceLevel)
	}
	lis, err := net.Listen("tcp", *address)
	if err != nil {
		log.Fatalf("error creating listener on %s: %v", *address, err)
	}
	var opts []grpc.ServerOption
	if *certFile != "" && *keyFile != "" {
		creds, err := credentials.NewServerTLSFromFile(*certFile, *keyFile)
		if err != nil {
			log.Fatalf("error generating credentials using provided cert and key %v", err)
		}
		opts = []grpc.ServerOption{grpc.Creds(creds)}
	}
	grpcServer := grpc.NewServer(opts...)
	pb.RegisterServiceServer(grpcServer, &pingServer{})
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("error: %v", err)
	}
	log.Info("done")
}
