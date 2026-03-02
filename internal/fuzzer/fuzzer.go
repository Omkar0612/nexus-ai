package fuzzer

import (
	"context"
	"fmt"
	"time"

	"github.com/Omkar0612/nexus-ai/internal/plugin/wasm"
	"github.com/rs/zerolog/log"
)

// Report contains the results of an adversarial fuzzing campaign.
type Report struct {
	TargetAgent   string
	TestsRun      int
	Vulnerabilities int
	FailedPayloads []string
	Passed        bool
}

// Engine acts as the "Attacker Agent". It takes a compiled Wasm module from the
// Auto-Forge and forcefully bombards it with edge-case and security payloads
// to ensure the generated code is safe before it enters the Multi-Agent Bus.
type Engine struct {
	timeout time.Duration
}

// New initializes the Agentic Fuzzer.
func New(timeout time.Duration) *Engine {
	if timeout == 0 {
		timeout = 2 * time.Second // Wasm executes fast, so fuzzing limits are tight
	}
	return &Engine{timeout: timeout}
}

// Fuzz attacks the provided Wasm agent module.
func (e *Engine) Fuzz(ctx context.Context, agent *wasm.AgentModule) *Report {
	log.Warn().Str("target", agent.Name()).Msg("🛡️ Initiating Agentic Fuzzing (Neuro-Fuzz) campaign...")

	report := &Report{
		TargetAgent: agent.Name(),
		Passed:      true,
	}

	payloads := append(SecurityPayloads, EdgeCasePayloads...)

	for _, payload := range payloads {
		// Enforce strict timeout for each fuzzing attempt to catch infinite loops
		fuzzCtx, cancel := context.WithTimeout(ctx, e.timeout)
		
		log.Debug().Str("vector", payload.Name).Msg("Injecting attack payload...")
		
		_, err := agent.Execute(fuzzCtx, payload.Input)
		cancel()

		report.TestsRun++

		if err != nil {
			// If context deadline exceeded, the agent got stuck in an infinite loop
			if err == context.DeadlineExceeded {
				log.Error().
					Str("vector", payload.Name).
					Msg("🚨 VULNERABILITY FOUND: Payload triggered an infinite loop (DoS)")
				report.Vulnerabilities++
				report.FailedPayloads = append(report.FailedPayloads, fmt.Sprintf("%s (DoS Loop)", payload.Name))
				report.Passed = false
				continue
			}

			// If the Wasm runtime panicked or returned a fatal error
			log.Error().
				Str("vector", payload.Name).
				Err(err).
				Msg("🚨 VULNERABILITY FOUND: Payload triggered a panic/crash")
			report.Vulnerabilities++
			report.FailedPayloads = append(report.FailedPayloads, fmt.Sprintf("%s (Crash)", payload.Name))
			report.Passed = false
		} else {
			log.Debug().Str("vector", payload.Name).Msg("Target survived payload.")
		}
	}

	if report.Passed {
		log.Info().
			Str("target", agent.Name()).
			Int("tests", report.TestsRun).
			Msg("✅ Fuzzing campaign complete. Target is hardened and secure.")
	} else {
		log.Error().
			Str("target", agent.Name()).
			Int("vulns", report.Vulnerabilities).
			Msg("❌ Fuzzing campaign failed. Rejecting agent deployment.")
	}

	return report
}
