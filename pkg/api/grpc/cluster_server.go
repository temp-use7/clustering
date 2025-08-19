package grpcapi

import (
	clusterpb "clustering/api/proto/cluster"
	"context"

	"github.com/hashicorp/raft"
)

type ClusterServer struct {
	clusterpb.UnimplementedClusterServiceServer
	raft *raft.Raft
}

func NewClusterServer(r *raft.Raft) *ClusterServer { return &ClusterServer{raft: r} }

func (s *ClusterServer) GetStatus(ctx context.Context, _ *clusterpb.Empty) (*clusterpb.ClusterStatus, error) {
	cfgF := s.raft.GetConfiguration()
	if err := cfgF.Error(); err != nil {
		return nil, err
	}
	cfg := cfgF.Configuration()
	var voters, learners int32
	for _, sv := range cfg.Servers {
		if sv.Suffrage == raft.Voter {
			voters++
		} else {
			learners++
		}
	}
	return &clusterpb.ClusterStatus{Leader: string(s.raft.Leader()), VoterCount: voters, NonvoterCount: learners}, nil
}

func (s *ClusterServer) ReconfigureMembership(ctx context.Context, req *clusterpb.ReconfigureRequest) (*clusterpb.ReconfigureResponse, error) {
	// Placeholder: actual logic handled by membership controller; accept for now
	return &clusterpb.ReconfigureResponse{Accepted: true}, nil
}
