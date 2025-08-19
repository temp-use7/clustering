package mock

import (
	"context"
	"log"
)

type Driver struct{}

func New() *Driver { return &Driver{} }

func (d *Driver) Start(ctx context.Context, vmID string) error {
	log.Printf("mock start vm %s", vmID)
	return nil
}
func (d *Driver) Stop(ctx context.Context, vmID string) error {
	log.Printf("mock stop vm %s", vmID)
	return nil
}
func (d *Driver) Migrate(ctx context.Context, vmID string, targetNode string) error {
	log.Printf("mock migrate vm %s -> %s", vmID, targetNode)
	return nil
}
