package main

import (
	"fmt"
	"net/http"
	"os"
)

func main() {
	if len(os.Args) > 1 && os.Args[1] == "test" {
		write_info(os.Stdout)
	} else {
		addr := ":8080"
		if len(os.Args) > 1 {
			addr = ":" + os.Args[1]
		}

		fmt.Printf("Test server started (listen %s).\n", addr)
		http.HandleFunc("/", handler_info)
		http.ListenAndServe(addr, nil)
	}
}
