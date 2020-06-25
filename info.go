package main

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"
)

var ResponseCount = 0

func write(w io.Writer) {
	pid := os.Getpid()
	hostname, _ := os.Hostname()

	ResponseCount += 1

	fmt.Fprintln(w, "<html><head><title>Test server</title></head><body><pre>")
	fmt.Fprintf(w, "Response: %d\n", ResponseCount)
	fmt.Fprintf(w, "Time: %s\n", time.Now())
	fmt.Fprintf(w, "Hostname: %s\n", hostname)
	fmt.Fprintf(w, "PID: %d\n", pid)

	fmt.Fprintf(w, "ARGV\n")
	for i, v := range os.Args {
		fmt.Fprintf(w, "  [%d] %s\n", i, v)
	}

	fmt.Fprintf(w, "ENV\n")
	for _, e := range os.Environ() {
		pair := strings.SplitN(e, "=", 2)
		fmt.Fprintf(w, "  %s = %s\n", pair[0], pair[1])
	}
	fmt.Fprintln(w, "</pre></body></html>")
}

func handler(w http.ResponseWriter, r *http.Request) {
	write(w)
}

func main() {
	if len(os.Args) > 1 && os.Args[1] == "test" {
		write(os.Stdout)
	} else {
		addr := ":8080"
		if len(os.Args) > 1 {
			addr = ":" + os.Args[1]
		}

		fmt.Printf("Test server started (listen %s).\n", addr)
		http.HandleFunc("/", handler)
		http.ListenAndServe(addr, nil)
	}
}
