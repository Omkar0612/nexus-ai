//go:build !ci
// +build !ci

package desktop

import "context"

// Desktop manages the system tray, global hotkey, and clipboard monitor.
// Run() must be called from the main OS thread.
type Desktop struct {
	webuiAddr string
	cancel    context.CancelFunc
}

// New creates a Desktop instance.
func New(webuiAddr string) *Desktop {
	return &Desktop{webuiAddr: webuiAddr}
}

// Run blocks — call from main() after spawning other goroutines.
// Initialises system tray, hotkey listener and clipboard watcher.
func (d *Desktop) Run(ctx context.Context) {
	ctx, d.cancel = context.WithCancel(ctx)
	go d.watchHotkeys(ctx)
	go d.watchClipboard(ctx)
	d.runTray()
}

func (d *Desktop) onExit() {
	if d.cancel != nil {
		d.cancel()
	}
}
