package main

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"path"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"syscall"
	"time"
)

var NodeName = "P1"

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

func make_node_stop_url(node_number int, method string) string {
	copiedURL := *node_urls[node_number-1]
	copiedURL.Path = path.Join(copiedURL.Path, "stop", method, "self")
	return copiedURL.String()
}

func stop_child(w http.ResponseWriter, target string, method string) {
	node_number, err := strconv.Atoi(target)
	if err != nil {
		fmt.Printf("Invalid target number: %s\n", target)
		return
	}
	if len(node_urls) < node_number {
		fmt.Printf("node not found: %d\n", node_number)
		return
	}
	fmt.Fprintf(w, "%s", get(make_node_stop_url(node_number, method)))
}

func who_am_i(who string) string {
	hostname, _ := os.Hostname()

	state, err := read_state()
	if err != nil {
		state = ""
	}
	return fmt.Sprintf("%s (host: %s, ver: %s, runat: %s, call: %d, PID: %d, file: %s).\n",
		who,
		hostname,
		AppVersion,
		RunAt.Format("2006-01-02 15:04:05"),
		ResponseCount,
		os.Getpid(),
		state)
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

func log_request(handler_name string, r *http.Request) {
	fmt.Printf("[%-7s] %-13s %s %s - %s\n",
		handler_name,
		r.RemoteAddr,
		r.Method,
		r.RequestURI,
		r.Header.Get("User-Agent"),
	)
}

func handler_root(w http.ResponseWriter, r *http.Request) {
	log_request("default", r)

	if r.RequestURI != "/" {
		w.WriteHeader(404)
		fmt.Fprintln(w, "unsupported path")
		return
	}
	ResponseCount += 1

	fmt.Fprintf(w, "%s %s %s %d\n",
		Now().Format("2006-01-02 15:04:05"),
		NodeName,
		AppVersion,
		ResponseCount,
	)
}

func handler_status(w http.ResponseWriter, r *http.Request) {
	log_request("status", r)
	ResponseCount += 1

	fmt.Fprintf(w, who_am_i(NodeName))
	for n, node_url := range node_urls {
		fmt.Fprintf(w, "  NODE[%d] ... %s\n",
			n+1,
			strings.TrimRight(get(node_url.String()), "\n"))
	}
}

func handler_health(w http.ResponseWriter, r *http.Request) {
	log_request("health", r)
	ResponseCount += 1

	fmt.Fprint(w, "OK")
}

func handler_file(w http.ResponseWriter, r *http.Request) {
	log_request("file", r)
	ResponseCount += 1

	if r.RequestURI == "/file" || r.RequestURI == "/file/" {
		state, _ := read_state()
		fmt.Fprintln(w, state)
	} else {
		fmt.Fprintln(w, write_state(r.RequestURI[6:]))
	}
}

func get_state_file() string {
	statefilepath := os.Getenv("STATEFILEPATH")
	if len(statefilepath) > 0 {
		return statefilepath
	} else {
		return filepath.Join(filepath.Dir(os.Args[0]), "server_state.data")
	}
}

func read_state() (string, error) {
	file, err := os.Open(get_state_file())
	if err != nil {
		return err.Error(), err
	}
	defer file.Close()

	buf := make([]byte, 100)
	n, err := file.Read(buf)
	return string(buf[:n]), nil
}

func write_state(s string) string {
	file, err := os.Create(get_state_file())
	if err != nil {
		return err.Error()
	}
	defer file.Close()
	file.Write(([]byte)(s))
	return "state saved."
}

func handler_stop(w http.ResponseWriter, r *http.Request) {
	log_request("stop", r)
	ResponseCount += 1

	parts := strings.Split(r.RequestURI, "/")
	if len(parts) != 4 {
		fmt.Printf("invalid parameter: %v\n", parts)
		return
	}

	stop_method := parts[2]
	stop_target := parts[3]
	fmt.Printf("method: %v, target %v\n", stop_method, stop_target)
	if stop_target == "self" {
		switch stop_method {
		case "panic":
			fmt.Println("panic request received.")
			time.AfterFunc(time.Second*5, func() {
				panic("see you!")
			})
			fmt.Fprintf(w, "Server %s will panic after 5 seconds.\n", NodeName)
		case "success":
			fmt.Println("stop success request received.")
			stop_status = 0
			stop_chan <- 1
			fmt.Fprintf(w, "Server %s will stop with success status after 5 seconds.\n", NodeName)
		case "error":
			fmt.Println("stop error request received.")
			stop_status = 1
			stop_chan <- 1
			fmt.Fprintf(w, "Server %s will stop with error status after 5 seconds.\n", NodeName)
		}
	} else {
		stop_child(w, stop_target, stop_method)
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

func handler_info(w http.ResponseWriter, r *http.Request) {
	log_request("info", r)
	ResponseCount += 1

	pid := os.Getpid()
	hostname, _ := os.Hostname()

	fmt.Fprintln(w, "<html><head><title>INFO</title></head><body><pre>")
	fmt.Fprintf(w, "Response: %d\n", ResponseCount)
	fmt.Fprintf(w, "Time    : %s\n\n", Now())

	fmt.Fprintf(w, "Hostname: %s\n", hostname)
	fmt.Fprintf(w, "PID     : %d\n", pid)

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

func make_random_string() string {
	s := Now().String()
	r := sha256.Sum256([]byte(s))
	b := r[:]
	return hex.EncodeToString(b)
}

func handler_api(w http.ResponseWriter, r *http.Request) {
	log_request("api", r)
	ResponseCount += 1

	fmt.Fprintln(w, "<html><head><title>API List</title></head><body>")
	fmt.Fprintln(w, "<h1>This Process API</h1>")
	fmt.Fprintln(w, "<ul>")
	entry := func(desc string, link string) {
		fmt.Fprintf(w, "<li>%s: <a href=\"%s\">%s</a></li>\n", desc, link, link)
	}
	entry("Server simple message", "/")
	entry("Server status", "/status")
	entry("Server health", "/health")
	entry("Show information", "./info")
	entry("Write file: random", "./file/"+make_random_string())
	now := regexp.MustCompile("[+ /:]").ReplaceAllString(Now().String(), "_")
	entry("Write file: time", "./file/"+now)
	entry("Read file", "./file")
	fmt.Fprintln(w, "</ul>")

	state, _ := read_state()
	fmt.Fprintf(w, "Current state file: <ul><li>path: %s</li><li>value: %s</li></ul>\n", get_state_file(), state)

	for n, node_url := range node_urls {
		fmt.Fprintf(w, "<h1>Node %d API</h1>\n", n+1)
		fmt.Fprintln(w, "<ul>")
		entry("Direct URL", node_url.String())
		entry("Stop with panic", path.Join("./", "stop", "panic", strconv.Itoa(n+1)))
		entry("Stop with error", path.Join("./", "stop", "error", strconv.Itoa(n+1)))
		entry("Stop with success", path.Join("./", "stop", "success", strconv.Itoa(n+1)))
		fmt.Fprintln(w, "</ul>")
	}
	fmt.Fprintln(w, "</body></html>")
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

func get_node_name() string {
	if len(os.Getenv("NODENAME")) > 0 {
		return os.Getenv("NODENAME")
	} else {
		hostname, err := os.Hostname()
		if err != nil {
			return "unknown"
		} else {
			return hostname
		}
	}
}

var node_urls []*url.URL

func load_node_config() {
	for n := 1; true; n += 1 {
		env := fmt.Sprintf("NODE%d", n)
		if len(os.Getenv(env)) == 0 {
			break
		}
		node_url, _ := url.Parse(os.Getenv(env))
		node_urls = append(node_urls, node_url)
	}
	fmt.Printf("%d nodes found.\n", len(node_urls))
}

var stop_chan chan int
var stop_status int

func run() {
	load_node_config()

	addr := get_listen_addr()
	http.HandleFunc("/", handler_root)
	http.HandleFunc("/health", handler_health)
	http.HandleFunc("/status", handler_status)
	http.HandleFunc("/api", handler_api)
	http.HandleFunc("/stop/", handler_stop)
	http.HandleFunc("/info", handler_info)
	http.HandleFunc("/file", handler_file)
	http.HandleFunc("/file/", handler_file)

	srv := http.Server{Addr: addr}
	go func() {
		stop_chan = make(chan int)
		<-stop_chan
		time.AfterFunc(time.Second*5, func() {
			fmt.Printf("shutting down server.\n")
			srv.Shutdown(context.TODO())
		})
	}()
	go func() {
		sig := make(chan os.Signal, 1)
		signal.Notify(sig, syscall.SIGTERM, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM)
		s := <-sig
		fmt.Printf("signal %s received.\n", s.String())
		srv.Shutdown(context.TODO())
	}()

	fmt.Printf("Parent node %s started (listen %s).\n", NodeName, addr)
	err := srv.ListenAndServe()
	if err != nil && err != http.ErrServerClosed {
		panic(err.Error())
	}
}

func main() {
	NodeName = get_node_name()

	run()

	fmt.Println("see you.")
	os.Exit(stop_status)
}
