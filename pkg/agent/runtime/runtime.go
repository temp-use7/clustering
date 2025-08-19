package runtime

import "context"

// VMRuntime abstracts VM lifecycle operations on a node.
type VMRuntime interface {
	Start(ctx context.Context, vmID string) error
	Stop(ctx context.Context, vmID string) error
	Migrate(ctx context.Context, vmID string, targetNode string) error
}
