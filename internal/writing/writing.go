// Package writing provides the NEXUS v1.7 AI Writing Studio.
// Operations: Draft, Rewrite, Summarise, Proofread, Expand, Translate.
// Backed by the shared LLM router — zero additional cost.
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
	r *router.Router
}

// New creates a writing agent backed by the given LLM router.
func New(r *router.Router) *Agent {
	return &Agent{r: r}
}

// completeText is an internal helper that maps a single prompt to the router's
// (ctx, systemPrompt, userMsg) signature.
func (a *Agent) completeText(ctx context.Context, system, user string) (string, error) {
	res, err := a.r.Complete(ctx, system, user)
	if err != nil {
		return "", fmt.Errorf("writing: llm: %w", err)
	}
	return res.Content, nil
}

// Draft generates a new piece of writing from a topic and style.
func (a *Agent) Draft(ctx context.Context, topic string, style Style, words int) (string, error) {
	system := "You are an expert writer. Output only the requested text with no preamble."
	user := fmt.Sprintf(
		"Write a %s-style piece about: %s\nTarget length: ~%d words.",
		style, topic, words,
	)
	return a.completeText(ctx, system, user)
}

// Rewrite rewrites existing text in the given style.
func (a *Agent) Rewrite(ctx context.Context, text string, style Style) (string, error) {
	system := "You are an expert editor. Output only the rewritten text with no preamble."
	user := fmt.Sprintf(
		"Rewrite the following text in a %s style. Keep the meaning, improve clarity and tone.\n\n%s",
		style, text,
	)
	return a.completeText(ctx, system, user)
}

// Summarise condenses text to a target word count.
func (a *Agent) Summarise(ctx context.Context, text string, targetWords int) (string, error) {
	system := "You are a precise summariser. Output only the summary with no preamble."
	user := fmt.Sprintf(
		"Summarise the following text in ~%d words. Be concise and capture key points.\n\n%s",
		targetWords, text,
	)
	return a.completeText(ctx, system, user)
}

// Proofread returns corrected text and a list of issues found.
func (a *Agent) Proofread(ctx context.Context, text string) (corrected string, issues []string, err error) {
	system := "You are a professional proofreader. Follow the output format exactly."
	user := fmt.Sprintf(
		"Proofread the following text. Respond with:\nCORRECTED: <full corrected text>\nISSUE: <one issue per line, prefix each with ISSUE:>\n\n%s",
		text,
	)
	raw, err := a.completeText(ctx, system, user)
	if err != nil {
		return "", nil, err
	}
	for _, line := range strings.Split(raw, "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "CORRECTED:") {
			corrected = strings.TrimSpace(strings.TrimPrefix(line, "CORRECTED:"))
		} else if strings.HasPrefix(line, "ISSUE:") {
			issues = append(issues, strings.TrimSpace(strings.TrimPrefix(line, "ISSUE:")))
		}
	}
	if corrected == "" {
		corrected = raw // fallback: return raw if LLM didn't follow format
	}
	return corrected, issues, nil
}

// Expand takes a bullet list or outline and expands it into full prose.
func (a *Agent) Expand(ctx context.Context, outline string, style Style) (string, error) {
	system := "You are an expert writer. Output only the expanded text with no preamble."
	user := fmt.Sprintf(
		"Expand the following outline into full %s prose. Each bullet becomes a paragraph.\n\n%s",
		style, outline,
	)
	return a.completeText(ctx, system, user)
}

// Translate translates text to the target language.
func (a *Agent) Translate(ctx context.Context, text, targetLang string) (string, error) {
	system := fmt.Sprintf("You are a professional translator. Translate to %s. Output only the translation.", targetLang)
	user := text
	return a.completeText(ctx, system, user)
}
