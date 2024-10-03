package main

import "net/http"

func main() {
	servers := []*Server{
		&Server{
			URL: "http://127.0.0.1:8000",
		},
		&Server{
			URL: "http://127.0.0.1:8001",
		},
	}

	lb := &LoadBalancer{
		servers: servers,
	}
	err := http.ListenAndServe(":8080", lb)
	if err != nil {
		panic(err)
	}
}