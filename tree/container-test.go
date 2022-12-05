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

var nodeName = "P1"

var runAt = now()
var responseCount = 0

func now() time.Time {
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

func makeNodeStopUrl(nodeNumber int, method string) string {
	copiedURL := *nodeUrls[nodeNumber-1]
	copiedURL.Path = path.Join(copiedURL.Path, "stop", method, "self")
	return copiedURL.String()
}

func stopChild(w http.ResponseWriter, target string, method string) {
	nodeNumber, err := strconv.Atoi(target)
	if err != nil {
		fmt.Printf("Invalid target number: %s\n", target)
		return
	}
	if len(nodeUrls) < nodeNumber {
		fmt.Printf("node not found: %d\n", nodeNumber)
		return
	}
	fmt.Fprintf(w, "%s", get(makeNodeStopUrl(nodeNumber, method)))
}

func whoAmI(who string) string {
	hostname, _ := os.Hostname()

	state, err := readState()
	if err != nil {
		state = ""
	}
	return fmt.Sprintf("%s (host: %s, ver: %s, runat: %s, call: %d, PID: %d, file: %s).\n",
		who,
		hostname,
		AppVersion,
		runAt.Format("2006-01-02 15:04:05"),
		responseCount,
		os.Getpid(),
		state)
}

func dumpRequest(w io.Writer, r *http.Request) {
	fmt.Fprintln(w, "\nREQUEST:")
	fmt.Fprintf(w, "  Method    : %s\n", r.Method)
	fmt.Fprintf(w, "  RequestURI: %s\n", r.RequestURI)
	fmt.Fprintf(w, "  RemoteAddr: %s\n", r.RemoteAddr)
	fmt.Fprintf(w, "  Proto     : %s\n", r.Proto)
	fmt.Fprintf(w, "  Host      : %s\n", r.Host)

	fmt.Fprintln(w, "  HEADER:")
	for _, name := range sortedKeys(r.Header) {
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

func logRequest(handlerName string, r *http.Request) {
	fmt.Printf("[%-7s] %-13s %s %s - %s\n",
		handlerName,
		r.RemoteAddr,
		r.Method,
		r.RequestURI,
		r.Header.Get("User-Agent"),
	)
}

func handlerRoot(w http.ResponseWriter, r *http.Request) {
	logRequest("default", r)

	if r.RequestURI != "/" {
		w.WriteHeader(404)
		fmt.Fprintln(w, "unsupported path")
		return
	}
	responseCount += 1

	fmt.Fprintf(w, "%s %s %s %d\n",
		now().Format("2006-01-02 15:04:05"),
		nodeName,
		AppVersion,
		responseCount,
	)
}

func handlerStatus(w http.ResponseWriter, r *http.Request) {
	logRequest("status", r)
	responseCount += 1

	fmt.Fprintf(w, whoAmI(nodeName))
	for n, nodeUrl := range nodeUrls {
		fmt.Fprintf(w, "  NODE[%d] ... %s\n",
			n+1,
			strings.TrimRight(get(nodeUrl.String()), "\n"))
	}
}

func handlerHealth(w http.ResponseWriter, r *http.Request) {
	logRequest("health", r)
	responseCount += 1

	fmt.Fprint(w, "OK")
}

func handlerFile(w http.ResponseWriter, r *http.Request) {
	logRequest("file", r)
	responseCount += 1

	if r.RequestURI == "/file" || r.RequestURI == "/file/" {
		state, _ := readState()
		fmt.Fprintln(w, state)
	} else {
		fmt.Fprintln(w, writeState(r.RequestURI[6:]))
	}
}

func getStateFile() string {
	statefilepath := os.Getenv("STATEFILEPATH")
	if len(statefilepath) > 0 {
		return statefilepath
	} else {
		return filepath.Join(filepath.Dir(os.Args[0]), "server_state.data")
	}
}

func readState() (string, error) {
	file, err := os.Open(getStateFile())
	if err != nil {
		return err.Error(), err
	}
	defer file.Close()

	buf := make([]byte, 100)
	n, err := file.Read(buf)
	return string(buf[:n]), nil
}

func writeState(s string) string {
	file, err := os.Create(getStateFile())
	if err != nil {
		return err.Error()
	}
	defer file.Close()
	file.Write(([]byte)(s))
	return "state saved."
}

func handlerStop(w http.ResponseWriter, r *http.Request) {
	logRequest("stop", r)
	responseCount += 1

	parts := strings.Split(r.RequestURI, "/")
	if len(parts) != 4 {
		fmt.Printf("invalid parameter: %v\n", parts)
		return
	}

	stopMethod := parts[2]
	stopTarget := parts[3]
	fmt.Printf("method: %v, target %v\n", stopMethod, stopTarget)
	if stopTarget == "self" {
		switch stopMethod {
		case "panic":
			fmt.Println("panic request received.")
			time.AfterFunc(time.Second*5, func() {
				panic("see you!")
			})
			fmt.Fprintf(w, "Server %s will panic after 5 seconds.\n", nodeName)
		case "success":
			fmt.Println("stop success request received.")
			stopStatus = 0
			stopChan <- 1
			fmt.Fprintf(w, "Server %s will stop with success status after 5 seconds.\n", nodeName)
		case "error":
			fmt.Println("stop error request received.")
			stopStatus = 1
			stopChan <- 1
			fmt.Fprintf(w, "Server %s will stop with error status after 5 seconds.\n", nodeName)
		}
	} else {
		stopChild(w, stopTarget, stopMethod)
	}
}

func sortedKeys(m map[string][]string) []string {
	s := make([]string, len(m))
	index := 0
	for key := range m {
		s[index] = key
		index++
	}
	sort.Strings(s)
	return s
}

func handlerInfo(w http.ResponseWriter, r *http.Request) {
	logRequest("info", r)
	responseCount += 1

	pid := os.Getpid()
	hostname, _ := os.Hostname()

	fmt.Fprintln(w, "<html><head><title>INFO</title></head><body><pre>")
	fmt.Fprintf(w, "Response: %d\n", responseCount)
	fmt.Fprintf(w, "Time    : %s\n\n", now())

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
		dumpRequest(w, r)
	}
	fmt.Fprintln(w, "</pre></body></html>")
}

func makeRandomString() string {
	s := now().String()
	r := sha256.Sum256([]byte(s))
	b := r[:]
	return hex.EncodeToString(b)
}

func handlerVersion(w http.ResponseWriter, r *http.Request) {
	logRequest("version", r)
	responseCount += 1
	fmt.Fprintln(w, AppVersion)
}

func handlerApi(w http.ResponseWriter, r *http.Request) {
	logRequest("api", r)
	responseCount += 1

	fmt.Fprintln(w, "<html><head><title>API List</title></head><body>")
	fmt.Fprintln(w, "<h1>This Process API</h1>")
	fmt.Fprintln(w, "<ul>")
	entry := func(desc string, link string) {
		fmt.Fprintf(w, "<li>%s: <a href=\"%s\">%s</a></li>\n", desc, link, link)
	}
	entry("Server simple message", "/")
	entry("Server version", "/version")
	entry("Server status", "/status")
	entry("Server health", "/health")
	entry("Show information", "./info")
	fmt.Fprintln(w, "<li>Actions:</li><ul>")
	entry("Write file: random", "./file/"+makeRandomString())
	now := regexp.MustCompile("[+ /:]").ReplaceAllString(now().String(), "_")
	entry("Write file: time", "./file/"+now)
	entry("Read file", "./file")
	entry("Stop me (panic)", "./stop/panic/self")
	entry("Stop me (success)", "./stop/success/self")
	entry("Stop me (error)", "./stop/error/self")
	fmt.Fprintln(w, "</ul>")
	fmt.Fprintln(w, "</ul>")

	state, _ := readState()
	fmt.Fprintf(w, "Current state file: <ul><li>path: %s</li><li>value: %s</li></ul>\n", getStateFile(), state)

	for n, nodeUrl := range nodeUrls {
		fmt.Fprintf(w, "<h1>Node %d API</h1>\n", n+1)
		fmt.Fprintln(w, "<ul>")
		entry("Direct URL", nodeUrl.String())
		entry("Stop with panic", path.Join("./", "stop", "panic", strconv.Itoa(n+1)))
		entry("Stop with error", path.Join("./", "stop", "error", strconv.Itoa(n+1)))
		entry("Stop with success", path.Join("./", "stop", "success", strconv.Itoa(n+1)))
		fmt.Fprintln(w, "</ul>")
	}
	fmt.Fprintln(w, "</body></html>")
}

func getListenAddr() string {
	if len(os.Getenv("PORT")) > 0 {
		return ":" + os.Getenv("PORT")
	} else if len(os.Args) > 1 {
		return ":" + os.Args[1]
	} else {
		return ":8080"
	}
}

func getNodeName() string {
	if len(os.Getenv("nodeName")) > 0 {
		return os.Getenv("nodeName")
	} else {
		hostname, err := os.Hostname()
		if err != nil {
			return "unknown"
		} else {
			return hostname
		}
	}
}

var nodeUrls []*url.URL

func loadNodeConfig() {
	for n := 1; true; n += 1 {
		env := fmt.Sprintf("NODE%d", n)
		if len(os.Getenv(env)) == 0 {
			break
		}
		nodeUrl, _ := url.Parse(os.Getenv(env))
		nodeUrls = append(nodeUrls, nodeUrl)
	}
	fmt.Printf("%d nodes found.\n", len(nodeUrls))
}

var stopChan chan int
var stopStatus int

func run() {
	loadNodeConfig()

	addr := getListenAddr()
	http.HandleFunc("/", handlerRoot)
	http.HandleFunc("/health", handlerHealth)
	http.HandleFunc("/status", handlerStatus)
	http.HandleFunc("/version", handlerVersion)
	http.HandleFunc("/api", handlerApi)
	http.HandleFunc("/stop/", handlerStop)
	http.HandleFunc("/info", handlerInfo)
	http.HandleFunc("/file", handlerFile)
	http.HandleFunc("/file/", handlerFile)

	srv := http.Server{Addr: addr}
	go func() {
		stopChan = make(chan int)
		<-stopChan
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

	fmt.Printf("Parent node %s started (listen %s).\n", nodeName, addr)
	err := srv.ListenAndServe()
	if err != nil && err != http.ErrServerClosed {
		panic(err.Error())
	}
}

func main() {
	nodeName = getNodeName()

	run()

	fmt.Println("see you.")
	os.Exit(stopStatus)
}
