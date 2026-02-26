package agents

/*
EmotionalIntelligence — detects user emotional state and adapts response tone.

Detects: frustrated, stressed, urgent, excited, confused, neutral
Adapts: verbosity, formality, lead-with style, emoji usage, max length

Nobody else has this in any open-source AI agent.
Source gap: 'AI can't read a room' — AI agent limitations report 2026
*/

import "strings"

// EmotionalContext holds the detected emotional state of a user message
type EmotionalContext struct {
	PrimaryEmotion string
	Urgency        float64
	Stress         float64
	Confidence     float64
}

// ToneProfile describes how NEXUS should respond given an emotional context
type ToneProfile struct {
	Verbosity    string // brief, normal, detailed
	Formality    string // casual, professional, empathetic
	LeadWithWhat string // solution, acknowledgment, question, data
	UseEmoji     bool
	MaxLength    int
}

// AnalyzeEmotion detects emotional context from message text
func AnalyzeEmotion(text string) EmotionalContext {
	lower := strings.ToLower(text)
	ctx := EmotionalContext{PrimaryEmotion: "neutral", Confidence: 0.5}

	frustration := countKW(lower, []string{"not working", "broken", "ugh", "useless", "terrible", "hate", "frustrated", "stupid", "won't work"})
	urgency := countKW(lower, []string{"urgent", "asap", "immediately", "emergency", "critical", "deadline", "quickly", "fast"})
	stress := countKW(lower, []string{"can't", "overwhelmed", "too much", "stressed", "pressure", "panic", "behind"})
	excitement := countKW(lower, []string{"amazing", "awesome", "great", "love", "excited", "finally", "perfect", "!!"})
	confusion := countKW(lower, []string{"confused", "don't understand", "what does", "unclear", "explain", "not sure"})

	scores := map[string]int{
		"frustrated": frustration * 3,
		"urgent":     urgency * 3,
		"stressed":   stress * 2,
		"excited":    excitement * 2,
		"confused":   confusion * 2,
	}

	maxEmotion, maxScore := "neutral", 0
	for emotion, score := range scores {
		if score > maxScore {
			maxScore = score
			maxEmotion = emotion
		}
	}
	ctx.PrimaryEmotion = maxEmotion
	ctx.Urgency = clampF(float64(urgency) / 3.0)
	ctx.Stress = clampF(float64(stress) / 3.0)
	ctx.Confidence = clampF(float64(maxScore) / 6.0)
	if strings.Count(text, "!") > 2 {
		ctx.Urgency = clampF(ctx.Urgency + 0.3)
	}
	return ctx
}

// AdaptTone returns the appropriate tone profile for an emotional context
func AdaptTone(ctx EmotionalContext) ToneProfile {
	switch ctx.PrimaryEmotion {
	case "frustrated":
		return ToneProfile{Verbosity: "brief", Formality: "empathetic", LeadWithWhat: "acknowledgment", UseEmoji: false, MaxLength: 400}
	case "urgent":
		return ToneProfile{Verbosity: "brief", Formality: "professional", LeadWithWhat: "solution", UseEmoji: false, MaxLength: 300}
	case "stressed":
		return ToneProfile{Verbosity: "normal", Formality: "empathetic", LeadWithWhat: "acknowledgment", UseEmoji: false, MaxLength: 500}
	case "excited":
		return ToneProfile{Verbosity: "normal", Formality: "casual", LeadWithWhat: "solution", UseEmoji: true, MaxLength: 600}
	case "confused":
		return ToneProfile{Verbosity: "detailed", Formality: "casual", LeadWithWhat: "question", UseEmoji: false, MaxLength: 800}
	default:
		return ToneProfile{Verbosity: "normal", Formality: "professional", LeadWithWhat: "solution", UseEmoji: false, MaxLength: 600}
	}
}

// BuildSystemPromptSuffix adds tone instructions to an LLM system prompt
func (t ToneProfile) BuildSystemPromptSuffix() string {
	var parts []string
	switch t.Formality {
	case "empathetic":
		parts = append(parts, "The user seems frustrated or stressed. Acknowledge their situation in 1 sentence before helping.")
	case "casual":
		parts = append(parts, "The user is in a positive mood. Match their energy. Be friendly.")
	case "professional":
		parts = append(parts, "Be concise and professional. Get to the point immediately.")
	}
	switch t.Verbosity {
	case "brief":
		parts = append(parts, "Keep your response under 150 words. Bullet points only. No preamble.")
	case "detailed":
		parts = append(parts, "Be thorough and explain step by step.")
	}
	if !t.UseEmoji {
		parts = append(parts, "Do not use emojis.")
	}
	return strings.Join(parts, " ")
}

func countKW(text string, keywords []string) int {
	count := 0
	for _, kw := range keywords {
		if strings.Contains(text, kw) {
			count++
		}
	}
	return count
}

func clampF(v float64) float64 {
	if v > 1.0 {
		return 1.0
	}
	if v < 0.0 {
		return 0.0
	}
	return v
}
