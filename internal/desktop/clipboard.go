//go:build !ci
// +build !ci

package desktop

import "context"

// watchClipboard is a stub for v1.6.
// Full implementation in v1.7.
func (d *Desktop) watchClipboard(ctx context.Context) {
	<-ctx.Done()
}
