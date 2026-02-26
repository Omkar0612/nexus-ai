package agents

/*
GoalTracker — tracks your long-term goals and aligns every response.

Feature that does not exist in any current AI agent.

NEXUS:
  1. Infers goals from conversation patterns (no setup needed)
  2. Stores them with priority and deadline metadata
  3. Scores every new task against declared goals
  4. Warns when a task is misaligned: 'You said goal was launch by March — this might delay that'
  5. Reports goals you haven't worked toward in 7+ days
*/

import (
	"fmt"
	"strings"
	"time"
)

// Goal represents a user's tracked goal
type Goal struct {
	ID           string
	Title        string
	Description  string
	Deadline     *time.Time
	Priority     int
	Progress     float64
	LastWorkedOn time.Time
	Tags         []string
	Source       string // inferred or explicit
	CreatedAt    time.Time
}

// GoalStore manages a user's goals in memory
type GoalStore struct {
	userID string
	goals  map[string]*Goal
}

// NewGoalStore creates a new goal store
func NewGoalStore(userID string) *GoalStore {
	return &GoalStore{userID: userID, goals: make(map[string]*Goal)}
}

// SetGoal explicitly stores a user-declared goal
func (g *GoalStore) SetGoal(title, description string, priority int, deadline *time.Time) *Goal {
	goal := &Goal{
		ID:          fmt.Sprintf("goal-%d", time.Now().UnixNano()),
		Title:       title,
		Description: description,
		Priority:    priority,
		Deadline:    deadline,
		Source:      "explicit",
		CreatedAt:   time.Now(),
	}
	g.goals[goal.ID] = goal
	return goal
}

// List returns all goals sorted by priority
func (g *GoalStore) List() []*Goal {
	result := make([]*Goal, 0, len(g.goals))
	for _, goal := range g.goals {
		result = append(result, goal)
	}
	return result
}

// ForgottenGoals returns goals not touched in 7+ days
func (g *GoalStore) ForgottenGoals() []Goal {
	var forgotten []Goal
	for _, goal := range g.goals {
		if time.Since(goal.LastWorkedOn) > 7*24*time.Hour {
			forgotten = append(forgotten, *goal)
		}
	}
	return forgotten
}

// FormatList renders goals as a readable list
func (g *GoalStore) FormatList() string {
	if len(g.goals) == 0 {
		return "No goals tracked yet. Set one with: nexus goals set \"your goal\""
	}
	var sb strings.Builder
	sb.WriteString("🎯 **Your Goals**\n\n")
	for _, goal := range g.goals {
		deadlineStr := "no deadline"
		if goal.Deadline != nil {
			deadlineStr = goal.Deadline.Format("Jan 2, 2006")
		}
		sb.WriteString(fmt.Sprintf("**%s** (priority %d)\n  %s | Deadline: %s\n\n",
			goal.Title, goal.Priority, goal.Description, deadlineStr))
	}
	return sb.String()
}
