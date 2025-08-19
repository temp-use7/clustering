package grpcapi

import (
	clusterpb "clustering/api/proto/node"
	"clustering/pkg/api"
	"context"
)

type fsmReader interface{ GetStateCopy() api.ClusterState }

type NodeServer struct {
	clusterpb.UnimplementedNodeServiceServer
	fsm fsmReader
}

func NewNodeServer(fsm fsmReader) *NodeServer { return &NodeServer{fsm: fsm} }

func (s *NodeServer) ListNodes(ctx context.Context, _ *clusterpb.Empty) (*clusterpb.ListNodesResponse, error) {
	st := s.fsm.GetStateCopy()
	resp := &clusterpb.ListNodesResponse{}
	for _, n := range st.Nodes {
		resp.Nodes = append(resp.Nodes, &clusterpb.Node{Id: n.ID, Address: n.Address, Role: n.Role, Status: n.Status})
	}
	return resp, nil
}
