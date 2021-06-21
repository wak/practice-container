package main

import (
	"fmt"
	"net/http"
	"os"
)

func main() {
	if len(os.Args) > 1 && os.Args[1] == "test" {
		write_info(os.Stdout, nil)
	} else {
		addr := get_listen_addr()
		fmt.Printf("Test server started (listen %s).\n", addr)
		http.HandleFunc("/all", handler_info)
		http.HandleFunc("/", handler_info_oneliner)
		http.ListenAndServe(addr, nil)
	}
}
