package main

import (
	"fmt"
	"log"
	"net/http"
)

func handler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Leroy skibidi rizz edge!!\n")
}

func main() {
	log.Print("server ready")
	http.HandleFunc("/", handler)
	http.ListenAndServe(":50051", nil)
}
