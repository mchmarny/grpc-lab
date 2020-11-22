package main

import (
	"context"
	"flag"
	"net"
	"os"
	"os/signal"

	"github.com/mchmarny/grpc-lab/pkg/config"
	"github.com/mchmarny/grpc-lab/pkg/service"
	log "github.com/sirupsen/logrus"
)

var (
	address = config.GetEnvVar("ADDRESS", ":50505")
	debug   = config.GetEnvBoolVar("DEBUG", false)
)

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

	srv := service.NewPingService(lis)

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
