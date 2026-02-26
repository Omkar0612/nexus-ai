package desktop

import "context"

// watchClipboard is a stub for v1.6.
// Full implementation in v1.7: poll golang.design/x/clipboard,
// trigger semantic search on new text content copied by the user.
func (d *Desktop) watchClipboard(ctx context.Context) {
	<-ctx.Done()
}
