package main

import (
	"cloud_gaming/pkg/coordinator"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	c := coordinator.New()
	c.Run()

	done := make(chan os.Signal, 1)
	signal.Notify(done, syscall.SIGINT, syscall.SIGTERM)
	<-done
}
