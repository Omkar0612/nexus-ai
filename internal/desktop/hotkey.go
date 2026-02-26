//go:build !ci
// +build !ci

package desktop

import (
	"context"
	"log"

	"golang.design/x/hotkey"
)

// watchHotkeys registers Ctrl+Shift+Space as the NEXUS quick-prompt trigger.
func (d *Desktop) watchHotkeys(ctx context.Context) {
	hk := hotkey.New(
		[]hotkey.Modifier{hotkey.ModCtrl, hotkey.ModShift},
		hotkey.KeySpace,
	)
	if err := hk.Register(); err != nil {
		log.Printf("[desktop] hotkey register failed: %v", err)
		return
	}
	defer hk.Unregister()
	log.Println("[desktop] hotkey Ctrl+Shift+Space registered")

	for {
		select {
		case <-hk.Keydown():
			log.Println("[desktop] NEXUS hotkey triggered")
			openBrowser(d.webuiAddr)
		case <-ctx.Done():
			return
		}
	}
}
