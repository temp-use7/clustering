package health

import (
	"log"
	"time"
)

type PingFunc func() error

type Controller struct {
	interval time.Duration
	ping     PingFunc
}

func NewController(p PingFunc) *Controller { return &Controller{interval: 15 * time.Second, ping: p} }

func (c *Controller) Run(stop <-chan struct{}) {
	t := time.NewTicker(c.interval)
	defer t.Stop()
	for {
		select {
		case <-stop:
			return
		case <-t.C:
			if err := c.ping(); err != nil {
				log.Printf("health ping error: %v", err)
			}
		}
	}
}
