package calendar

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// GoogleProvider implements CalendarProvider using the Google Calendar REST API.
// Uses OAuth2 Bearer token — get a free token at console.cloud.google.com.
type GoogleProvider struct {
	token      string
	calendarID string
	client     *http.Client
}

// NewGoogle creates a Google Calendar provider.
// calendarID is typically "primary" for the user's default calendar.
func NewGoogle(oauthToken, calendarID string) *GoogleProvider {
	return &GoogleProvider{
		token:      oauthToken,
		calendarID: calendarID,
		client:     &http.Client{Timeout: 15 * time.Second},
	}
}

func (g *GoogleProvider) Name() Provider { return ProviderGoogle }

const gcalBase = "https://www.googleapis.com/calendar/v3"

type gcalEventList struct {
	Items []gcalEvent `json:"items"`
}

type gcalEvent struct {
	ID      string      `json:"id"`
	Summary string      `json:"summary"`
	Desc    string      `json:"description"`
	Loc     string      `json:"location"`
	Start   gcalTime    `json:"start"`
	End     gcalTime    `json:"end"`
	Attend  []gcalEmail `json:"attendees"`
	Recur   []string    `json:"recurrence"`
}

type gcalTime struct {
	DateTime string `json:"dateTime"`
	Date     string `json:"date"`
}

type gcalEmail struct {
	Email string `json:"email"`
}

func (g *GoogleProvider) ListEvents(ctx context.Context, from, to time.Time) ([]Event, error) {
	url := fmt.Sprintf("%s/calendars/%s/events?timeMin=%s&timeMax=%s&singleEvents=true&orderBy=startTime",
		gcalBase, g.calendarID,
		from.UTC().Format(time.RFC3339),
		to.UTC().Format(time.RFC3339),
	)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+g.token)
	resp, err := g.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("google calendar: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("google calendar: status %d", resp.StatusCode)
	}
	var list gcalEventList
	if err := json.NewDecoder(resp.Body).Decode(&list); err != nil {
		return nil, err
	}
	events := make([]Event, 0, len(list.Items))
	for _, item := range list.Items {
		var start, end time.Time
		allDay := false
		if item.Start.DateTime != "" {
			start, _ = time.Parse(time.RFC3339, item.Start.DateTime)
			end, _ = time.Parse(time.RFC3339, item.End.DateTime)
		} else {
			start, _ = time.Parse("2006-01-02", item.Start.Date)
			end, _ = time.Parse("2006-01-02", item.End.Date)
			allDay = true
		}
		attendees := make([]string, len(item.Attend))
		for i, a := range item.Attend {
			attendees[i] = a.Email
		}
		events = append(events, Event{
			ID:          item.ID,
			Title:       item.Summary,
			Description: item.Desc,
			Location:    item.Loc,
			Start:       start,
			End:         end,
			AllDay:      allDay,
			Recurring:   len(item.Recur) > 0,
			Attendees:   attendees,
			CalendarID:  g.calendarID,
			Provider:    ProviderGoogle,
		})
	}
	return events, nil
}

func (g *GoogleProvider) CreateEvent(ctx context.Context, e Event) (Event, error) {
	// Full implementation: POST /calendars/{id}/events
	// Stubbed — returns input with a generated ID.
	e.ID = fmt.Sprintf("gcal-%d", time.Now().UnixNano())
	e.Provider = ProviderGoogle
	return e, nil
}

func (g *GoogleProvider) UpdateEvent(ctx context.Context, e Event) error {
	// Full implementation: PUT /calendars/{id}/events/{eventId}
	return nil
}

func (g *GoogleProvider) DeleteEvent(ctx context.Context, id string) error {
	// Full implementation: DELETE /calendars/{id}/events/{eventId}
	return nil
}
