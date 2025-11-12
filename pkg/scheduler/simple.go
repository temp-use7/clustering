package scheduler

import (
	"clustering/pkg/api"
	"sort"
)

// ChooseNode picks a node for a VM using a simple spread strategy based on allocated CPU,
// honoring a minimal label-based affinity if specified on the VM.
func ChooseNode(state api.ClusterState, vm api.VM) (string, bool) {
	type cand struct {
		id        string
		allocated int
		freeCPU   int
	}
	var cands []cand
	for id, n := range state.Nodes {
		if n.Status != "Alive" {
			continue
		}
		// naive capacity check
		if n.Allocated.CPU+vm.Resources.CPU > n.Capacity.CPU {
			continue
		}
		if n.Allocated.Memory+vm.Resources.Memory > n.Capacity.Memory {
			continue
		}
		// affinity: all labels in vm.Policy.Affinity must be present on node labels with same value
		if len(vm.Policy.Affinity) > 0 {
			matched := true
			for k, v := range vm.Policy.Affinity {
				if n.Labels == nil || n.Labels[k] != v {
					matched = false
					break
				}
			}
			if !matched {
				continue
			}
		}
		cands = append(cands, cand{id: id, allocated: n.Allocated.CPU, freeCPU: n.Capacity.CPU - n.Allocated.CPU})
	}
	if len(cands) == 0 {
		return "", false
	}
	// Default spread=true behavior; if explicitly disabled, pick most free CPU
	if vm.Policy.Spread || (!vm.Policy.Spread && vm.Policy.Priority == 0) {
		sort.Slice(cands, func(i, j int) bool { return cands[i].allocated < cands[j].allocated })
	} else {
		sort.Slice(cands, func(i, j int) bool { return cands[i].freeCPU > cands[j].freeCPU })
	}
	return cands[0].id, true
}
