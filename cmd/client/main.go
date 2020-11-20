package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"time"

	"github.com/mchmarny/grpc-lab/pkg/creds"
	"github.com/mchmarny/grpc-lab/pkg/id"
	pb "github.com/mchmarny/grpc-lab/pkg/proto/v1"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc"
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

func start(ctx context.Context, target, clientID string, c *creds.Config) error {
	if c == nil {
		return errors.New("config required")
	}
	transportOption := grpc.WithInsecure()
	if c.HasCerts() {
		creds, err := creds.GetClientCredentials(c)
		if err != nil {
			return errors.Wrapf(err, "error getting credentials (cert:%s, key:%s)", c.Cert, c.Key)
		}
		transportOption = grpc.WithTransportCredentials(creds)
	}

	log.Infof("dialing: %s...)", target)
	conn, err := grpc.Dial(target, transportOption)
	if err != nil {
		return errors.Wrap(err, "error dialling")
	}
	defer conn.Close()
	log.Info("connected")

	var msg string
	scanner := bufio.NewScanner(os.Stdin)
	scanner.Split(bufio.ScanBytes)
	client := pb.NewServiceClient(conn)

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
			return nil
		}

		req := &pb.PingRequest{
			Id:      id.NewID(),
			Message: msg,
			Metadata: map[string]string{
				"client-id":  clientID,
				"created-on": time.Now().UTC().Format(time.RFC3339),
				"host":       c.Host,
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

func main() {
	flag.Parse()
	log.SetFormatter(&log.JSONFormatter{})
	log.SetOutput(os.Stdout)
	log.SetLevel(log.WarnLevel)
	if *debug {
		log.SetLevel(log.TraceLevel)
	}

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt)
	ctx, cancel := context.WithCancel(context.Background())

	go func() {
		<-sigCh
		cancel()
		os.Exit(0)
	}()

	c := &creds.Config{
		CA:   *caPath,
		Cert: *certPath,
		Key:  *keyPath,
		Host: *host,
	}

	start(ctx, *address, *clientID, c)
	log.Info("done")
}
