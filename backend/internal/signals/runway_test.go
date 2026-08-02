package signals

import (
	"testing"
	"time"
)

func TestExtractWeddingRunway_MonthDayYear(t *testing.T) {
	now := time.Date(2026, 3, 1, 0, 0, 0, 0, time.UTC)
	r := ExtractWeddingRunway("We're getting married October 12, 2026 💍", "", "", nil, now)
	if r.Date == nil {
		t.Fatal("expected date")
	}
	if r.Date.Month() != time.October || r.Date.Day() != 12 || r.Date.Year() != 2026 {
		t.Fatalf("got %v", r.Date)
	}
	if r.Band != "green" {
		t.Fatalf("band=%s want green", r.Band)
	}
	if r.SuppressOutreach {
		t.Fatal("should not suppress long runway")
	}
	if r.Factor < 0.9 {
		t.Fatalf("factor=%v", r.Factor)
	}
}

func TestExtractWeddingRunway_TooSoon(t *testing.T) {
	now := time.Date(2026, 3, 1, 0, 0, 0, 0, time.UTC)
	r := ExtractWeddingRunway("wedding on March 20, 2026!!!", "", "", nil, now)
	if r.DaysUntil == nil || *r.DaysUntil > 30 {
		t.Fatalf("days=%v", r.DaysUntil)
	}
	if r.Band != "red" {
		t.Fatalf("band=%s", r.Band)
	}
	if !r.SuppressOutreach {
		t.Fatal("expected suppress")
	}
}

func TestExtractWeddingRunway_Numeric(t *testing.T) {
	now := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	r := ExtractWeddingRunway("save the date 09/15/2026", "", "", nil, now)
	if r.Date == nil || r.Date.Month() != time.September {
		t.Fatalf("got %+v", r)
	}
}

func TestExtractWeddingRunway_Unknown(t *testing.T) {
	r := ExtractWeddingRunway("just engaged today", "love my fiancé", "", nil, time.Time{})
	if r.Band != "unknown" {
		t.Fatalf("band=%s", r.Band)
	}
	if r.Factor != 0.75 {
		t.Fatalf("factor=%v", r.Factor)
	}
}

func TestExtractWeddingRunway_HashtagYear(t *testing.T) {
	now := time.Date(2026, 2, 1, 0, 0, 0, 0, time.UTC)
	r := ExtractWeddingRunway("yay", "", "", []string{"2027bride", "justengaged"}, now)
	if r.Date == nil || r.Date.Year() != 2027 {
		t.Fatalf("got %+v", r)
	}
}

func TestFormatRunwayLabel(t *testing.T) {
	d := 90
	r := WeddingRunway{DaysUntil: &d, Band: "amber"}
	if got := FormatRunwayLabel(r); got == "" {
		t.Fatal("empty label")
	}
}
