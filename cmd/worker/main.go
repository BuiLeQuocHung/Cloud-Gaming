package main

import (
	"cloud_gaming/pkg/worker"
	"log"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	w, err := worker.New()
	if err != nil {
		log.Fatal(err)
	}
	w.Run()

	done := make(chan os.Signal, 1)
	signal.Notify(done, syscall.SIGINT, syscall.SIGTERM)
	<-done
}
