package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"

	app "github.com/Bronn1/s3-test/object-storage/application"
	storagepb "github.com/Bronn1/s3-test/proto"
	"google.golang.org/grpc"
)

func main() {
	port := flag.Int("port", 1455, "port to start server")
	flag.Parse()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	grpcServer := grpc.NewServer()
	storagepb.RegisterObjectStorageServer(grpcServer, app.NewServer(fmt.Sprintf("./data_%d", *port)))
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", *port))
	if err != nil {
		log.Fatal(err.Error())
	}
	interruptCh := make(chan os.Signal, 1)
	shutdownSignals := []os.Signal{
		os.Interrupt,
		syscall.SIGTERM,
		syscall.SIGINT,
		syscall.SIGQUIT,
	}
	signal.Notify(interruptCh, shutdownSignals...)
	go func() {
		if err := grpcServer.Serve(lis); err != nil {
			log.Fatal(err.Error())
		}
	}()

	select {
	case <-ctx.Done():
	case <-interruptCh:
		log.Print("Stopping server")
		cancel()
	}
}
