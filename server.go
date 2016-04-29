package main

import (
	"log"
	"net/http"
	"time"
)

func HttpHandler(w http.ResponseWriter, r *http.Request) {
	status := http.StatusOK
	uri := r.URL.Path

	log.Println(uri)

	switch uri {
	case "/500":
		status = http.StatusInternalServerError
	case "/wait3":
		time.Sleep(3 * time.Second)
	case "/wait5":
		time.Sleep(5 * time.Second)
	case "/wait60":
		time.Sleep(time.Minute)
	}

	w.WriteHeader(status)
}

func main() {
	http.HandleFunc("/", HttpHandler)
	log.Fatal(http.ListenAndServe(":8080", nil))
}
