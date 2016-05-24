package main

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
	"net/http/httputil"
	"os"
	"strconv"
	"time"

	"github.com/zach-klippenstein/goregen"
	"gopkg.in/redis.v3"
	"io/ioutil"
)

const (
	keylen int = 8
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
	if len(inbox) != keylen {
		return
	}

	inboxKey := "inbox:" + inbox

	exists, _ := rclient.Exists(inboxKey).Result()
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
	if len(inbox) != keylen {
		return
	}

	inboxKey := "inbox:" + inbox

	rclient.Expire(inboxKey, 1 * time.Hour)

	record, err := dumpRequest(r)
	if err != nil {
		return
	}

	rclient.RPush(inboxKey, record)

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

func dumpBody(r *http.Request) (string, error) {
	bodyBytes, err := ioutil.ReadAll(r.Body)
	return string(bodyBytes), err
}

func dumpRequest(r *http.Request) (string, error) {
	dump, err := httputil.DumpRequest(r, true)
	if err != nil {
		return "", err
	}
	return string(dump), nil
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
	log.Fatal(http.ListenAndServe(":" + port, nil))
}
