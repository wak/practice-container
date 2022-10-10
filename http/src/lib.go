package main

import (
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"os/user"
	"path"
	"sort"
	"strings"
	"time"
)

// for each server
var AppRevision = "r1"
var ServerName = "P1"

// for parent
var ChildUrl1 *url.URL
var ChildUrl2 *url.URL

// common
var RunAt = Now()
var ResponseCount = 0

func Now() time.Time {
	jst := time.FixedZone("Asia/Tokyo", 9*60*60)
	return time.Now().In(jst)
}

func get(url string) string {
	client := &http.Client{
		Timeout: 5 * time.Second,
	}
	resp, err := client.Get(url)
	if err != nil {
		return "FAILED: " + err.Error()
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "FAILED: " + err.Error()
	}
	return string(body)
}

func child_crash_url(url *url.URL) string {
	copiedURL := *url
	copiedURL.Path = path.Join(copiedURL.Path, "./crash")
	return copiedURL.String()
}

func who_am_i(who string) string {
	hostname, _ := os.Hostname()

	return fmt.Sprintf("I am %s (host: %s, runat: %s, rev: %s, call: %d, PID: %d).\n",
		who,
		hostname,
		RunAt.Format("2006-01-02 03:04:05"),
		AppRevision,
		ResponseCount,
		os.Getpid())
}

func write_parent_default(w io.Writer) {
	ResponseCount += 1

	fmt.Fprintf(w, who_am_i(ServerName))
	fmt.Fprintf(w, "  C1 ... %s\n", strings.TrimRight(get(ChildUrl1.String()), "\n"))
	fmt.Fprintf(w, "  C2 ... %s\n", strings.TrimRight(get(ChildUrl2.String()), "\n"))
}

func handler_parent_default(w http.ResponseWriter, r *http.Request) {
	fmt.Printf("handle %s\n", r.RequestURI)

	if r.RequestURI != "/" {
		fmt.Fprintln(w, "unsupported path")
		return
	}

	write_parent_default(w)
}

func handler_crash(w http.ResponseWriter, r *http.Request) {
	fmt.Printf("handle %s\n", r.RequestURI)
	fmt.Println("crash request received.")

	time.AfterFunc(time.Second*5, func() {
		panic("crash!")
	})
	fmt.Fprintf(w, "Server %s will crash after 2 seconds.\n", ServerName)
}

func handler_crash_c1(w http.ResponseWriter, r *http.Request) {
	fmt.Printf("handle %s\n", r.RequestURI)
	fmt.Println("crash request to c1 received.")
	fmt.Fprintf(w, get(child_crash_url(ChildUrl1)))
}

func handler_crash_c2(w http.ResponseWriter, r *http.Request) {
	fmt.Printf("handle %s\n", r.RequestURI)
	fmt.Println("crash request to c2 received.")
	fmt.Fprintf(w, get(child_crash_url(ChildUrl2)))
}

func handler_crash_c_all(w http.ResponseWriter, r *http.Request) {
	fmt.Printf("handle %s\n", r.RequestURI)
	fmt.Println("crash request to all children received.")
	fmt.Fprintf(w, get(child_crash_url(ChildUrl1)))
	fmt.Fprintf(w, get(child_crash_url(ChildUrl2)))
}

func write_info(w io.Writer, r *http.Request) {
	pid := os.Getpid()
	hostname, _ := os.Hostname()
	userCurrent, _ := user.Current()

	ResponseCount += 1

	fmt.Fprintln(w, "<html><head><title>INFO</title></head><body><pre>")
	fmt.Fprintf(w, "Response: %d\n", ResponseCount)
	fmt.Fprintf(w, "Time    : %s\n\n", Now())

	fmt.Fprintf(w, "Hostname: %s\n", hostname)
	fmt.Fprintf(w, "PID     : %d\n", pid)
	fmt.Fprintf(w, "UID     : %s\n", userCurrent.Uid)
	fmt.Fprintf(w, "GID     : %s\n", userCurrent.Gid)

	fmt.Fprintf(w, "\nARGV\n")
	for i, v := range os.Args {
		fmt.Fprintf(w, "  [%d] %s\n", i, v)
	}

	fmt.Fprintf(w, "\nENV\n")
	env := os.Environ()
	sort.Strings(env)
	for _, e := range env {
		pair := strings.SplitN(e, "=", 2)
		fmt.Fprintf(w, "  %-25s = %s\n", pair[0], pair[1])
	}

	if r != nil {
		write_request(w, r)
	}
	fmt.Fprintln(w, "</pre></body></html>")
}

