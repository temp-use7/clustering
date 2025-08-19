package config

// ClusterConfig holds operator-tunable parameters.
type ClusterConfig struct {
	DesiredVoters    int `json:"desiredVoters"`
	DesiredNonVoters int `json:"desiredNonVoters"`
}

func Default() ClusterConfig {
	return ClusterConfig{DesiredVoters: 5, DesiredNonVoters: 2}
}
