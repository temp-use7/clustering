package grpcapi

import (
	vmpb "clustering/api/proto/vm"
	"clustering/pkg/api"
	"clustering/pkg/scheduler"
	"clustering/pkg/store"
	"context"
)

type fsmVMReader interface{ GetStateCopy() api.ClusterState }

type VMServer struct {
	vmpb.UnimplementedVMServiceServer
	st  *store.Manager
	fsm fsmVMReader
}

func NewVMServer(st *store.Manager, fsm fsmVMReader) *VMServer { return &VMServer{st: st, fsm: fsm} }

func (s *VMServer) ListVMs(ctx context.Context, _ *vmpb.Empty) (*vmpb.ListVMsResponse, error) {
	st := s.fsm.GetStateCopy()
	resp := &vmpb.ListVMsResponse{}
	for _, v := range st.VMs {
		resp.Vms = append(resp.Vms, &vmpb.VM{Id: v.ID, Name: v.Name, NodeId: v.NodeID, Cpu: int32(v.Resources.CPU), Memory: int32(v.Resources.Memory), Disk: int32(v.Resources.Disk), Phase: v.Phase})
	}
	return resp, nil
}
func (s *VMServer) UpsertVM(ctx context.Context, req *vmpb.UpsertVMRequest) (*vmpb.Empty, error) {
	v := req.Vm
	vm := api.VM{ID: v.Id, Name: v.Name, NodeID: v.NodeId, Phase: v.Phase, Resources: api.Resources{CPU: int(v.Cpu), Memory: int(v.Memory), Disk: int(v.Disk)}}
	if vm.NodeID == "" {
		st := s.fsm.GetStateCopy()
		if nid, ok := scheduler.ChooseNode(st, vm); ok {
			vm.NodeID = nid
		}
	}
	if err := s.st.Apply(ctx, store.NewCommand("UpsertVM", vm)); err != nil {
		return nil, err
	}
	return &vmpb.Empty{}, nil
}
func (s *VMServer) DeleteVM(ctx context.Context, req *vmpb.DeleteVMRequest) (*vmpb.Empty, error) {
	if err := s.st.Apply(ctx, store.NewCommand("DeleteVM", req.Id)); err != nil {
		return nil, err
	}
	return &vmpb.Empty{}, nil
}
func (s *VMServer) Migrate(ctx context.Context, req *vmpb.MigrateRequest) (*vmpb.Empty, error) {
	st := s.fsm.GetStateCopy()
	vm, ok := st.VMs[req.Id]
	if !ok {
		return &vmpb.Empty{}, nil
	}
	if req.TargetNode != "" {
		vm.NodeID = req.TargetNode
	} else {
		if nid, ok := scheduler.ChooseNode(st, vm); ok {
			vm.NodeID = nid
		}
	}
	vm.Phase = "Migrating"
	if err := s.st.Apply(ctx, store.NewCommand("UpsertVM", vm)); err != nil {
		return nil, err
	}
	return &vmpb.Empty{}, nil
}

// Additional parity RPCs (stubs; extend proto in real codegen)
func (s *VMServer) Clone(ctx context.Context, srcId, newId string) (*vmpb.Empty, error) {
	st := s.fsm.GetStateCopy()
	vm, ok := st.VMs[srcId]
	if !ok {
		return &vmpb.Empty{}, nil
	}
	vm.ID = newId
	vm.Name = vm.Name + "-clone"
	return s.UpsertVM(ctx, &vmpb.UpsertVMRequest{Vm: &vmpb.VM{Id: vm.ID, Name: vm.Name, NodeId: vm.NodeID, Cpu: int32(vm.Resources.CPU), Memory: int32(vm.Resources.Memory), Disk: int32(vm.Resources.Disk), Phase: vm.Phase}})
}

func (s *VMServer) CloneFromTemplate(ctx context.Context, templateId, newId string) (*vmpb.Empty, error) {
	st := s.fsm.GetStateCopy()
	tpl, ok := st.Templates[templateId]
	if !ok {
		return &vmpb.Empty{}, nil
	}
	vm := api.VM{ID: newId, Name: tpl.Name + "-inst", Resources: tpl.Resources, Phase: "Pending"}
	if nid, ok := scheduler.ChooseNode(st, vm); ok {
		vm.NodeID = nid
	}
	if err := s.st.Apply(ctx, store.NewCommand("UpsertVM", vm)); err != nil {
		return nil, err
	}
	return &vmpb.Empty{}, nil
}
