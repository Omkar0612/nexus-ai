package fuzzer

// AttackPayload represents a single test case designed to break the agent.
type AttackPayload struct {
	Name        string
	Input       string
	Description string
}

// SecurityPayloads contains common attack vectors to test against generated code.
var SecurityPayloads = []AttackPayload{
	{
		Name:        "SQL Injection",
		Input:       "'; DROP TABLE users; --",
		Description: "Attempts to break database sanitization",
	},
	{
		Name:        "Path Traversal",
		Input:       "../../../../etc/passwd",
		Description: "Attempts to read host filesystem (should be blocked by WASI)",
	},
	{
		Name:        "Command Injection",
		Input:       "127.0.0.1; cat /etc/shadow",
		Description: "Attempts to execute arbitrary host commands",
	},
}

// EdgeCasePayloads contains inputs designed to cause panics or out-of-bounds errors.
var EdgeCasePayloads = []AttackPayload{
	{
		Name:        "Null Byte",
		Input:       "test\x00data",
		Description: "Attempts to truncate strings in C/Wasm memory boundaries",
	},
	{
		Name:        "Maximum Buffer",
		Input:       generateHugeString(10 * 1024 * 1024), // 10MB string
		Description: "Attempts to trigger an Out Of Memory (OOM) panic in the Wasm runtime",
	},
	{
		Name:        "Unicode Overflow",
		Input:       "🚀🔥" + generateHugeString(5000), // Emoji byte-length mismatch
		Description: "Attempts to break rune/byte length counting logic",
	},
}

func generateHugeString(size int) string {
	b := make([]byte, size)
	for i := range b {
		b[i] = 'A'
	}
	return string(b)
}
