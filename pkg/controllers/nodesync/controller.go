package nodesync

import (
	"context"
	"log"
	"strconv"
	"time"

	"clustering/pkg/api"
	"clustering/pkg/store"
)

type MemberInfo struct {
	ID     string
	Addr   string
	Role   string
	Status string // Alive/Failed/Left
	Tags   map[string]string
}

type ListMembersFunc func() []MemberInfo

type Controller struct {
	list     ListMembersFunc
	store    *store.Manager
	interval time.Duration
	isLeader func() bool
}

func NewController(list ListMembersFunc, st *store.Manager, isLeader func() bool) *Controller {
	return &Controller{list: list, store: st, interval: 5 * time.Second, isLeader: isLeader}
}

func (c *Controller) Run(stop <-chan struct{}) {
	t := time.NewTicker(c.interval)
	defer t.Stop()
	for {
		select {
		case <-stop:
			return
		case <-t.C:
			c.syncOnce()
		}
	}
}

func (c *Controller) syncOnce() {
	if c.isLeader != nil && !c.isLeader() {
		return
	}
	members := c.list()
	for _, m := range members {
		n := MemberToNode(m)
		if err := c.store.Apply(context.Background(), store.NewCommand("UpsertNode", n)); err != nil {
			log.Printf("nodesync upsert %s: %v", m.ID, err)
		}
	}
}

// MemberToNode converts MemberInfo into api.Node, reading capacity tags when present.
func MemberToNode(m MemberInfo) api.Node {
	n := api.Node{ID: m.ID, Address: m.Addr, Role: m.Role, Voter: false, Capacity: api.Resources{CPU: 8000, Memory: 32768, Disk: 512}, Status: m.Status}
	if m.Tags != nil {
		if v, ok := m.Tags["cpu"]; ok {
			if iv, err := strconv.Atoi(v); err == nil {
				n.Capacity.CPU = iv
			}
		}
		if v, ok := m.Tags["memory"]; ok {
			if iv, err := strconv.Atoi(v); err == nil {
				n.Capacity.Memory = iv
			}
		}
		if v, ok := m.Tags["disk"]; ok {
			if iv, err := strconv.Atoi(v); err == nil {
				n.Capacity.Disk = iv
			}
		}
	}
	return n
}
