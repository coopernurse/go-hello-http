package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

func main() {
	http.HandleFunc("/", HomeHandler)
	http.HandleFunc("/fail", FailHandler)
	http.HandleFunc("/sleep", SleepHandler)
	http.HandleFunc("/log", LogHandler)
	http.HandleFunc("/call", CallHandler)
	http.HandleFunc("/env", EnvHandler)
	if _, err := os.Stat("/static"); !os.IsNotExist(err) {
		log.Printf("Enabling /static http server")
		http.HandleFunc("/static/", func(w http.ResponseWriter, r *http.Request) {
			http.ServeFile(w, r, r.URL.Path)
		})
	}
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

func LogHandler(w http.ResponseWriter, r *http.Request) {
	msg := fmt.Sprintf("Current time: %v", time.Now())
	fmt.Println(msg)
	fmt.Fprintf(w, "Wrote this to stdout: %s", msg)
}

// makes HTTP request to another component via maelstrom
func CallHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	component := r.Form.Get("component")
	baseUrl := os.Getenv("MAELSTROM_PRIVATE_URL")
	url := fmt.Sprintf("%s/log", baseUrl)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		log.Printf("ERROR: CallHandler error creating req for %s - %v", url, err)
		w.WriteHeader(500)
		w.Write([]byte("Server Error - couldn't form request"))
		return
	}

	req.Header.Add("MAELSTROM-COMPONENT", component)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Printf("ERROR: CallHandler error calling %s - %v", url, err)
		w.WriteHeader(500)
		w.Write([]byte("Server Error - couldn't connect to component"))
		return
	}

	_, err = io.Copy(w, resp.Body)
	resp.Body.Close()
	if err != nil {
		log.Printf("ERROR: CallHandler error copying response from %s - %v", url, err)
	}
}

func EnvHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	prefix := strings.ToUpper(r.Form.Get("prefix"))
	found := false

	w.Header().Set("Content-Type", "text/plain")
	for _, kv := range os.Environ() {
		if prefix == "" || strings.HasPrefix(strings.ToUpper(kv), prefix) {
			fmt.Fprintf(w, "%s\n", kv)
			found = true
		}
	}
	if !found {
		fmt.Fprintf(w, "No env vars found matching prefix: %s\n", prefix)
	}
}