func write_info_oneliner(w io.Writer, r *http.Request) {
	pid := os.Getpid()
	hostname, _ := os.Hostname()
	userCurrent, _ := user.Current()

	fmt.Fprintf(w, "[%s] host:%s pid:%d, uid:%s, gid:%s",
		Now().Format("2006/01/02 15:04:05"),
		hostname,
		pid, userCurrent.Uid, userCurrent.Gid)
}

func write_request(w io.Writer, r *http.Request) {
	fmt.Fprintln(w, "\nREQUEST:")
	fmt.Fprintf(w, "  Method    : %s\n", r.Method)
	fmt.Fprintf(w, "  RequestURI: %s\n", r.RequestURI)
	fmt.Fprintf(w, "  RemoteAddr: %s\n", r.RemoteAddr)
	fmt.Fprintf(w, "  Proto     : %s\n", r.Proto)
	fmt.Fprintf(w, "  Host      : %s\n", r.Host)

	fmt.Fprintln(w, "  HEADER:")
	for _, name := range sorted_keys(r.Header) {
		value := r.Header[name]

		if len(value) == 1 {
			fmt.Fprintf(w, "    %-25s = %s\n", name, value[0])
		} else {
			sort.Strings(value)
			fmt.Fprintf(w, "    %-25s =\n", name)
			for _, v := range value {
				fmt.Fprintf(w, "      %s\n", v)
			}
		}
	}
}

func sorted_keys(m map[string][]string) []string {
	s := make([]string, len(m))
	index := 0
	for key := range m {
		s[index] = key
		index++
	}
	sort.Strings(s)
	return s
}

func handler_info_oneliner(w http.ResponseWriter, r *http.Request) {
	fmt.Printf("handle %s\n", r.RequestURI)
	write_info_oneliner(w, r)
}

func handler_info(w http.ResponseWriter, r *http.Request) {
	fmt.Printf("handle %s\n", r.RequestURI)
	write_info(w, r)
}

func get_listen_addr() string {
	if len(os.Getenv("PORT")) > 0 {
		return ":" + os.Getenv("PORT")
	} else if len(os.Args) > 1 {
		return ":" + os.Args[1]
	} else {
		return ":8080"
	}
}

func run_parent(server_name string, revision string) {
	ServerName = server_name
	AppRevision = revision

	if len(os.Getenv("C1")) == 0 || len(os.Getenv("C2")) == 0 {
		panic("Environment variable C1 and C2 are required.")
	}

	ChildUrl1, _ = url.Parse(os.Getenv("C1"))
	ChildUrl2, _ = url.Parse(os.Getenv("C2"))

	if len(os.Args) > 1 && os.Args[1] == "test" {
		write_parent_default(os.Stdout)
		return
	}

	addr := get_listen_addr()
	fmt.Printf("Test parent server %s started (listen %s).\n", ServerName, addr)

	http.HandleFunc("/", handler_parent_default)
	http.HandleFunc("/info", handler_info)
	http.HandleFunc("/crash", handler_crash)

	http.HandleFunc("/child/crash/1", handler_crash_c1)
	http.HandleFunc("/child/crash/2", handler_crash_c2)
	http.HandleFunc("/child/crash/all", handler_crash_c_all)
	err := http.ListenAndServe(addr, nil)
	if err != nil {
		panic(err.Error())
	}
}

func handler_child_default(w http.ResponseWriter, r *http.Request) {
	fmt.Printf("handle %s\n", r.RequestURI)
	ResponseCount += 1
	fmt.Fprintf(w, who_am_i(ServerName))
}

func run_child(server_name string, revision string) {
	ServerName = server_name
	AppRevision = revision

	addr := get_listen_addr()
	fmt.Printf("Test server %s started (listen %s).\n", ServerName, addr)
	http.HandleFunc("/", handler_child_default)
	http.HandleFunc("/info", handler_info)
	http.HandleFunc("/crash", handler_crash)
	err := http.ListenAndServe(addr, nil)
	if err != nil {
		panic(err.Error())
	}
}
