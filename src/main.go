package main

import "net/http"
import "time"
import "mini-lb/loadbalancer"
import "mini-lb/config"
import "fmt"

func main() {
	cnf, err := config.Parse()
    if err != nil {
       panic(err)
    }	

	lb := loadbalancer.NewLoadBalancer(
		cnf.Servers,
		loadbalancer.NewOpts().
			Timeout(10*time.Second).
			MaxConnections(100),
	)

	lb.RunHealthCheck()

	err = http.ListenAndServe(fmt.Sprintf(":%d", cnf.Port), lb)
	if err != nil {
		panic(err)
	}
}
