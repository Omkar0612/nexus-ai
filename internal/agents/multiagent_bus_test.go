package agents

import (
	"context"
	"testing"
	"time"
)

func TestBusRegisterAndSend(t *testing.T) {
	bus := NewBus(5 * time.Second)
	err := bus.Register(&SubAgent{
		Role: RoleResearcher,
		Name: "Test Researcher",
		Handler: func(ctx context.Context, msg BusMessage) (BusMessage, error) {
			return BusMessage{Type: MsgResult, Payload: "research result: " + msg.Payload}, nil
		},
	})
	if err != nil {
		t.Fatalf("Register: %v", err)
	}

	result, err := bus.Send(context.Background(), BusMessage{
		Type: MsgTask, From: RoleOrchestrator, To: RoleResearcher,
		Payload: "what is Go?",
	})
	if err != nil {
		t.Fatalf("Send: %v", err)
	}
	if result.Type != MsgResult {
		t.Errorf("expected MsgResult, got %s", result.Type)
	}
}

func TestBusUnknownRole(t *testing.T) {
	bus := NewBus(5 * time.Second)
	_, err := bus.Send(context.Background(), BusMessage{
		Type: MsgTask, From: RoleOrchestrator, To: RoleCoder, Payload: "fix this",
	})
	if err == nil {
		t.Error("expected error for unregistered role")
	}
}

func TestBusInferRole(t *testing.T) {
	bus := NewBus(5 * time.Second)
	cases := []struct {
		task string
		want AgentRole
	}{
		{"search for Go tutorials", RoleResearcher},
		{"write a function to sort", RoleCoder},
		{"write a summary of this", RoleWriter},
		{"analyse the data trends", RoleAnalyst},
		{"review this code", RoleReviewer},
	}
	for _, c := range cases {
		got := bus.inferRole(c.task)
		if got != c.want {
			t.Errorf("inferRole(%q) = %s, want %s", c.task, got, c.want)
		}
	}
}

func TestBusBroadcast(t *testing.T) {
	bus := NewBus(5 * time.Second)
	for _, role := range []AgentRole{RoleResearcher, RoleWriter} {
		r := role
		_ = bus.Register(&SubAgent{
			Role: r, Name: string(r),
			Handler: func(ctx context.Context, msg BusMessage) (BusMessage, error) {
				return BusMessage{Type: MsgResult, Payload: string(r) + " done"}, nil
			},
		})
	}
	results := bus.Broadcast(context.Background(), "process this task")
	if len(results) != 2 {
		t.Errorf("expected 2 broadcast results, got %d", len(results))
	}
}

func TestBusStats(t *testing.T) {
	bus := NewBus(5 * time.Second)
	_ = bus.Register(&SubAgent{
		Role: RoleAnalyst, Name: "Analyst",
		Handler: func(ctx context.Context, msg BusMessage) (BusMessage, error) {
			return BusMessage{Type: MsgResult, Payload: "analysis done"}, nil
		},
	})
	bus.Send(context.Background(), BusMessage{Type: MsgTask, From: RoleOrchestrator, To: RoleAnalyst, Payload: "analyse"})
	stats := bus.Stats()
	if stats.TotalTasks < 1 {
		t.Error("expected at least 1 task in stats")
	}
}
