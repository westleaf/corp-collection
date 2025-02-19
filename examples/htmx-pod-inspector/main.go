package main

import (
	"log"
	"net/http"
	"text/template"
)

type Pods struct {
	Id  int
	Pod string
}

func main() {
	data := map[string][]Pods{
		"Todos": {
			Pods{Id: 1, Pod: "Milk"},
		},
	}

	handler := func(w http.ResponseWriter, r *http.Request) {
		templ := template.Must(template.ParseFiles("index.html"))
		templ.Execute(w, data)
	}

	addTodoHandler := func(w http.ResponseWriter, r *http.Request) {
		pods := r.PostFormValue("pods")
		templ := template.Must(template.ParseFiles("index.html"))
		pod := Pods{Id: len(data["Pods"]) + 1, Pod: pods}
		data["Pods"] = append(data["Pods"], pod)

		templ.ExecuteTemplate(w, "pod-list-element", pod)
	}

	http.HandleFunc("/", handler)
	http.HandleFunc("/getpods", addTodoHandler)

	log.Fatal(http.ListenAndServe(":8000", nil))
}
