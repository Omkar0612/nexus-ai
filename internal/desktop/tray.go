//go:build !ci
// +build !ci

package desktop

import (
	_ "embed"
	"os/exec"
	"runtime"

	"github.com/getlantern/systray"
)

//go:embed icons/nexus.ico
var nexusIcon []byte

func (d *Desktop) onReady() {
	systray.SetIcon(nexusIcon)
	systray.SetTitle("NEXUS AI")
	systray.SetTooltip("NEXUS AI — Your local AI assistant")

	mOpen  := systray.AddMenuItem("Open Dashboard", "Open the web UI in browser")
	mPause := systray.AddMenuItem("Pause Agents", "Pause all running agents")
	systray.AddSeparator()
	mQuit  := systray.AddMenuItem("Quit NEXUS", "Exit NEXUS AI")

	go func() {
		for {
			select {
			case <-mOpen.ClickedCh:
				openBrowser(d.webuiAddr)
			case <-mPause.ClickedCh:
				// TODO: call internal/router PauseAll() via HITL gate
				systray.SetTooltip("NEXUS AI — Agents paused")
			case <-mQuit.ClickedCh:
				systray.Quit()
			}
		}
	}()
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
