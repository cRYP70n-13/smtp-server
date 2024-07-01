package main

import (
	"log"
	"smtp/server"
)

func main() {
    s := server.NewServer("0.0.0.0:25")
    if err := s.ListenAndAccept(); err != nil {
        log.Fatal(err)
    }

    select {} // Block this because we are running the server in a separated go routine
}
