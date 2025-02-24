package main

import (
	"log"
	"net/http"
	"text/template"
)

type Todo struct {
	Id      int
	Message string
}

func main() {
	data := map[string][]Todo{
		"Todos": {
			Todo{Id: 1, Message: "Milk"},
		},
	}

	handler := func(w http.ResponseWriter, r *http.Request) {
		templ := template.Must(template.ParseFiles("index.html"))
		templ.Execute(w, data)
	}

	addTodoHandler := func(w http.ResponseWriter, r *http.Request) {
		message := r.PostFormValue("message")
		templ := template.Must(template.ParseFiles("index.html"))
		todo := Todo{Id: len(data["Todos"]) + 1, Message: message}
		data["Todo"] = append(data["Todo"], todo)

		templ.ExecuteTemplate(w, "todo-list-element", todo)
	}

	http.HandleFunc("/", handler)
	http.HandleFunc("/add-todo", addTodoHandler)

	log.Fatal(http.ListenAndServe(":8000", nil))
}
