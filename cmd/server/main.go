package main

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"flag"
	"io/ioutil"
	"net"
	"os"
	"sync"

	"github.com/mchmarny/grpc-lab/pkg/id"
	pb "github.com/mchmarny/grpc-lab/pkg/proto/v1"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/reflection"
)

var (
	caPath   = flag.String("ca", "", "Path to file containing the CA root cert file")
	certPath = flag.String("cert", "", "Path to TLS cert file")
	keyPath  = flag.String("key", "", "Path to TLS key file")
	address  = flag.String("address", ":50505", "The server address")
	debug    = flag.Bool("debug", false, "Verbose logging")
)

type pingServer struct {
	pb.UnimplementedServiceServer
	messageCount uint64
	lock         sync.Mutex
}

func (s *pingServer) Ping(ctx context.Context, req *pb.PingRequest) (res *pb.PingResponse, err error) {
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
		Reversed: reverse(req.Message),
		Count:    s.messageCount,
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

func getCredentials() (credentials.TransportCredentials, error) {
	if *certPath == "" || *keyPath == "" || *caPath == "" {
		return nil, errors.New("missing certificates")
	}

	log.Infof("using TLS (ca:%s, cert:%s, key:%s)", *caPath, *certPath, *keyPath)

	ca, err := ioutil.ReadFile(*caPath)
	if err != nil {
		return nil, errors.Wrapf(err, "error reading ca file: %s", *caPath)
	}

	certPool := x509.NewCertPool()
	if !certPool.AppendCertsFromPEM(ca) {
		return nil, errors.New("error adding client CA")
	}

	serverCert, err := tls.LoadX509KeyPair(*certPath, *keyPath)
	if err != nil {
		return nil, errors.Wrapf(err, "error loading cert (%s) and key files (%s)", *certPath, *keyPath)
	}

	config := &tls.Config{
		Certificates: []tls.Certificate{serverCert},
		ClientAuth:   tls.RequireAndVerifyClientCert,
		ClientCAs:    certPool,
	}

	return credentials.NewTLS(config), nil
}

func startServer(lis net.Listener) error {
	opts := []grpc.ServerOption{}
	if *certPath != "" && *keyPath != "" {
		creds, err := getCredentials()
		if err != nil {
			return errors.Wrapf(err, "error getting credentials (cert:%s, key:%s) %v", *certPath, *keyPath, err)
		}
		opts = append(opts, grpc.Creds(creds))
	}

	srv := &pingServer{}
	grpcServer := grpc.NewServer(opts...)
	reflection.Register(grpcServer)
	pb.RegisterServiceServer(grpcServer, srv)

	log.Infof("starting gRPC server: %s", lis.Addr().String())
	return grpcServer.Serve(lis)
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

	if err := startServer(lis); err != nil {
		log.Fatal(err)
	}
}
