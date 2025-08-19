package schedulerctl

import (
	"context"
	"log"
	"time"

	"clustering/pkg/api"
	"clustering/pkg/scheduler"
	"clustering/pkg/store"
)

type StateReader interface{ GetStateCopy() api.ClusterState }

type Controller struct {
	state    StateReader
	st       *store.Manager
	interval time.Duration
}

func NewController(sr StateReader, st *store.Manager) *Controller {
	return &Controller{state: sr, st: st, interval: 5 * time.Second}
}

func (c *Controller) Run(stop <-chan struct{}) {
	t := time.NewTicker(c.interval)
	defer t.Stop()
	for {
		select {
		case <-stop:
			return
		case <-t.C:
			c.tick()
		}
	}
}

func (c *Controller) tick() {
	st := c.state.GetStateCopy()
	for _, vm := range st.VMs {
		if vm.NodeID == "" || vm.Phase == "Pending" {
			if nid, ok := scheduler.ChooseNode(st, vm); ok {
				vm.NodeID = nid
				vm.Phase = "Running"
				if err := c.st.Apply(context.Background(), store.NewCommand("UpsertVM", vm)); err != nil {
					log.Printf("schedule vm %s: %v", vm.ID, err)
				}
			}
		}
	}
}
