package main

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
)

func get_state_file() string {
	statefilepath := os.Getenv("STATEFILEPATH")
	if len(statefilepath) > 0 {
		return statefilepath
	} else {
		return filepath.Join(filepath.Dir(os.Args[0]), "server_state.data")
	}
}

func read_state() string {
	file, err := os.Open(get_state_file())
	if err != nil {
		return err.Error()
	}
	defer file.Close()

	buf := make([]byte, 100)
	n, err := file.Read(buf)
	return string(buf[:n])
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

func handler(w http.ResponseWriter, r *http.Request) {
	fmt.Printf("handle %s\n", r.RequestURI)

	if r.RequestURI == "/" {
		fmt.Fprintln(w, read_state())
	} else {
		fmt.Fprintln(w, write_state(r.RequestURI))
	}
}

func main() {
	fmt.Println("state launch.")
	addr := get_listen_addr()

	fmt.Printf("State server started (listen %s).\n", addr)
	fmt.Printf("State file = %s.\n", get_state_file())
	http.HandleFunc("/", handler)
	http.HandleFunc("/info", handler_info)
	err := http.ListenAndServe(addr, nil)
	if err != nil {
		panic(err.Error())
	}
}
