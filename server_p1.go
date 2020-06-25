package main

import (
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"path"
	"strings"
	"time"
)

var AppRevision = "r1"
var RunAt = time.Now()

var ServerName = "P1"
var ResponseCount = 0

var ChildUrl1 *url.URL
var ChildUrl2 *url.URL

func get(url string) string {
	client := &http.Client{
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

func child_crash_url(url *url.URL) string {
	copiedURL := *url
    copiedURL.Path = path.Join(copiedURL.Path, "./crash")
	return copiedURL.String()
}

func write_default(w io.Writer) {
	hostname, _ := os.Hostname()
	ResponseCount += 1

	fmt.Fprintf(w,
		"I am parent server %s (host: %s, runat: %s, rev: %s, call: %d).\n",
		ServerName,
		hostname,
		RunAt.Format("2006-01-02 03:04:05"),
		AppRevision,
		ResponseCount)
	fmt.Fprintf(w, "  C1 ... %s\n", strings.TrimRight(get(ChildUrl1.String()), "\n"))
	fmt.Fprintf(w, "  C2 ... %s\n", strings.TrimRight(get(ChildUrl2.String()), "\n"))
}

func handler_default(w http.ResponseWriter, r *http.Request) {
	if r.RequestURI != "/" {
		fmt.Fprintln(w, "unsupported path")
		return
	}
	
	write_default(w)
}

func handler_crash(w http.ResponseWriter, r *http.Request) {
	fmt.Println("crash request received.")

	time.AfterFunc(time.Second * 5, func() {
		panic("crash!")
	})
	fmt.Fprintf(w, "Server %s will crash after 2 seconds.\n", ServerName)
}

func handler_crash_c1(w http.ResponseWriter, r *http.Request) {
	fmt.Println("crash request to c1 received.")
	fmt.Fprintf(w, get(child_crash_url(ChildUrl1)))
}

func handler_crash_c2(w http.ResponseWriter, r *http.Request) {
	fmt.Println("crash request to c2 received.")
	fmt.Fprintf(w, get(child_crash_url(ChildUrl2)))
}

func handler_crash_c_all(w http.ResponseWriter, r *http.Request) {
	fmt.Println("crash request to all children received.")
	fmt.Fprintf(w, get(child_crash_url(ChildUrl1)))
	fmt.Fprintf(w, get(child_crash_url(ChildUrl2)))
}

func main() {
	if (len(os.Getenv("C1")) == 0 || len(os.Getenv("C2")) == 0) {
		panic("Environment variable C1 and C2 are required.")
	}

	ChildUrl1, _ = url.Parse(os.Getenv("C1"))
	ChildUrl2, _ = url.Parse(os.Getenv("C2"))

	if (len(os.Args) > 1 && os.Args[1] == "test") {
		write_default(os.Stdout)
		return
	}
	
	addr := ":8080"
	if len(os.Args) > 1 {
		addr = ":" + os.Args[1]
	}
	
	fmt.Println("Test parent server", ServerName, "started.")

	http.HandleFunc("/", handler_default)
	http.HandleFunc("/crash", handler_crash)
	
	http.HandleFunc("/child/crash/1", handler_crash_c1)
	http.HandleFunc("/child/crash/2", handler_crash_c2)
	http.HandleFunc("/child/crash/all", handler_crash_c_all)
	http.ListenAndServe(addr, nil)
}
