package main

import (
    "net/http"
    "os"
)

func main() {
    port := "8000" // Default port
    if len(os.Args) > 1 {
        port = os.Args[1] // Use the first CLI argument as the port
    }

    http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
        if r.Method == http.MethodGet {
            w.WriteHeader(http.StatusOK) // Respond with 200 OK
        }
    })
	
	http.HandleFunc("/healthcheck", func(w http.ResponseWriter, r *http.Request) {
        if r.Method == http.MethodGet {
            w.WriteHeader(http.StatusOK) // Respond with 200 OK
        }
    })

    err := http.ListenAndServe(":"+port, nil) // Start the server on the specified port
	if err != nil {
		panic(err)
	}

}