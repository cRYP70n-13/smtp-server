package main

import (
	"log"
	"os"
	"os/signal"
	"smtp/server"
	"syscall"
)

func main() {
	s := server.NewServer("0.0.0.0:25")

	go func() {
		if err := s.ListenAndAccept(); err != nil {
			log.Fatal(err)
		}
	}()

    quit := make(chan os.Signal, 1)
    signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
    <-quit

    log.Println("Shutting down the server")
    if err := s.Stop(); err != nil {
        log.Fatal(err)
    }
}
