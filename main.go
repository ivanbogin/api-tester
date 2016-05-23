package main

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/zach-klippenstein/goregen"
	"gopkg.in/redis.v3"
)

var (
	templates = template.Must(template.ParseFiles("templates/home.html", "templates/view.html"))
	rclient   = redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "", // no password set
		DB:       0,  // use default DB
	})
)

func homeHandler(w http.ResponseWriter, r *http.Request) {
	inbox, _ := regen.Generate("[a-z0-9]{8}")
	data := struct{ Inbox string }{inbox}
	err := templates.ExecuteTemplate(w, "home.html", data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func viewHandler(w http.ResponseWriter, r *http.Request) {
	inbox := r.URL.Path[len("/view/"):]
	data := struct {
		Inbox string
		URL   string
	}{
		inbox,
		getInboxURL(r.URL.Scheme, r.Host, inbox),
	}
	err := templates.ExecuteTemplate(w, "view.html", data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func inboxHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Content-type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "%d", time.Now().UnixNano())
}

func getInboxURL(scheme string, host string, inbox string) string {
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

	http.HandleFunc("/", homeHandler)
	http.HandleFunc("/view/", viewHandler)
	http.HandleFunc("/in/", inboxHandler)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}
