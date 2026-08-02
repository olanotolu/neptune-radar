package store

import "testing"

func TestLessonForSuppress(t *testing.T) {
	if lessonForSuppress("vendor_vendor_pair", 0.5) == "" {
		t.Fatal("empty lesson")
	}
	hi := lessonForSuppress("not_a_couple", 0.95)
	if hi == "" {
		t.Fatal("empty high score lesson")
	}
}

func TestLessonForIgnore(t *testing.T) {
	if lessonForIgnore("review", 0.95) == "" {
		t.Fatal("empty")
	}
	if lessonForIgnore("concierge_review", 0.6) == "" {
		t.Fatal("empty risk")
	}
}
