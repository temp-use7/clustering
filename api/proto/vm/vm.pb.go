package vmpb

import (
	"context"

	"google.golang.org/grpc"
)

type Empty struct{}

type VM struct {
	Id     string
	Name   string
	NodeId string
	Cpu    int32
	Memory int32
	Disk   int32
	Phase  string
}

type ListVMsResponse struct{ Vms []*VM }
type UpsertVMRequest struct{ Vm *VM }
type DeleteVMRequest struct{ Id string }
type MigrateRequest struct {
	Id         string
	TargetNode string
}

type VMServiceServer interface {
	ListVMs(context.Context, *Empty) (*ListVMsResponse, error)
	UpsertVM(context.Context, *UpsertVMRequest) (*Empty, error)
	DeleteVM(context.Context, *DeleteVMRequest) (*Empty, error)
	Migrate(context.Context, *MigrateRequest) (*Empty, error)
}

type UnimplementedVMServiceServer struct{}

func RegisterVMServiceServer(s *grpc.Server, srv VMServiceServer) {}
