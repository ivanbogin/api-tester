package main

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/zach-klippenstein/goregen"
	"gopkg.in/redis.v3"
)

var (
	templates *template.Template
	rclient   *redis.Client
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
	if len(inbox) != 8 {
		return
	}

	key := "inbox:" + inbox

	exists, _ := rclient.Exists(key).Result()
	if exists == false {
		w.Write([]byte("Inbox expired"))
		return
	}

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
	inbox := r.URL.Path[len("/in/"):]
	if len(inbox) != 8 {
		return
	}

	key := "inbox:" + inbox

	rclient.HIncrBy(key, "requests", 1)
	rclient.Expire(key, 1*time.Hour)

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

	templates = template.Must(template.ParseFiles("templates/home.html", "templates/view.html"))

	redisDb, _ := strconv.ParseInt(os.Getenv("REDIS_DB"), 10, 64)
	rclient = redis.NewClient(&redis.Options{
		Addr:     os.Getenv("REDIS_ADDR"),
		Password: os.Getenv("REDIS_PASS"),
		DB:       redisDb,
	})

	_, err := rclient.Ping().Result()
	if err != nil {
		log.Fatal(err)
	}

	http.HandleFunc("/", homeHandler)
	http.HandleFunc("/view/", viewHandler)
	http.HandleFunc("/in/", inboxHandler)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}
