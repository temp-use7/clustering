package templatepb

import (
	"context"

	"google.golang.org/grpc"
)

type Empty struct{}

type Template struct {
	Id        string
	Name      string
	BaseImage string
	Cpu       int32
	Memory    int32
	Disk      int32
}

type ListTemplatesResponse struct{ Templates []*Template }
type UpsertTemplateRequest struct{ Template *Template }
type DeleteTemplateRequest struct{ Id string }
type InstantiateRequest struct {
	TemplateId string
	NewId      string
}

type TemplateServiceServer interface {
	ListTemplates(context.Context, *Empty) (*ListTemplatesResponse, error)
	UpsertTemplate(context.Context, *UpsertTemplateRequest) (*Empty, error)
	DeleteTemplate(context.Context, *DeleteTemplateRequest) (*Empty, error)
	Instantiate(context.Context, *InstantiateRequest) (*Empty, error)
}

type UnimplementedTemplateServiceServer struct{}

func RegisterTemplateServiceServer(s *grpc.Server, srv TemplateServiceServer) {}

