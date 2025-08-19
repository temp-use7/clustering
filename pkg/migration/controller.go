package migration

import (
	"context"
	"log"
	"time"

	"clustering/pkg/api"
	"clustering/pkg/scheduler"
	"clustering/pkg/store"
)

type StateReader interface {
	GetStateCopy() api.ClusterState
}

type Controller struct {
	st       *store.Manager
	state    StateReader
	interval time.Duration
}

func NewController(st *store.Manager, sr StateReader) *Controller {
	return &Controller{st: st, state: sr, interval: 8 * time.Second}
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
	// find VMs on non-Alive nodes and re-place them
	for _, vm := range st.VMs {
		if vm.NodeID == "" {
			continue
		}
		n, ok := st.Nodes[vm.NodeID]
		if !ok || n.Status != "Alive" {
			// choose new node
			if nid, ok := scheduler.ChooseNode(st, vm); ok {
				vm.NodeID = nid
				vm.Phase = "Migrating"
				if err := c.st.Apply(context.Background(), store.NewCommand("UpsertVM", vm)); err != nil {
					log.Printf("migration propose error %s -> %s: %v", vm.ID, nid, err)
				}
			}
		}
	}
}
