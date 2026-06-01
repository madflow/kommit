package workflow

import "testing"

type fakeOutput struct {
	events []Event
}

func (o *fakeOutput) Emit(event Event) {
	o.events = append(o.events, event)
}

func assertEventKinds(t *testing.T, events []Event, want []EventKind) {
	t.Helper()

	if len(events) != len(want) {
		t.Fatalf("expected %d events, got %d", len(want), len(events))
	}

	for i, kind := range want {
		if events[i].Kind != kind {
			t.Fatalf("expected event %d to be %q, got %q", i, kind, events[i].Kind)
		}
	}
}
