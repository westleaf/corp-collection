package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
)

type Handler struct{}

var ApiUrl string = "/api/HttpExample"

func (h *Handler) helloHandler(w http.ResponseWriter, r *http.Request) {
	message := "This trigger was executed successfully!"
	name := r.URL.Query().Get("name")
	if name != "" {
		message = fmt.Sprintf("Hello, %s. This trigger was executed successfully\n", name)
	}
	fmt.Fprint(w, message)
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	uri := r.URL.Path
	log.Default().Println(uri)
	if uri != "/api/HttpExample" {
		http.Redirect(w, r, "/api/HttpExample", http.StatusSeeOther)
		return
	} else {
		h.helloHandler(w, r) // Call the helloHandler for /api/HttpExample
		return
	}
}

func main() {
	listenAddr := ":8080"
	if val, ok := os.LookupEnv("FUNCTIONS_CUSTOMHANDLER_PORT"); ok {
		listenAddr = ":" + val
	}
	handler := new(Handler)
	http.HandleFunc("/api/HttpExample", handler.helloHandler)

	log.Printf("About to listen on %s. Go to https://127.0.0.1%s/", listenAddr, listenAddr)
	log.Fatal(http.ListenAndServe(listenAddr, handler))
}
