package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"strings"

	"github.com/mchmarny/grpc-lab/pkg/client"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

var (
	address  = flag.String("address", ":50505", "Server address (:50505)")
	clientID = flag.String("client", "demo", "ID of this client")
	debug    = flag.Bool("debug", false, "Verbose logging")
)

func run(ctx context.Context, c *client.PingClient) error {
	if c == nil {
		return errors.New("client required")
	}

	var msg string
	scanner := bufio.NewScanner(os.Stdin)
	scanner.Split(bufio.ScanBytes)

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
			// exit promots
			return nil
		}

		revMsg, msgCount, err := c.Ping(ctx, msg)
		if err != nil {
			fmt.Printf("error: %v", err)
			continue
		}

		fmt.Printf("%v - #%d\n", revMsg, msgCount)
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

	c, err := client.NewPingClient(ctx, *address, *clientID)
	if err != nil {
		log.Fatalf("error creating client: %v", err)
	}

	if err := run(ctx, c); err != nil {
		log.Fatalf("error starting: %v", err)
	}
	log.Info("done")
}
