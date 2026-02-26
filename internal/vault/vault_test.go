package vault

import (
	"path/filepath"
	"strings"
	"testing"
)

func openTestVault(t *testing.T) *Vault {
	t.Helper()
	v, err := Open(filepath.Join(t.TempDir(), "vault.db"), "test-passphrase")
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	return v
}

func TestVaultStoreAndGet(t *testing.T) {
	v := openTestVault(t)

	if err := v.Store("GROQ_API_KEY", "gsk_supersecret", "api_key", "business"); err != nil {
		t.Fatalf("Store: %v", err)
	}

	val, err := v.Get("GROQ_API_KEY")
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if val != "gsk_supersecret" {
		t.Errorf("expected gsk_supersecret, got %q", val)
	}
}

func TestVaultList(t *testing.T) {
	v := openTestVault(t)

	_ = v.Store("KEY_A", "val_a", "api_key", "business")
	_ = v.Store("KEY_B", "val_b", "note", "personal")

	biz, err := v.List("business")
	if err != nil {
		t.Fatalf("List business: %v", err)
	}
	if len(biz) != 1 || biz[0].Name != "KEY_A" {
		t.Errorf("expected 1 business secret KEY_A, got %+v", biz)
	}

	all, err := v.List("")
	if err != nil {
		t.Fatalf("List all: %v", err)
	}
	if len(all) != 2 {
		t.Errorf("expected 2 secrets, got %d", len(all))
	}
}

func TestVaultDelete(t *testing.T) {
	v := openTestVault(t)

	_ = v.Store("TMP_KEY", "to-be-deleted", "api_key", "personal")
	if err := v.Delete("TMP_KEY"); err != nil {
		t.Fatalf("Delete: %v", err)
	}
	_, err := v.Get("TMP_KEY")
	if err == nil {
		t.Error("expected error after deletion, got nil")
	}
}

func TestVaultRedactPrompt(t *testing.T) {
	v := openTestVault(t)

	secretVal := "gsk_this_is_a_real_secret_value_abc123"
	_ = v.Store("MY_KEY", secretVal, "api_key", "business")

	prompt := "Please call the API with key: " + secretVal + " and let me know."
	redacted := v.RedactPrompt(prompt)

	if strings.Contains(redacted, secretVal) {
		t.Errorf("secret value leaked in redacted prompt: %s", redacted)
	}
	if !strings.Contains(redacted, "[REDACTED:") {
		t.Errorf("expected [REDACTED:...] placeholder, got: %s", redacted)
	}
}

func TestVaultWrongPassphrase(t *testing.T) {
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "vault.db")

	// Store with correct passphrase
	v1, _ := Open(dbPath, "correct-pass")
	_ = v1.Store("SECRET", "my-secret-value", "api_key", "personal")

	// Try to read with wrong passphrase
	v2, _ := Open(dbPath, "wrong-pass")
	_, err := v2.Get("SECRET")
	if err == nil {
		t.Error("expected decryption error with wrong passphrase")
	}
}
