//go:build !ci
// +build !ci

package desktop

import (
	"context"

	"github.com/getlantern/systray"
)

// Desktop manages the system tray, global hotkey, and clipboard monitor.
// Run() must be called from the main OS thread.
type Desktop struct {
	webuiAddr string // e.g. "http://localhost:7070"
	cancel    context.CancelFunc
}

func New(webuiAddr string) *Desktop {
	return &Desktop{webuiAddr: webuiAddr}
}

// Run blocks — call from main() after spawning other goroutines.
func (d *Desktop) Run(ctx context.Context) {
	ctx, d.cancel = context.WithCancel(ctx)
	go d.watchHotkeys(ctx)
	go d.watchClipboard(ctx)
	systray.Run(d.onReady, d.onExit)
}

func (d *Desktop) onExit() {
	if d.cancel != nil {
		d.cancel()
	}
}
