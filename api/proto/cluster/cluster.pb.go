package clusterpb

import (
	"context"

	"google.golang.org/grpc"
)

type Empty struct{}

type ClusterStatus struct {
	Leader        string
	VoterCount    int32
	NonvoterCount int32
}

type ReconfigureRequest struct {
	DesiredVoters    int32
	DesiredNonvoters int32
}

type ReconfigureResponse struct{ Accepted bool }

type ClusterServiceServer interface {
	GetStatus(context.Context, *Empty) (*ClusterStatus, error)
	ReconfigureMembership(context.Context, *ReconfigureRequest) (*ReconfigureResponse, error)
}

type UnimplementedClusterServiceServer struct{}

func RegisterClusterServiceServer(s *grpc.Server, srv ClusterServiceServer) {}
