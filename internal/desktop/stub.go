//go:build ci
// +build ci

// Package desktop is stubbed in CI builds.
// The real implementation (tray, hotkey, clipboard) lives in desktop.go / tray.go / hotkey.go
// behind the !ci build tag and requires native GUI libs not available in headless CI.
package desktop

import "context"

// Desktop is a no-op stub used in CI.
type Desktop struct{}

// New returns a no-op Desktop.
func New(_ string) *Desktop { return &Desktop{} }

// Run is a no-op in CI.
func (d *Desktop) Run(_ context.Context) {}
