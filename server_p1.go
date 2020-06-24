package main

import (
	"os"
	"fmt"
	"strings"
	"time"
	"io/ioutil"
	"net/http"
)

var ServerName = "C1"
var ResponseCount = 0

func get(url string) string {
	client := &http.Client {
		Timeout: 5 * time.Second,
	}
	resp, err := client.Get(url)
	if err != nil {
		return err.Error()
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err.Error()
	}
	return string(body)
}

func handler(w http.ResponseWriter, r *http.Request) {
	hostname, _ := os.Hostname()
	ResponseCount += 1
	
	fmt.Fprintf(w, "I am parent server on host %s (%d times).\n", hostname, ResponseCount)
	fmt.Fprintf(w, "  C1 ... %s\n", strings.TrimRight(get(os.Getenv("C1")), "\n"))
	fmt.Fprintf(w, "  C2 ... %s\n", strings.TrimRight(get(os.Getenv("C2")), "\n"))
}

func main() {
	fmt.Println("Test parent server started.")

	http.HandleFunc("/", handler)
	http.ListenAndServe(":8080", nil)
}
