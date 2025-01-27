package main

import (
	"flag"
	"log"
	"path/filepath"
)

var (
	path = flag.String("path", ".", "path to serve files from")
	port = flag.String("port", "8080", "port to listen on")
)

func main() {
	flag.Parse()

	dirname, err := filepath.Abs(*path)
	if err != nil {
		log.Fatalf("failed to get absolute path: %v", err)
	}

	log.Printf("serving files from %s", dirname)

	err = Serve(dirname, *port)
	if err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
