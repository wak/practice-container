package main

import (
	"fmt"
	"net/http"
	"os"
)

var ServerName = "C2"
var ResponseCount = 0

func handler(w http.ResponseWriter, r *http.Request) {
	ResponseCount += 1
	hostname, _ := os.Hostname()

	fmt.Fprintf(w, "I am server %s on host '%s' (%d times).\n",
		ServerName, hostname, ResponseCount)
}

func main() {
	fmt.Println("Test server", ServerName, "started.")

	http.HandleFunc("/", handler)
	http.ListenAndServe(":8080", nil)
}
