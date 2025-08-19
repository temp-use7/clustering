package membership

import (
	"log"
	"time"

	"github.com/hashicorp/raft"
)

type AliveMember struct {
	ID       string
	RaftAddr string
}

type ListAliveMembersFunc func() []AliveMember

type Controller struct {
	raftNode      *raft.Raft
	listAlive     ListAliveMembersFunc
	interval      time.Duration
	desiredVoters int
	desiredFunc   func() int
}

func NewController(r *raft.Raft, listAlive ListAliveMembersFunc) *Controller {
	return &Controller{raftNode: r, listAlive: listAlive, interval: 10 * time.Second, desiredVoters: 5}
}

// WithDesiredVotersFunc allows dynamic desired voters from state/config.
func (c *Controller) WithDesiredVotersFunc(fn func() int) *Controller {
	c.desiredFunc = fn
	return c
}

func (c *Controller) Run(stop <-chan struct{}) {
	ticker := time.NewTicker(c.interval)
	defer ticker.Stop()
	for {
		select {
		case <-stop:
			return
		case <-ticker.C:
			if c.raftNode.State() == raft.Leader {
				c.reconcileOnce()
			}
		}
	}
}

func (c *Controller) reconcileOnce() {
	if c.desiredFunc != nil {
		if dv := c.desiredFunc(); dv > 0 {
			c.desiredVoters = dv
		}
	}
	cfgFut := c.raftNode.GetConfiguration()
	if err := cfgFut.Error(); err != nil {
		log.Printf("membership: get config error: %v", err)
		return
	}
	cfg := cfgFut.Configuration()
	aliveList := c.listAlive()
	alive := map[string]string{}
	for _, a := range aliveList {
		alive[a.ID] = a.RaftAddr
	}

	var existing []ExistingServer
	for _, s := range cfg.Servers {
		es := ExistingServer{ID: string(s.ID), Address: string(s.Address)}
		if s.Suffrage == raft.Voter {
			es.Suffrage = "voter"
		} else {
			es.Suffrage = "nonvoter"
		}
		existing = append(existing, es)
	}

	addNonvoters, promote, demote := Plan(existing, alive, c.desiredVoters)

	for _, id := range addNonvoters {
		addr := alive[id]
		if err := c.raftNode.AddNonvoter(raft.ServerID(id), raft.ServerAddress(addr), 0, 0).Error(); err != nil {
			log.Printf("add nonvoter %s error: %v", id, err)
		} else {
			log.Printf("added nonvoter %s (%s)", id, addr)
		}
	}

	if len(promote) > 0 || len(demote) > 0 {
		log.Printf("membership plan: promote=%v demote=%v", promote, demote)
	}

	for _, idStr := range promote {
		id := raft.ServerID(idStr)
		addr := alive[idStr]
		if err := c.raftNode.AddVoter(id, raft.ServerAddress(addr), 0, 0).Error(); err != nil {
			log.Printf("promote error %s: %v", id, err)
		}
	}
	for _, idStr := range demote {
		id := raft.ServerID(idStr)
		if err := c.raftNode.RemoveServer(id, 0, 0).Error(); err != nil {
			log.Printf("remove voter error %s: %v", id, err)
			continue
		}
		addr := alive[idStr]
		if err := c.raftNode.AddNonvoter(id, raft.ServerAddress(addr), 0, 0).Error(); err != nil {
			log.Printf("add nonvoter error %s: %v", id, err)
		}
	}
}
