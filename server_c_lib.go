package main

import (
	"fmt"
	"time"
	"net/http"
	"os"
)

var RunAt = time.Now()
var AppRevision = "undefined"
var ServerName = "C2"
var ResponseCount = 0

func handler_default(w http.ResponseWriter, r *http.Request) {
	ResponseCount += 1
	hostname, _ := os.Hostname()

	fmt.Fprintf(w,
		"I am %s (host: %s, runat: %s, rev: %s, call: %d).\n",
		ServerName,
		hostname,
		RunAt.Format("2006-01-02 03:04:05"),
		AppRevision,
		ResponseCount)
}

func handler_crash(w http.ResponseWriter, r *http.Request) {
	fmt.Println("crash request received.")

	time.AfterFunc(time.Second * 5, func() {
		panic("crash!")
	})
	fmt.Fprintf(w, "Server %s will crash after 2 seconds.\n", ServerName)
}

func run_child(server_name string, revision string) {
	ServerName = server_name
	AppRevision = revision

	fmt.Println("Test server", ServerName, "started.")

	http.HandleFunc("/", handler_default)
	http.HandleFunc("/crash", handler_crash)
	http.ListenAndServe(":8080", nil)
}
