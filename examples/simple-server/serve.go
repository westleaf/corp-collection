package main

import "net/http"

func Serve(dirname, port string) error {
	http.Handle("/", http.FileServer(http.Dir(dirname)))
	return http.ListenAndServe(":"+port, nil)
}
