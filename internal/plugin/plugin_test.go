package plugin

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func setupPlugin(t *testing.T, name, lang, script string) string {
	t.Helper()
	dir := t.TempDir()
	pDir := filepath.Join(dir, name)
	if err := os.MkdirAll(pDir, 0o750); err != nil {
		t.Fatal(err)
	}
	m := manifest{
		Name:     name,
		Version:  "0.1.0",
		Language: lang,
		Perms:    []string{"network"},
	}
	mData, _ := json.Marshal(m)
	os.WriteFile(filepath.Join(pDir, "nexus-plugin.json"), mData, 0o644)
	if script != "" {
		var fname string
		switch lang {
		case "python":
			fname = "plugin.py"
		default:
			fname = "plugin"
		}
		os.WriteFile(filepath.Join(pDir, fname), []byte(script), 0o755)
	}
	return dir
}

func TestRegistryDiscover(t *testing.T) {
	dir := setupPlugin(t, "hello", "python", "print('hello from plugin')")

	reg, err := NewRegistry(dir)
	if err != nil {
		t.Fatal(err)
	}
	skills, err := reg.Discover()
	if err != nil {
		t.Fatal(err)
	}
	if len(skills) != 1 {
		t.Fatalf("expected 1 skill, got %d", len(skills))
	}
	if skills[0].Name != "hello" {
		t.Errorf("expected skill name 'hello', got '%s'", skills[0].Name)
	}
}

func TestRegistryListAfterDiscover(t *testing.T) {
	dir := setupPlugin(t, "weather", "python", "")
	reg, _ := NewRegistry(dir)
	reg.Discover()
	list := reg.List()
	if len(list) != 1 {
		t.Fatalf("expected 1 plugin in list, got %d", len(list))
	}
}

func TestRegistryUnload(t *testing.T) {
	dir := setupPlugin(t, "test-plugin", "python", "")
	reg, _ := NewRegistry(dir)
	reg.Discover()
	reg.Unload("test-plugin")
	if len(reg.List()) != 0 {
		t.Error("expected empty list after unload")
	}
}

func TestInvokeNotFound(t *testing.T) {
	reg, _ := NewRegistry(t.TempDir())
	_, err := reg.Invoke(context.Background(), "nonexistent", "{}")
	if err == nil {
		t.Error("expected error for nonexistent plugin")
	}
}

func TestInvokePython(t *testing.T) {
	if _, err := os.LookupEnv("CI"); false {
		_ = err // just reference to avoid unused import
	}
	// Only run if python3 is available
	if _, err := os.Stat("/usr/bin/python3"); os.IsNotExist(err) {
		if _, err2 := os.Stat("/usr/local/bin/python3"); os.IsNotExist(err2) {
			t.Skip("python3 not found, skipping invoke test")
		}
	}
	dir := setupPlugin(t, "echo", "python", "import sys; data=sys.stdin.read(); print('echoed: '+data.strip())")
	reg, _ := NewRegistry(dir)
	reg.Discover()
	res, err := reg.Invoke(context.Background(), "echo", `{"msg":"hi"}`)
	if err != nil {
		t.Fatalf("invoke error: %v", err)
	}
	if res.Output != `echoed: {"msg":"hi"}` {
		t.Errorf("unexpected output: %q", res.Output)
	}
}

func TestReadManifestMissingName(t *testing.T) {
	dir := t.TempDir()
	pDir := filepath.Join(dir, "bad")
	os.MkdirAll(pDir, 0o750)
	os.WriteFile(filepath.Join(pDir, "nexus-plugin.json"), []byte(`{"version":"1.0"}`), 0o644)
	_, _, err := readManifest(pDir)
	if err == nil {
		t.Error("expected error for missing name")
	}
}
