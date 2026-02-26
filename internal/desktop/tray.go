//go:build !ci
// +build !ci

package desktop

import (
	_ "embed"
	"os/exec"
	"runtime"
)

//go:embed icons/nexus.ico
var nexusIcon []byte

// runTray starts the system tray. Import systray locally when building for desktop.
// To enable: go get github.com/energye/systray and uncomment the lines below.
// We keep this file dep-free in the repo so CI never needs to resolve systray.
func (d *Desktop) runTray() {
	// TODO(v1.7): import systray and call systray.Run(d.onReady, d.onExit)
	// Stub: just block until ctx is cancelled via onExit
}

func (d *Desktop) onReady() {
	// TODO(v1.7): systray.SetIcon(nexusIcon) etc.
	_ = nexusIcon
}

func openBrowser(url string) {
	var cmd string
	switch runtime.GOOS {
	case "windows":
		cmd = "explorer"
	case "darwin":
		cmd = "open"
	default:
		cmd = "xdg-open"
	}
	exec.Command(cmd, url).Start() //nolint:errcheck
}
