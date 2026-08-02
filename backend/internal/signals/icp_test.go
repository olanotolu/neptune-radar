package signals

import "testing"

func TestExtractICPFit_TechEquity(t *testing.T) {
	icp := ExtractICPFit(
		"SWE @ Google · RSUs 📈",
		"PM at Meta",
		"San Francisco",
		"CA",
	)
	if icp.Score < 0.5 {
		t.Fatalf("score too low: %v tags=%v", icp.Score, icp.Tags)
	}
	if len(icp.Employers) < 1 {
		t.Fatalf("expected employers, got %v", icp.Employers)
	}
	if icp.MarketPriority < 0.9 {
		t.Fatalf("sf should be priority, got %v (%s)", icp.MarketPriority, icp.MarketLabel)
	}
	found := false
	for _, tag := range icp.Tags {
		if tag == "dual_professional" {
			found = true
		}
	}
	if !found {
		t.Fatalf("expected dual_professional, tags=%v", icp.Tags)
	}
}

func TestExtractICPFit_PhysicianNYC(t *testing.T) {
	icp := ExtractICPFit("MD | cardiology fellow", "attorney in NYC", "New York", "NY")
	if icp.Score < 0.45 {
		t.Fatalf("score=%v", icp.Score)
	}
	if icp.MarketPriority < 0.9 {
		t.Fatalf("nyc market=%v", icp.MarketPriority)
	}
}

func TestExtractICPFit_UnknownBaseline(t *testing.T) {
	icp := ExtractICPFit("dog mom 🐶", "travel ✈️", "", "")
	if icp.Score > 0.35 {
		t.Fatalf("unexpected high score %v for weak bios", icp.Score)
	}
}

func TestNeptuneRank(t *testing.T) {
	// Strong engagement + ICP + runway
	hi := NeptuneRank(0.95, 0.8, 1.0, 1.0)
	lo := NeptuneRank(0.95, 0.2, 0.25, 0.5) // red runway
	if hi <= lo {
		t.Fatalf("hi=%v should beat lo=%v", hi, lo)
	}
	if NeptuneRank(0.9, 0.9, 0, 1) > 0.05 {
		t.Fatal("zero runway should crush rank")
	}
}
