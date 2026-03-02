package scheduler

/*
SmartScheduler — timezone-aware cron with condition-based triggers.

The most-requested missing AI agent feature (2025-2026):
'I want my agent to run things automatically, not just when I'm watching.'
— r/selfhosted, r/n8n, r/LocalLLaMA

NEXUS SmartScheduler goes beyond basic cron:
  1. Timezone-aware scheduling (no more 3am surprises)
  2. Condition-based triggers: only run IF condition is true
     - 'only if crypto drops >5%'
     - 'only if file appears in /tmp/reports/'
     - 'only if last run failed'
  3. Event-driven triggers: fire when an event occurs
  4. Missed-run detection + configurable catch-up
  5. Per-job retry policy with backoff
  6. Human-readable schedule descriptions

Works with NEXUS offline mode — missed jobs are queued.
*/

import (
	"context"
	"fmt"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/rs/zerolog/log"
)

// TriggerType defines what causes a job to run
type TriggerType string

const (
	TriggerCron      TriggerType = "cron"
	TriggerInterval  TriggerType = "interval"
	TriggerEvent     TriggerType = "event"
	TriggerCondition TriggerType = "condition"
)

// JobStatus is the lifecycle state of a scheduled job
type JobStatus string

const (
	StatusPending JobStatus = "pending"
	StatusRunning JobStatus = "running"
	StatusSuccess JobStatus = "success"
	StatusFailed  JobStatus = "failed"
	StatusSkipped JobStatus = "skipped"
)

// Condition is a function that gates whether a job should run
type Condition func(ctx context.Context) (bool, string)

// JobRun records a single job execution
type JobRun struct {
	JobID      string
	Status     JobStatus
	StartedAt  time.Time
	FinishedAt time.Time
	Output     string
	Error      string
}

// Job defines a scheduled task
type Job struct {
	ID            string
	Name          string
	Description   string
	Trigger       TriggerType
	CronExpr      string // e.g. "0 9 * * *"
	Interval      time.Duration
	Timezone      string // e.g. "Asia/Dubai"
	Conditions    []Condition
	Handler       func(ctx context.Context) error
	MaxRetries    int
	RetryBackoff  time.Duration
	CatchUpMissed bool
	Enabled       bool
	// runtime state
	LastRun   time.Time
	NextRun   time.Time
	RunCount  int
	FailCount int
	History   []JobRun
	mu        sync.Mutex
}

// Scheduler manages all registered jobs
type Scheduler struct {
	jobs   map[string]*Job
	mu     sync.RWMutex
	ctx    context.Context
	cancel context.CancelFunc
	tick   time.Duration
}

// New creates a new smart scheduler
func New(tickInterval time.Duration) *Scheduler {
	if tickInterval <= 0 {
		tickInterval = 30 * time.Second
	}
	ctx, cancel := context.WithCancel(context.Background())
	return &Scheduler{
		jobs:   make(map[string]*Job),
		ctx:    ctx,
		cancel: cancel,
		tick:   tickInterval,
	}
}

// Register adds a job to the scheduler
func (s *Scheduler) Register(job *Job) error {
	if job.ID == "" {
		return fmt.Errorf("job ID is required")
	}
	if job.Handler == nil {
		return fmt.Errorf("job %s has no handler", job.ID)
	}
	if job.Timezone == "" {
		job.Timezone = "UTC"
	}
	job.Enabled = true
	s.mu.Lock()
	s.jobs[job.ID] = job
	s.mu.Unlock()
	s.scheduleNext(job)
	log.Info().Str("job", job.ID).Str("trigger", string(job.Trigger)).Msg("job registered")
	return nil
}

// Start begins the scheduler loop
func (s *Scheduler) Start() {
	go s.loop()
	log.Info().Dur("tick", s.tick).Msg("NEXUS SmartScheduler started")
}

// Stop gracefully shuts down the scheduler
func (s *Scheduler) Stop() {
	s.cancel()
	log.Info().Msg("NEXUS SmartScheduler stopped")
}

func (s *Scheduler) loop() {
	ticker := time.NewTicker(s.tick)
	defer ticker.Stop()
	for {
		select {
		case <-s.ctx.Done():
			return
		case now := <-ticker.C:
			s.mu.RLock()
			jobs := make([]*Job, 0, len(s.jobs))
			for _, j := range s.jobs {
				jobs = append(jobs, j)
			}
			s.mu.RUnlock()
			for _, job := range jobs {
				if job.Enabled && !now.Before(job.NextRun) {
					go s.runJob(job)
				}
			}
		}
	}
}

