package store

import "testing"

func TestShouldAdvanceJourney(t *testing.T) {
	cases := []struct {
		cur, target string
		want        bool
	}{
		{"detected", "in_chat", true},
		{"in_chat", "booked", true},
		{"booked", "in_chat", false},
		{"closed_won", "closed_lost", false},
		{"invited", "closed_lost", true},
		{"booked", "closed_won", true},
		{"detected", "invited", true},
	}
	for _, c := range cases {
		got := shouldAdvanceJourney(c.cur, c.target)
		if got != c.want {
			t.Errorf("shouldAdvance(%q→%q)=%v want %v", c.cur, c.target, got, c.want)
		}
	}
}

func TestEventToJourneyMap(t *testing.T) {
	for _, ev := range []string{"chat_started", "consult_booked", "closed_won", "closed_lost", "handoff_clicked"} {
		if eventToJourney[ev] == "" {
			t.Errorf("missing mapping for %s", ev)
		}
	}
}
