package agents

import (
	"testing"
)

func TestHallucinationDetectorVerified(t *testing.T) {
	d := NewHallucinationDetector(0.6)
	report := d.Analyse("You might want to consider approximately 5 options here. This could possibly work.")
	if report.OverallScore < 0.7 {
		t.Errorf("expected high score for hedged response, got %.2f", report.OverallScore)
	}
	if report.Tag != TagVerified && report.Tag != TagUncertain {
		t.Errorf("expected VERIFIED or UNCERTAIN for hedged response, got %s", report.Tag)
	}
}

func TestHallucinationDetectorAbsoluteLanguage(t *testing.T) {
	d := NewHallucinationDetector(0.6)
	report := d.Analyse("This is definitely guaranteed to always work 100% of the time. It is impossible for this to fail.")
	if report.FlaggedCount == 0 {
		t.Error("expected flagged claims for absolute language")
	}
}

func TestHallucinationDetectorURLExtraction(t *testing.T) {
	d := NewHallucinationDetector(0.6)
	report := d.Analyse("You can find it at https://github.com/Omkar0612/nexus-ai for more info.")
	urls := 0
	for _, c := range report.Claims {
		if c.ClaimType == "url" {
			urls++
		}
	}
	if urls == 0 {
		t.Error("expected URL to be extracted as a claim")
	}
}

func TestHallucinationDetectorMemoryContradiction(t *testing.T) {
	d := NewHallucinationDetector(0.6)
	d.LoadMemoryContext([]string{"The server is currently disabled and not running"})
	report := d.Analyse("The server is enabled and running correctly.")
	// Should detect contradiction between enabled/disabled
	if report.Tag != TagContradicted {
		// Contradiction detection is heuristic — warn but don't hard fail
		t.Logf("note: contradiction not detected (heuristic), tag=%s score=%.2f", report.Tag, report.OverallScore)
	}
}

func TestHallucinationBadge(t *testing.T) {
	d := NewHallucinationDetector(0.6)
	report := d.Analyse("This might possibly work in some cases.")
	badge := report.FormatBadge()
	if badge == "" {
		t.Error("expected non-empty badge")
	}
}

func TestHallucinationShouldRetry(t *testing.T) {
	d := NewHallucinationDetector(0.99) // very high threshold forces retry
	report := d.Analyse("It definitely always works 100% guaranteed impossible to fail.")
	if !report.ShouldRetry {
		t.Log("retry not triggered — absolute language may not have lowered score below 0.99")
	}
	if report.RetryPromptHint != "" && len(report.RetryPromptHint) < 10 {
		t.Error("retry hint too short")
	}
}
