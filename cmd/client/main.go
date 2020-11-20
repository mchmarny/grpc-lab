package main

import (
	"bufio"
	"context"
	"crypto/tls"
	"crypto/x509"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"os/signal"
	"strings"
	"time"

	"github.com/mchmarny/grpc-lab/pkg/id"
	pb "github.com/mchmarny/grpc-lab/pkg/proto/v1"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

var (
	caPath   = flag.String("ca", "", "Path to file containing the CA root cert file")
	certPath = flag.String("cert", "", "Path to TLS cert file")
	keyPath  = flag.String("key", "", "Path to TLS key file")
	address  = flag.String("address", ":50505", "Server address (:50505)")
	host     = flag.String("host", "demo.grpc.thingz.io", "Hostname returned in TLS handshake (demo.grpc.thingz.io)")
	clientID = flag.String("client", "demo", "ID of this client")
	debug    = flag.Bool("debug", false, "Verbose logging")
)

func start(ctx context.Context, client pb.ServiceClient) {
	scanner := bufio.NewScanner(os.Stdin)
	scanner.Split(bufio.ScanBytes)

	var msg string

	for {
		fmt.Print("message (enter to exit): ")
		for scanner.Scan() {
			if scanner.Text() == "\n" {
				break
			} else {
				msg += scanner.Text()
			}
		}
		if strings.TrimSpace(msg) == "" {
			break
		}

		req := &pb.PingRequest{
			Id:      id.NewID(),
			Message: msg,
			Metadata: map[string]string{
				"client-id":  *clientID,
				"created-on": time.Now().UTC().Format(time.RFC3339),
			},
		}

		pingCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
		defer cancel()

		resp, err := client.Ping(pingCtx, req)
		if err != nil {
			fmt.Printf("error on ping: %v", err)
			continue
		}

		fmt.Printf("%v - #%d\n", resp.Reversed, resp.Count)
		fmt.Println()
		msg = ""
	}
}

func getCredentials() (credentials.TransportCredentials, error) {
	if *certPath == "" || *keyPath == "" || *caPath == "" {
		return nil, errors.New("missing certificates")
	}

	ca, err := ioutil.ReadFile(*caPath)
	if err != nil {
		return nil, errors.Wrapf(err, "error reading ca file: %s", *caPath)
	}

	certPool := x509.NewCertPool()
	if !certPool.AppendCertsFromPEM(ca) {
		return nil, errors.New("error adding client CA")
	}

	clientCert, err := tls.LoadX509KeyPair(*certPath, *keyPath)
	if err != nil {
		return nil, err
	}

	// Create the credentials and return it
	config := &tls.Config{
		Certificates: []tls.Certificate{clientCert},
		RootCAs:      certPool,
		ServerName:   *host,
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

	transportOption := grpc.WithInsecure()
	if *caPath != "" {
		creds, err := getCredentials()
		if err != nil {
			log.Fatalf("error getting credentials (cert:%s, key:%s) %v", *certPath, *keyPath, err)
		}

		transportOption = grpc.WithTransportCredentials(creds)
	}

	log.Infof("dialing: %s...)", *address)
	conn, err := grpc.Dial(*address, transportOption)
	if err != nil {
		log.Fatalf("error dialling: %v", err)
	}
	defer conn.Close()
	log.Infof("connected")

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	ctx, cancel := context.WithCancel(context.Background())

	go func() {
		<-c
		cancel()
		os.Exit(0)
	}()

	start(ctx, pb.NewServiceClient(conn))
	log.Info("done")
}
