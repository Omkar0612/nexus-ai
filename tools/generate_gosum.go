//go:build ignore
// +build ignore

// Run this locally to regenerate go.sum:
//
//   go run tools/generate_gosum.go
//
// Or simply:
//
//   go mod tidy

package main

import (
	"fmt"
	"os"
	"os/exec"
)

func main() {
	cmds := [][]string{
		{"go", "mod", "tidy"},
		{"go", "mod", "download"},
		{"go", "mod", "verify"},
	}
	for _, args := range cmds {
		cmd := exec.Command(args[0], args[1:]...)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			fmt.Fprintf(os.Stderr, "failed: %v\n", err)
			os.Exit(1)
		}
	}
	fmt.Println("go.sum regenerated successfully")
}
