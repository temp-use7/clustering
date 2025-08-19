package failover

import (
	"clustering/pkg/store"
	"log"
	"time"
)

type StateReader interface {
	GetStateCopy() interface{}
}

type Controller struct {
	st       *store.Manager
	interval time.Duration
	isLeader func() bool
}

func NewController(st *store.Manager, isLeader func() bool) *Controller {
	return &Controller{st: st, interval: 10 * time.Second, isLeader: isLeader}
}

func (c *Controller) Run(stop <-chan struct{}) {
	t := time.NewTicker(c.interval)
	defer t.Stop()
	for {
		select {
		case <-stop:
			return
		case <-t.C:
			if c.isLeader != nil && !c.isLeader() {
				continue
			}
			// TODO: inspect state for NotReady nodes and reassign VMs
			log.Printf("failover: tick")
		}
	}
}
