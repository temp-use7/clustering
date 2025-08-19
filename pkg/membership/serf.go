package membership

import (
	"log"
	"net"

	"github.com/hashicorp/serf/serf"
)

type Config struct {
	NodeID   string
	BindAddr string // host:port
}

type Event struct {
	Type string
	Node string
}

// MustStartSerf starts a Serf agent for gossip/membership and returns the instance and an events channel.
func MustStartSerf(cfg Config) (*serf.Serf, <-chan Event) {
	sc := serf.DefaultConfig()
	sc.Init()
	sc.NodeName = cfg.NodeID
	host, port := hostPort(cfg.BindAddr)
	// Normalize advertise/bind to loopback if unspecified or wildcard
	sc.MemberlistConfig.BindAddr = host
	sc.MemberlistConfig.BindPort = port
	sc.MemberlistConfig.AdvertiseAddr = host
	sc.MemberlistConfig.AdvertisePort = port

	ch := make(chan serf.Event, 64)
	sc.EventCh = ch

	s, err := serf.Create(sc)
	if err != nil {
		log.Fatalf("serf create: %v", err)
	}

	out := make(chan Event, 64)
	go func() {
		for ev := range ch {
			switch e := ev.(type) {
			case serf.MemberEvent:
				for _, m := range e.Members {
					out <- Event{Type: e.EventType().String(), Node: m.Name}
				}
			case serf.UserEvent:
				out <- Event{Type: "user:" + e.Name, Node: string(e.Payload)}
			}
		}
	}()
	return s, out
}

func hostPort(addr string) (string, int) {
	h, ps, err := net.SplitHostPort(addr)
	if err != nil {
		log.Fatalf("bad addr %s: %v", addr, err)
	}
	port, err := net.LookupPort("tcp", ps)
	if err != nil {
		log.Fatalf("lookup port: %v", err)
	}
	if h == "" || h == "0.0.0.0" || h == "::" {
		h = "127.0.0.1"
	}
	return h, port
}
