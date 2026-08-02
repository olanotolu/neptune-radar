package connectors

import "testing"

func TestParseHTMLParishesTable(t *testing.T) {
	html := []byte(`<table><tr><td><a href="https://stmary.org">St. Mary Church</a></td><td>123 Main St, Anytown</td></tr>
<tr><td><a href="/parishes/stjohn">St. John the Baptist</a></td><td>45 Oak Ave</td></tr>
<tr><td><a href="/about">About Us</a></td><td>no address</td></tr></table>`)
	got := parseHTMLParishes(html, "https://diocese.org/parishes")
	if len(got) != 2 {
		t.Fatalf("want 2 parishes, got %d: %+v", len(got), got)
	}
	if got[0].Name != "St. Mary Church" || got[0].WebsiteURL != "https://stmary.org" {
		t.Errorf("row0: %+v", got[0])
	}
	if got[0].Address == "" || !contains(got[0].Address, "Main St") { // contains defined in marriage_scraper_test.go
		t.Errorf("row0 address: %q", got[0].Address)
	}
	if got[1].WebsiteURL != "https://diocese.org/parishes/stjohn" {
		t.Errorf("row1 relative link not resolved: %q", got[1].WebsiteURL)
	}
}

func TestParseJSONParishes(t *testing.T) {
	body := []byte(`{"parishes":[{"name":"St. Mary","website":"https://stmary.org","address":"123 Main St"},{"name":"Holy Cross","websiteUrl":"https://holycross.org"}]}`)
	got, ok := parseJSONParishes(body, "https://diocese.org/directory")
	if !ok {
		t.Fatal("expected JSON parse ok")
	}
	if len(got) != 2 {
		t.Fatalf("want 2, got %d: %+v", len(got), got)
	}
	if got[0].Name != "St. Mary" || got[0].WebsiteURL != "https://stmary.org" {
		t.Errorf("rec0: %+v", got[0])
	}
	if got[1].Name != "Holy Cross" || got[1].WebsiteURL != "https://holycross.org" {
		t.Errorf("rec1: %+v", got[1])
	}
}

func TestDedupByName(t *testing.T) {
	in := []DiscoveredParish{
		{Name: "St. Mary", WebsiteURL: "https://stmary.org"},
		{Name: "St. Mary", WebsiteURL: "https://stmary.org"},
		{Name: "St. John", WebsiteURL: "https://stjohn.org"},
	}
	got := dedupByName(in)
	if len(got) != 2 {
		t.Fatalf("want 2 after dedup, got %d", len(got))
	}
}
