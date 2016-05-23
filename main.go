package main

import (
	"fmt"
	"github.com/zach-klippenstein/goregen"
	"html/template"
	"log"
	"net/http"
	"os"
	"time"
)

var templates = template.Must(template.ParseFiles("templates/home.html", "templates/view.html"))

func HomeHandler(w http.ResponseWriter, r *http.Request) {
	inbox, _ := regen.Generate("[a-z0-9]{8}")
	data := struct{ Inbox string }{inbox}
	err := templates.ExecuteTemplate(w, "home.html", data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func ViewHandler(w http.ResponseWriter, r *http.Request) {
	inbox := r.URL.Path[len("/view/"):]
	data := struct {
		Inbox string
		Url   string
	}{
		inbox,
		getInboxUrl(r.URL.Scheme, r.Host, inbox),
	}
	err := templates.ExecuteTemplate(w, "view.html", data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func InboxHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Content-type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "%d", time.Now().UnixNano())
}

func getInboxUrl(scheme string, host string, inbox string) string {
	if len(scheme) == 0 {
		scheme = "http"
	}
	return fmt.Sprintf("%s://%s/in/%s", scheme, host, inbox)
}

func main() {
	port := os.Getenv("PORT")

	if port == "" {
		log.Fatal("$PORT must be set")
	}

	http.HandleFunc("/", HomeHandler)
	http.HandleFunc("/view/", ViewHandler)
	http.HandleFunc("/in/", InboxHandler)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}
