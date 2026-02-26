//go:build !ci
// +build !ci

package desktop

import (
	"context"
	"log"
)

// watchHotkeys is stubbed until golang.design/x/hotkey is added back.
// To enable: go get golang.design/x/hotkey and restore full implementation.
func (d *Desktop) watchHotkeys(ctx context.Context) {
	log.Println("[desktop] hotkey listener stub — full impl in v1.7")
	<-ctx.Done()
}