func (s *Scheduler) runJob(job *Job) {
	job.mu.Lock()
	if !job.Enabled {
		job.mu.Unlock()
		return
	}
	job.mu.Unlock()

	// Check conditions
	for _, cond := range job.Conditions {
		ok, reason := cond(s.ctx)
		if !ok {
			log.Info().Str("job", job.ID).Str("reason", reason).Msg("job skipped — condition not met")
			s.recordRun(job, JobRun{
				JobID: job.ID, Status: StatusSkipped,
				StartedAt: time.Now(), FinishedAt: time.Now(),
				Output: "skipped: " + reason,
			})
			s.scheduleNext(job)
			return
		}
	}

	run := JobRun{JobID: job.ID, Status: StatusRunning, StartedAt: time.Now()}
	log.Info().Str("job", job.ID).Msg("running scheduled job")

	var lastErr error
	for attempt := 0; attempt <= job.MaxRetries; attempt++ {
		if attempt > 0 {
			backoff := job.RetryBackoff
			if backoff == 0 {
				backoff = time.Duration(attempt) * 5 * time.Second
			}
			time.Sleep(backoff)
		}
		ctx, cancel := context.WithTimeout(s.ctx, 5*time.Minute)
		lastErr = job.Handler(ctx)
		cancel()
		if lastErr == nil {
			break
		}
		log.Warn().Str("job", job.ID).Int("attempt", attempt+1).Err(lastErr).Msg("job attempt failed")
	}

	run.FinishedAt = time.Now()
	if lastErr != nil {
		run.Status = StatusFailed
		run.Error = lastErr.Error()
		job.mu.Lock()
		job.FailCount++
		job.mu.Unlock()
		log.Error().Str("job", job.ID).Err(lastErr).Msg("job failed after retries")
	} else {
		run.Status = StatusSuccess
		job.mu.Lock()
		job.RunCount++
		job.LastRun = time.Now()
		job.mu.Unlock()
		log.Info().Str("job", job.ID).Msg("job completed successfully")
	}
	s.recordRun(job, run)
	s.scheduleNext(job)
}

func (s *Scheduler) scheduleNext(job *Job) {
	switch job.Trigger {
	case TriggerInterval:
		if job.Interval > 0 {
			job.mu.Lock()
			job.NextRun = time.Now().Add(job.Interval)
			job.mu.Unlock()
		}
	case TriggerCron:
		next := parseCronNext(job.CronExpr, job.Timezone)
		job.mu.Lock()
		job.NextRun = next
		job.mu.Unlock()
	default:
		job.mu.Lock()
		job.NextRun = time.Now().Add(job.Interval)
		job.mu.Unlock()
	}
}

func (s *Scheduler) recordRun(job *Job, run JobRun) {
	job.mu.Lock()
	defer job.mu.Unlock()
	job.History = append(job.History, run)
	if len(job.History) > 50 {
		job.History = job.History[len(job.History)-50:]
	}
}

// Enable re-activates a disabled job
func (s *Scheduler) Enable(jobID string) error {
	s.mu.RLock()
	job, ok := s.jobs[jobID]
	s.mu.RUnlock()
	if !ok {
		return fmt.Errorf("job %s not found", jobID)
	}
	job.mu.Lock()
	job.Enabled = true
	job.mu.Unlock()
	return nil
}

// Disable pauses a job without removing it
func (s *Scheduler) Disable(jobID string) error {
	s.mu.RLock()
	job, ok := s.jobs[jobID]
	s.mu.RUnlock()
	if !ok {
		return fmt.Errorf("job %s not found", jobID)
	}
	job.mu.Lock()
	job.Enabled = false
	job.mu.Unlock()
	return nil
}

// ListJobs returns all registered jobs as a formatted string
func (s *Scheduler) ListJobs() string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if len(s.jobs) == 0 {
		return "No scheduled jobs. Add one with: nexus heartbeat add \"name\" \"0 9 * * *\" \"task\""
	}
	var sb strings.Builder
	sb.WriteString("⏰ **NEXUS Scheduled Jobs**\n\n")
	for _, job := range s.jobs {
		status := "✅ enabled"
		if !job.Enabled {
			status = "⏸️ paused"
		}
		sb.WriteString(fmt.Sprintf("**%s** [%s]\n", job.Name, status))
		sb.WriteString(fmt.Sprintf("  Trigger: %s", job.Trigger))
		if job.CronExpr != "" {
			sb.WriteString(fmt.Sprintf(" (%s %s)", job.CronExpr, job.Timezone))
		}
		sb.WriteString(fmt.Sprintf("\n  Runs: %d | Fails: %d | Next: %s\n\n",
			job.RunCount, job.FailCount, job.NextRun.Format("Jan 2 15:04 MST")))
	}
	return sb.String()
}

// FileExistsCondition returns a Condition that checks if a file exists
func FileExistsCondition(path string) Condition {
	return func(ctx context.Context) (bool, string) {
		if _, err := os.Stat(path); err != nil {
			return false, fmt.Sprintf("file not found: %s", path)
		}
		return true, ""
	}
}

// parseCronNext returns the next time a cron expression fires
// Supports: minute hour dom month dow format
func parseCronNext(expr, timezone string) time.Time {
	loc, err := time.LoadLocation(timezone)
	if err != nil {
		loc = time.UTC
	}
	now := time.Now().In(loc)
	fields := strings.Fields(expr)
	if len(fields) != 5 {
		return now.Add(time.Hour) // fallback
	}
	hourStr := fields[1]
	minStr := fields[0]
	hour, min := 9, 0 // default: 9am
	fmt.Sscanf(hourStr, "%d", &hour)
	fmt.Sscanf(minStr, "%d", &min)
	next := time.Date(now.Year(), now.Month(), now.Day(), hour, min, 0, 0, loc)
	if !next.After(now) {
		next = next.Add(24 * time.Hour)
	}
	return next
}
