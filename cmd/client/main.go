package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/mchmarny/grpc-lab/pkg/id"
	pb "github.com/mchmarny/grpc-lab/pkg/proto/v1/service"
	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

var (
	caFile   = flag.String("ca", "", "Path to file containing the CA root cert file")
	address  = flag.String("address", ":50505", "Server address (:50505)")
	host     = flag.String("host", "server.demo.thingz.io", "Name to verify hostname returned in TLS handshake")
	clientID = flag.String("client", "demo", "ID of this client")
	debug    = flag.Bool("debug", false, "Verbose logging")
)

func start(conn *grpc.ClientConn) {
	client := pb.NewServiceClient(conn)
	ctx := context.Background()

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

		resp, err := client.Ping(ctx, req)
		if err != nil {
			fmt.Printf("error on ping: %v", err)
			continue
		}
		fmt.Printf("%v\n", resp.Reversed)
		fmt.Println()
		msg = ""
	}
}

func main() {
	flag.Parse()
	log.SetFormatter(&log.JSONFormatter{})
	log.SetOutput(os.Stdout)
	log.SetLevel(log.WarnLevel)
	if *debug {
		log.SetLevel(log.TraceLevel)
	}

	var opts []grpc.DialOption
	if *caFile != "" {
		creds, err := credentials.NewClientTLSFromFile(*caFile, *host)
		if err != nil {
			log.Fatalf("error creating TLS credentials %v", err)
		}
		opts = append(opts, grpc.WithTransportCredentials(creds))
	} else {
		opts = append(opts, grpc.WithInsecure())
	}

	opts = append(opts, grpc.WithBlock())
	conn, err := grpc.Dial(*address, opts...)
	if err != nil {
		log.Fatalf("error dialling: %v", err)
	}
	defer conn.Close()

	start(conn)
	log.Info("done")
}
