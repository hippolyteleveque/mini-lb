package main

import "net/http"

func main() {
	servers := []*Server{
		&Server{
			URL:     "http://127.0.0.1:8000",
			healthy: true,
		},
		&Server{
			URL:     "http://127.0.0.1:8001",
			healthy: true,
		},
	}

	lb := &LoadBalancer{
		servers: servers,
	}

	lb.RunHealthCheck()

	err := http.ListenAndServe(":8080", lb)
	if err != nil {
		panic(err)
	}
}
