package loadbalancer

import "net/url"
import "fmt"
import "net/http"
import "io"
import "errors"
import "time"
import "sync"

type Server struct {
	URL     string
	healthy bool
}

type Opts struct {
	maxConnections int
	timeout        time.Duration
}

type ConnectionPool struct {
	*Opts
	clients map[string][]*http.Client
	mu      sync.Mutex
}

type LoadBalancer struct {
	servers []*Server
	idx     int
	mu      sync.Mutex
	cp      *ConnectionPool
}

func (lb *LoadBalancer) NextServer() (*Server, error) {
	lb.mu.Lock()
	defer lb.mu.Unlock()
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
	client := lb.cp.Get(server.URL)
	defer lb.cp.Push(server.URL, client)
	u, err := url.Parse(server.URL)
	if err != nil {
		return nil, err
	}
	fullUrl := u.ResolveReference(&url.URL{Path: uri})
	res, err := client.Get(fullUrl.String())
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

func (cp *ConnectionPool) Get(server string) *http.Client {
	cp.mu.Lock()
	defer cp.mu.Unlock()
	if clients, ok := cp.clients[server]; ok && len(clients) > 0 {
		client := clients[len(clients)-1]
		clients = clients[:len(clients)-1]
		cp.clients[server] = clients
		return client
	}
	return &http.Client{
		Timeout: cp.timeout,
	}
}

func (cp *ConnectionPool) Push(server string, client *http.Client) error {
	cp.mu.Lock()
	defer cp.mu.Unlock()
	if len(cp.clients[server]) > cp.maxConnections {
		return fmt.Errorf("connection pool limit exceeded for server '%s'", server)
	}
	cp.clients[server] = append(cp.clients[server], client)
	return nil
}

func NewOpts() *Opts {
	return &Opts{
		maxConnections: 10,
		timeout:        5 * time.Second,
	}
}

func (opts *Opts) MaxConnections(maxConnections int) *Opts {
	opts.maxConnections = maxConnections
	return opts
}

func (opts *Opts) Timeout(timeout time.Duration) *Opts {
	opts.timeout = timeout
	return opts
}

func NewConnectionPool(opts *Opts) *ConnectionPool {
	return &ConnectionPool{
		Opts:    opts,
		clients: make(map[string][]*http.Client),
	}
}

func NewLoadBalancer(urls []string, opts *Opts) *LoadBalancer {
	servers := make([]*Server, len(urls))
	for i, url := range urls {
		server := &Server{
			URL:     url,
			healthy: true,
		}
		servers[i] = server
	}
	cp := NewConnectionPool(opts)
	return &LoadBalancer{
		servers: servers,
		idx:     0,
		cp:      cp,
	}
}
