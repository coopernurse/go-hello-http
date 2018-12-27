package main

import (
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"
)

func main() {
	http.HandleFunc("/", HomeHandler)
	http.HandleFunc("/fail", FailHandler)
	http.HandleFunc("/sleep", SleepHandler)
	log.Fatal(http.ListenAndServe(":8080", nil))

}

func HomeHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Hello, %s!", r.URL.Path[1:])
}

func FailHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(500)
	w.Write([]byte("Server Error"))
}

func SleepHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	secs, _ := strconv.Atoi(r.Form.Get("seconds"))
	if secs < 1 {
		secs = 5
	}
	time.Sleep(time.Duration(secs) * time.Second)
	HomeHandler(w, r)
}
