package api

type Node struct {
	ID        string            `json:"id"`
	Address   string            `json:"address"`
	Role      string            `json:"role"` // control-plane or node
	Voter     bool              `json:"voter"`
	Capacity  Resources         `json:"capacity"`
	Allocated Resources         `json:"allocated"`
	Labels    map[string]string `json:"labels"`
	Taints    map[string]string `json:"taints"`
	Status    string            `json:"status"` // Alive/Failed/Left
}

type Resources struct {
	CPU    int `json:"cpu"`    // millicores
	Memory int `json:"memory"` // MiB
	Disk   int `json:"disk"`   // GiB
}

type VM struct {
	ID        string             `json:"id"`
	Name      string             `json:"name"`
	Resources Resources          `json:"resources"`
	NodeID    string             `json:"nodeId"`
	Phase     string             `json:"phase"` // Pending, Running, Migrating, Stopped
	Labels    map[string]string  `json:"labels"`
	Policy    VMSchedulingPolicy `json:"policy"`
}

type VMSchedulingPolicy struct {
	Priority int               `json:"priority"`
	Spread   bool              `json:"spread"`
	Affinity map[string]string `json:"affinity"`
}

type ClusterState struct {
	Nodes         map[string]Node        `json:"nodes"`
	VMs           map[string]VM          `json:"vms"`
	Networks      map[string]Network     `json:"networks"`
	StoragePools  map[string]StoragePool `json:"storagePools"`
	Config        ClusterConfig          `json:"config"`
	ConfigVersion int                    `json:"configVersion"`
	ConfigHistory []ClusterConfig        `json:"configHistory"`
}

// Future extensions
type Volume struct {
	ID   string `json:"id"`
	Size int    `json:"size"` // GiB
	Node string `json:"node"`
}

type Network struct {
	ID   string `json:"id"`
	CIDR string `json:"cidr"`
}

// ClusterConfig holds operator-tunable parameters.
type ClusterConfig struct {
	DesiredVoters    int `json:"desiredVoters"`
	DesiredNonVoters int `json:"desiredNonVoters"`
}

// StoragePool models a storage pool resource (placeholder for storage mgmt).
type StoragePool struct {
	ID   string `json:"id"`
	Type string `json:"type"` // e.g., local, nfs, iscsi (planned)
	Size int    `json:"size"` // GiB
}
