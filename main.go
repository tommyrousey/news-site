package main

import (
	"http/template"
	"net/http"
	"os"
)

// create a html template output
// the must makes sure execution does not continue with a bad template
var tpl = template.Must(template.ParseFiles("index.html"))

func indexHandler(w http.ResponseWriter, r *http.Request) {
	// generates html safe from code injection
	tpl.Execute(w, nil)
}

func main() {
	// get a port
	port := os.Getenv("PORT")
	if port == "" {
		port = "3000"
	}

	// make a server mux
	mux := http.NewServeMux()

	// give the mux the indexHandler function
	mux.HandleFunc("/", indexHandler)
	// listen on the given port with the mux
	http.ListenAndServe(":"+port, mux)

}
