package main

import (
	"context"
	"flag"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	cfg "github.com/Bronn1/s3-test/config"
	app "github.com/Bronn1/s3-test/uploader/application"
)

func main() {
	confPath := flag.String("config", "./config/config.json", "config to start server")
	flag.Parse()
	conf, err := cfg.NewConfig(*confPath)
	if err != nil {
		log.Fatal("Cannot read config: ", err.Error())
	}

	serverConnections, err := app.NewSimpleBalancerRoundR(conf.Storages)
	if err != nil {
		log.Fatal("Cannot connect to storages: ", err.Error())
	}
	ser := app.New(
		app.WithLogger(),
		app.WithFileMetaService(app.NewFileMetaServiceWithMock("./mock_db.json")),
		app.WithStorageServers(serverConnections),
	)
	http.HandleFunc("/upload", ser.HandleUpload)
	http.HandleFunc("/get", ser.HandleGetFile)

	interruptCh := make(chan os.Signal, 1)
	shutdownSignals := []os.Signal{
		os.Interrupt,
		syscall.SIGTERM,
		syscall.SIGINT,
		syscall.SIGQUIT,
	}
	signal.Notify(interruptCh, shutdownSignals...)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go func() {
		log.Printf("Starting server on port 8081")
		if err := http.ListenAndServe(":8081", nil); err != nil {
			log.Fatal(err.Error())
		}
	}()

	select {
	case <-ctx.Done():
	case <-interruptCh:
		log.Print("Stopping server")
		cancel()
		ser.StorageConnections.Close()
	}
}
