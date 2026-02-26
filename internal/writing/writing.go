// Package writing provides the NEXUS v1.7 AI Writing Studio.
// Operations: Draft, Rewrite, Summarise, Proofread, Expand, Translate.
// Uses the same LLM router as the rest of NEXUS — zero additional cost.
package writing

import (
	"context"
	"fmt"
	"strings"

	"github.com/Omkar0612/nexus-ai/internal/router"
)

// Style controls the writing tone.
type Style string

const (
	StyleProfessional Style = "professional"
	StyleCasual       Style = "casual"
	StylePersuasive   Style = "persuasive"
	StyleAcademic     Style = "academic"
	StyleCreative     Style = "creative"
)

// Agent is the writing studio agent.
type Agent struct {
	router *router.Router
}

// New creates a writing agent backed by the given LLM router.
func New(r *router.Router) *Agent {
	return &Agent{router: r}
}

// Draft generates a new piece of writing from a topic and style.
func (a *Agent) Draft(ctx context.Context, topic string, style Style, words int) (string, error) {
	prompt := fmt.Sprintf(
		"Write a %s-style piece about: %s\nTarget length: ~%d words.\nOnly output the text, no preamble.",
		style, topic, words,
	)
	return a.router.Complete(ctx, prompt)
}

// Rewrite rewrites existing text in the given style.
func (a *Agent) Rewrite(ctx context.Context, text string, style Style) (string, error) {
	prompt := fmt.Sprintf(
		"Rewrite the following text in a %s style. Keep the meaning, improve clarity and tone.\n\n%s",
		style, text,
	)
	return a.router.Complete(ctx, prompt)
}

// Summarise condenses text to a target word count.
func (a *Agent) Summarise(ctx context.Context, text string, targetWords int) (string, error) {
	prompt := fmt.Sprintf(
		"Summarise the following text in ~%d words. Be concise and capture key points.\n\n%s",
		targetWords, text,
	)
	return a.router.Complete(ctx, prompt)
}

// Proofread returns corrected text and a list of issues found.
func (a *Agent) Proofread(ctx context.Context, text string) (corrected string, issues []string, err error) {
	prompt := fmt.Sprintf(
		"Proofread the following text. Return:\nLINE 1: CORRECTED: <full corrected text>\nLINE 2+: ISSUE: <description of each issue found>\n\n%s",
		text,
	)
	raw, err := a.router.Complete(ctx, prompt)
	if err != nil {
		return "", nil, err
	}
	lines := strings.Split(raw, "\n")
	for _, line := range lines {
		if strings.HasPrefix(line, "CORRECTED:") {
			corrected = strings.TrimPrefix(line, "CORRECTED: ")
		} else if strings.HasPrefix(line, "ISSUE:") {
			issues = append(issues, strings.TrimPrefix(line, "ISSUE: "))
		}
	}
	if corrected == "" {
		corrected = raw
	}
	return corrected, issues, nil
}

// Expand takes a bullet list or outline and expands it into full prose.
func (a *Agent) Expand(ctx context.Context, outline string, style Style) (string, error) {
	prompt := fmt.Sprintf(
		"Expand the following outline into full %s prose. Each bullet becomes a paragraph.\n\n%s",
		style, outline,
	)
	return a.router.Complete(ctx, prompt)
}

// Translate translates text to the target language.
func (a *Agent) Translate(ctx context.Context, text, targetLang string) (string, error) {
	prompt := fmt.Sprintf(
		"Translate the following text to %s. Output only the translation, no explanation.\n\n%s",
		targetLang, text,
	)
	return a.router.Complete(ctx, prompt)
}
