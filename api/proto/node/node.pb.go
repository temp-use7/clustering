package nodepb

import (
	"context"

	"google.golang.org/grpc"
)

type Empty struct{}

type Node struct {
	Id      string
	Address string
	Role    string
	Status  string
}

type ListNodesResponse struct{ Nodes []*Node }

type NodeServiceServer interface {
	ListNodes(context.Context, *Empty) (*ListNodesResponse, error)
}

type UnimplementedNodeServiceServer struct{}

func RegisterNodeServiceServer(s *grpc.Server, srv NodeServiceServer) {}
