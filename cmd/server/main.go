package main

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"flag"
	"io/ioutil"
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
	caPath   = flag.String("ca", "", "Path to file containing the CA root cert file")
	certPath = flag.String("cert", "", "Path to TLS cert file")
	keyPath  = flag.String("key", "", "Path to TLS key file")
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

	opts := []grpc.ServerOption{}
	if *certPath != "" && *keyPath != "" {
		creds, err := getCredentials()
		if err != nil {
			log.Fatalf("error getting credentials (cert:%s, key:%s) %v", *certPath, *keyPath, err)
		}
		opts = append(opts, grpc.Creds(creds))
	}

	grpcServer := grpc.NewServer(opts...)
	pb.RegisterServiceServer(grpcServer, &pingServer{})

	log.Printf("starting server: %s", lis.Addr().String())
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("error: %v", err)
	}
	log.Info("done")
}
