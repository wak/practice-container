package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

var nodeName = "P1"
var responseCount = 0

func now() time.Time {
	jst := time.FixedZone("Asia/Tokyo", 9*60*60)
	return time.Now().In(jst)
}

func logRequest(r *http.Request) {
	fmt.Printf("%-13s %s %s - %s\n",
		r.RemoteAddr,
		r.Method,
		r.RequestURI,
		r.Header.Get("User-Agent"),
	)
}

func handlerRoot(w http.ResponseWriter, r *http.Request) {
	logRequest(r)

	if r.RequestURI != "/" {
		w.WriteHeader(404)
		fmt.Fprintln(w, "unsupported path")
		return
	}
	responseCount += 1

	fmt.Fprintf(w, "[%d] %s %s\n",
		responseCount,
		now().Format("2006-01-02 15:04:05"),
		nodeName,
	)
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

func getHostname() string {
	hostname, err := os.Hostname()
	if err != nil {
		return "unknown"
	} else {
		return hostname
	}
}

func run() {
	addr := getListenAddr()
	http.HandleFunc("/", handlerRoot)

	srv := http.Server{Addr: addr}
	go func() {
		sig := make(chan os.Signal, 1)
		signal.Notify(sig, syscall.SIGTERM, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM)
		s := <-sig
		fmt.Printf("signal %s received.\n", s.String())
		srv.Shutdown(context.TODO())
	}()

	fmt.Printf("Server started (listen %s).\n", addr)
	err := srv.ListenAndServe()
	if err != nil && err != http.ErrServerClosed {
		panic(err.Error())
	}
}

func main() {
	nodeName = getHostname()
	run()
	fmt.Println("see you.")
}
