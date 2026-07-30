package signals

import "testing"

func TestExtractCoupleNames_BetweenPattern(t *testing.T) {
	cap := "sweet and intimate moments shared between Alida and Andrew during this engagement session"
	a, b, ok := ExtractCoupleNamesFromCaption(cap)
	if !ok {
		t.Fatal("expected names from between-pattern")
	}
	if a != "Alida" || b != "Andrew" {
		t.Fatalf("got %s & %s", a, b)
	}
}

func TestExtractCoupleNames_Capturing(t *testing.T) {
	cap := "Absolutely loved capturing these moments of Alissa and Jon at the Statehouse"
	a, b, ok := ExtractCoupleNamesFromCaption(cap)
	if !ok {
		t.Fatal("expected names")
	}
	if a != "Alissa" || b != "Jon" {
		t.Fatalf("got %s & %s", a, b)
	}
}

func TestExtractCoupleNames_RejectsNonNames(t *testing.T) {
	cap := "Light and airy photographer | Bright and airy | Natural photography"
	if _, _, ok := ExtractCoupleNamesFromCaption(cap); ok {
		t.Fatal("should not extract from marketing tags")
	}
}

func TestResolveCoupleFirstNames_CaptionWins(t *testing.T) {
	a, b := ResolveCoupleFirstNames(
		"moments shared between Alida and Andrew during this engagement",
		"jerseyshorepicnicco", "maddyroserestaurants",
		"", "",
		"jerseyshorepicnicco", "maddyroserestaurants",
	)
	if a.First != "Alida" || b.First != "Andrew" {
		t.Fatalf("got %s (%s) & %s (%s)", a.First, a.Source, b.First, b.Source)
	}
	if a.Source != "caption" {
		t.Fatalf("source=%s", a.Source)
	}
}

func TestFirstNameFromDisplay(t *testing.T) {
	n, ok := FirstNameFromDisplay("Alida Smith | NYC")
	if !ok || n != "Alida" {
		t.Fatalf("got %q ok=%v", n, ok)
	}
}

func TestParseDisplayName_FirstLast(t *testing.T) {
	f, l, ok := ParseDisplayName("Alida Smith | NYC")
	if !ok || f != "Alida" || l != "Smith" {
		t.Fatalf("got %q %q ok=%v", f, l, ok)
	}
}

func TestExtractCoupleFullNames(t *testing.T) {
	af, al, bf, bl, ok := ExtractCoupleFullNamesFromCaption(
		"capturing Alida Smith and Andrew Jones during their engagement session",
	)
	if !ok {
		t.Fatal("expected full names")
	}
	if af != "Alida" || al != "Smith" || bf != "Andrew" || bl != "Jones" {
		t.Fatalf("got %s %s / %s %s", af, al, bf, bl)
	}
}

func TestResolveCouple_MergesLastFromDisplay(t *testing.T) {
	a, b := ResolveCoupleFirstNames(
		"moments shared between Alida and Andrew during this engagement",
		"Alida Smith", "Andrew Jones",
		"", "",
		"alida_s", "ajones",
	)
	if a.First != "Alida" {
		t.Fatalf("a first %s", a.First)
	}
	if a.Last != "Smith" {
		t.Fatalf("expected last Smith from display, got %q", a.Last)
	}
	if b.First != "Andrew" || b.Last != "Jones" {
		t.Fatalf("b got %s %s", b.First, b.Last)
	}
}
