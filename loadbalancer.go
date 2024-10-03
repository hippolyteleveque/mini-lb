package main

import "net/url"
import "fmt"
import "net/http"
import "io"
import "errors"
import "time"

type Server struct {
	URL     string
	healthy bool
}

type LoadBalancer struct {
	servers []*Server
	idx     int
}

func (lb *LoadBalancer) NextServer() (*Server, error) {
	if lb.hasUnhealthy() {
		idx := 0
		for ; idx < len(lb.servers); idx++ {
			if lb.servers[idx].healthy {
				break
			}
		}
		lb.idx = idx
	}

	if lb.idx == len(lb.servers) {
		lb.idx = 0
		return nil, errors.New("no healthy server") // {{ edit_3 }}
	}
	server := lb.servers[lb.idx]
	lb.idx = (lb.idx + 1) % len(lb.servers)
	return server, nil
}

func (lb *LoadBalancer) hasUnhealthy() bool {
	for _, srv := range lb.servers {
		if !srv.healthy {
			return true
		}
	}
	return false
}

func (lb *LoadBalancer) ForwardRequest(server *Server, uri string) (*http.Response, error) {
	fmt.Printf("Forwarding request to : %s\n", server.URL) // {{ edit_3 }}
	// client := lb.cp.Pop(server.url)
	// defer lb.cp.Push(server.url, client)
	u, err := url.Parse(server.URL)
	if err != nil {
		return nil, err
	}
	fullUrl := u.ResolveReference(&url.URL{Path: uri})
	res, err := http.Get(fullUrl.String())
	if err != nil {
		return nil, err
	}
	return res, nil
}

func (lb *LoadBalancer) ServeHTTP(writer http.ResponseWriter, req *http.Request) { // {{ edit_4 }}
	nextServer, err := lb.NextServer()
	if err != nil {
		panic(err)
	}

	res, err := lb.ForwardRequest(nextServer, req.RequestURI)

	if err != nil {
		panic(err)
	}

	defer res.Body.Close()
	body, err := io.ReadAll(res.Body)
	_, err = writer.Write(body)
	if err != nil {
		panic(err)
	}
}

func (lb *LoadBalancer) HealthCheck() {
	for _, server := range lb.servers {
		res, err := http.Get(server.URL + "/healthcheck")
		if err != nil || res.StatusCode != http.StatusOK {
			server.healthy = false
			fmt.Printf("Server [%s] is down\n", server.URL)
		} else {
			server.healthy = true
			fmt.Printf("Server [%s] is up \n", server.URL)
		}
	}
}

func (lb *LoadBalancer) RunHealthCheck() {
	ticker := time.NewTicker(10 * time.Second)

	go func() {
		for {
			select {
			case <-ticker.C:
				lb.HealthCheck()
			}
		}
	}()
}
