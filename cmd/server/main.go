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
	address  = config.GetEnvVar("ADDRESS", "0.0.0.0")
	grpcPort = config.GetEnvVar("GRPC_PORT", "50505")
	httpPort = config.GetEnvVar("HTTP_PORT", "")
	debug    = config.GetEnvBoolVar("DEBUG", false)
)

func main() {
	flag.Parse()
	log.SetFormatter(&log.JSONFormatter{})
	log.SetOutput(os.Stdout)
	log.SetLevel(log.WarnLevel)
	if debug {
		log.SetLevel(log.TraceLevel)
	}

	addr := net.JoinHostPort(address, grpcPort)
	lis, err := net.Listen("tcp", addr)
	if err != nil {
		log.Fatalf("error creating listener on %s: %v", addr, err)
	}
	defer lis.Close()

	srv := service.NewPingService(lis)

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt)
	ctx, cancel := context.WithCancel(context.Background())

	exitCh := make(chan error, 1)

	go func() {
		if err := srv.Start(ctx); err != nil && err.Error() != "closed" {
			log.Error("grpc server error")
			exitCh <- err
		}
		exitCh <- nil
	}()

	if httpPort != "" {
		go func() {
			addr := net.JoinHostPort(address, httpPort)
			if err := srv.StartHTTP(ctx, addr); err != nil && err.Error() != "closed" {
				log.Error("http server error")
				exitCh <- err
			}
			exitCh <- nil
		}()
	}

	for {
		select {
		case <-sigCh:
			cancel()
			os.Exit(0)
		case err := <-exitCh:
			if err != nil {
				log.Error(err)
				os.Exit(1)
				break
			}
			os.Exit(0)
		}
	}
}
