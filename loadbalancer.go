package main
import "net/url"
import "fmt"
import "net/http" 
import "io"      

type Server struct {
	URL string
}

type LoadBalancer struct {
	servers []*Server
	idx     int
}

func (lb *LoadBalancer) NextServer() *Server {
	server := lb.servers[lb.idx]
	lb.idx = (lb.idx + 1) % len(lb.servers)

	return server
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
	nextServer := lb.NextServer()
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