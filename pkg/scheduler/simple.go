package scheduler

import (
    "sort"
    "clustering/pkg/api"
)

// ChooseNode picks a node for a VM using a simple spread strategy based on allocated CPU.
func ChooseNode(state api.ClusterState, vm api.VM) (string, bool) {
    type cand struct{ id string; score int }
    var cands []cand
    for id, n := range state.Nodes {
        if n.Status != "Alive" { continue }
        // naive capacity check
        if n.Allocated.CPU+vm.Resources.CPU > n.Capacity.CPU { continue }
        if n.Allocated.Memory+vm.Resources.Memory > n.Capacity.Memory { continue }
        cands = append(cands, cand{id: id, score: n.Allocated.CPU})
    }
    if len(cands) == 0 { return "", false }
    sort.Slice(cands, func(i, j int) bool { return cands[i].score < cands[j].score })
    return cands[0].id, true
}



