package health

import (
	"log"
	"time"

	"clustering/pkg/metrics"
)

type PingFunc func() error

type Controller struct {
	interval time.Duration
	ping     PingFunc
}

func NewController(p PingFunc) *Controller { return &Controller{interval: 15 * time.Second, ping: p} }

// WithInterval allows overriding the probe interval (useful for tests)
func (c *Controller) WithInterval(d time.Duration) *Controller {
	c.interval = d
	return c
}

func (c *Controller) Run(stop <-chan struct{}) {
	t := time.NewTicker(c.interval)
	defer t.Stop()
	for {
		select {
		case <-stop:
			return
		case <-t.C:
			if err := c.ping(); err != nil {
				metrics.IncCounter("health_ping_errors_total")
				log.Printf("health ping error: %v", err)
			} else {
				metrics.IncCounter("health_ping_ok_total")
			}
		}
	}
}
