package main

import (
	"flag"
	"log"
	"net/http"
	"strconv"
	"strings"

	"clustering/pkg/membership"
)

func main() {
	var (
		httpAddr  string
		nodeID    string
		serfBind  string
		join      string
		joinToken string
		cpu       int
		memory    int
		disk      int
	)
	flag.StringVar(&httpAddr, "http", ":9090", "node agent http addr")
	flag.StringVar(&nodeID, "node-id", "node-1", "node id")
	flag.StringVar(&serfBind, "serf-bind", ":7947", "serf bind addr")
	flag.StringVar(&join, "join", "", "comma separated serf peers to join")
	flag.StringVar(&joinToken, "join-token", "", "shared join token for control-plane validation")
	flag.IntVar(&cpu, "cpu", 8000, "capacity CPU (millicores)")
	flag.IntVar(&memory, "memory", 32768, "capacity memory (MiB)")
	flag.IntVar(&disk, "disk", 512, "capacity disk (GiB)")
	flag.Parse()

	s, _ := membership.MustStartSerf(membership.Config{NodeID: nodeID, BindAddr: serfBind})
	// advertise HTTP port via tag
	httpPort := httpAddr
	if strings.HasPrefix(httpPort, ":") {
		httpPort = strings.TrimPrefix(httpPort, ":")
	} else {
		if i := strings.LastIndex(httpPort, ":"); i >= 0 && i+1 < len(httpPort) {
			httpPort = httpPort[i+1:]
		}
	}
	tags := map[string]string{"role": "node", "http": httpPort, "cpu": strconv.Itoa(cpu), "memory": strconv.Itoa(memory), "disk": strconv.Itoa(disk)}
	if joinToken != "" {
		tags["token"] = joinToken
	}
	if err := s.SetTags(tags); err != nil {
		log.Printf("serf set tags: %v", err)
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(200) })
	log.Printf("nodeagent listening on %s", httpAddr)
	if err := http.ListenAndServe(httpAddr, mux); err != nil {
		log.Fatal(err)
	}
}
