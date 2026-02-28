package main

import (
	"fmt"
	"log"
	"net/http"
	"reverse-proxy/load_balancer"
)

const configPath = "app.yaml"

func main() {
	var config, err = LoadConfig(configPath)

	if err != nil {
		log.Fatal(err)
	}

	http.HandleFunc("/", indexHandler)
	port := config.Server.Port

	if config.LoadBalancer.Enabled {
		var backends = config.LoadBalancer.Resources
		var health = config.LoadBalancer.Health
		loadBalancer, err := load_balancer.NewLoadBalancer(backends, health.Path, health.Interval, health.Timeout)
		if err != nil {
			log.Fatal(err)
		}
		defer loadBalancer.Stop()
	}

	log.Printf("Listening on port %s", port)
	log.Printf("Open http://localhost:%s in the browser", port)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%s", port), nil))
}

func indexHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	_, err := fmt.Fprint(w, "Hello, World!")
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
	}
}
