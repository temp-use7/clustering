package membership

import "sort"

// ExistingServer models a current raft server entry for planning.
type ExistingServer struct {
	ID       string
	Suffrage string // "voter" or "nonvoter"
	Address  string
}

// Plan computes membership actions given existing servers, alive members, and desired voter count.
// alive maps node ID -> raft address.
func Plan(existing []ExistingServer, alive map[string]string, desiredVoters int) (addNonvoters []string, promote []string, demote []string) {
	if desiredVoters < 1 {
		desiredVoters = 1
	}
	// set of existing IDs
	existingSet := map[string]ExistingServer{}
	var voters []string
	var nonvoters []string
	for _, s := range existing {
		existingSet[s.ID] = s
		if s.Suffrage == "voter" {
			voters = append(voters, s.ID)
		} else {
			nonvoters = append(nonvoters, s.ID)
		}
	}
	// Any alive that's not in existing should be added as nonvoter
	for id := range alive {
		if _, ok := existingSet[id]; !ok {
			addNonvoters = append(addNonvoters, id)
		}
	}
	sort.Strings(voters)
	sort.Strings(nonvoters)
	// Promotions or demotions to reach desired voters
	if len(voters) < desiredVoters {
		need := desiredVoters - len(voters)
		// promote first nonvoters deterministically
		for i := 0; i < need && i < len(nonvoters); i++ {
			promote = append(promote, nonvoters[i])
		}
	} else if len(voters) > desiredVoters {
		surplus := voters[desiredVoters:]
		demote = append(demote, surplus...)
	}
	return
}
