package grpcapi

import (
	"context"

	templatepb "clustering/api/proto/template"
	"clustering/pkg/api"
	"clustering/pkg/scheduler"
	"clustering/pkg/store"
)

type fsmTplReader interface{ GetStateCopy() api.ClusterState }

type TemplateServer struct {
	templatepb.UnimplementedTemplateServiceServer
	st  *store.Manager
	fsm fsmTplReader
}

func NewTemplateServer(st *store.Manager, fsm fsmTplReader) *TemplateServer {
	return &TemplateServer{st: st, fsm: fsm}
}

func (s *TemplateServer) ListTemplates(ctx context.Context, _ *templatepb.Empty) (*templatepb.ListTemplatesResponse, error) {
	st := s.fsm.GetStateCopy()
	resp := &templatepb.ListTemplatesResponse{}
	for _, t := range st.Templates {
		resp.Templates = append(resp.Templates, &templatepb.Template{Id: t.ID, Name: t.Name, BaseImage: t.BaseImage, Cpu: int32(t.Resources.CPU), Memory: int32(t.Resources.Memory), Disk: int32(t.Resources.Disk)})
	}
	return resp, nil
}

func (s *TemplateServer) UpsertTemplate(ctx context.Context, req *templatepb.UpsertTemplateRequest) (*templatepb.Empty, error) {
	t := req.Template
	tpl := api.VMTemplate{ID: t.Id, Name: t.Name, BaseImage: t.BaseImage, Resources: api.Resources{CPU: int(t.Cpu), Memory: int(t.Memory), Disk: int(t.Disk)}}
	if err := s.st.Apply(ctx, store.NewCommand("UpsertTemplate", tpl)); err != nil {
		return nil, err
	}
	return &templatepb.Empty{}, nil
}

func (s *TemplateServer) DeleteTemplate(ctx context.Context, req *templatepb.DeleteTemplateRequest) (*templatepb.Empty, error) {
	if err := s.st.Apply(ctx, store.NewCommand("DeleteTemplate", req.Id)); err != nil {
		return nil, err
	}
	return &templatepb.Empty{}, nil
}

func (s *TemplateServer) Instantiate(ctx context.Context, req *templatepb.InstantiateRequest) (*templatepb.Empty, error) {
	stCopy := s.fsm.GetStateCopy()
	tpl, ok := stCopy.Templates[req.TemplateId]
	if !ok {
		return &templatepb.Empty{}, nil
	}
	vm := api.VM{ID: req.NewId, Name: tpl.Name + "-inst", Resources: tpl.Resources, Phase: "Pending"}
	if nid, ok := scheduler.ChooseNode(stCopy, vm); ok {
		vm.NodeID = nid
	}
	if err := s.st.Apply(ctx, store.NewCommand("UpsertVM", vm)); err != nil {
		return nil, err
	}
	return &templatepb.Empty{}, nil
}
