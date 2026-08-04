package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"neptune-social-radar/backend/internal/llm"
	"neptune-social-radar/backend/internal/mail"
	"neptune-social-radar/backend/internal/outreach"
	"neptune-social-radar/backend/internal/records"
	"neptune-social-radar/backend/internal/store"
)

func main() {
	kitID := "kit_9a33602a-536a-4e85-afad-08bef002a112"
	if len(os.Args) > 1 {
		kitID = os.Args[1]
	}
	dsn := os.Getenv("DATABASE_URL")
	s, err := store.Open(dsn)
	if err != nil {
		log.Fatal(err)
	}
	defer s.Close()

	k, _ := s.GetCongratulateKit(kitID)
	fmt.Printf("KIT %s A=%s %s B=%s %s city=%s %s\n", k.ID, k.FirstNameA, k.LastNameA, k.FirstNameB, k.LastNameB, k.AddressCity, k.AddressRegion)
	fmt.Println("Apify:", records.ApifyTPSStatus())

	agent := &outreach.Agent{
		Store: s, LLM: llm.NewInterpreter(),
		Records: records.NewMulti(), Mail: mail.NewFromEnv(),
	}
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Minute)
	defer cancel()
	start := time.Now()
	out, err := agent.RunDetective(ctx, kitID)
	fmt.Printf("elapsed=%s err=%v\n", time.Since(start).Round(time.Second), err)
	fmt.Printf("source=%s conf=%.2f city=%s cands=%d\n", out.AddressSource, out.AddressConfidence, out.AddressCity, len(out.AddressCandidates))
	streetN := 0
	for i, c := range out.AddressCandidates {
		isSt := c.Line1 != "" && !strings.HasPrefix(strings.ToLower(c.Line1), "http") && c.Kind != "research_link"
		if isSt {
			streetN++
		}
		fmt.Printf("[%d] kind=%s street=%q city=%s %s %s conf=%.0f%% src=%s\n",
			i, c.Kind, c.Line1, c.City, c.Region, c.Postal, c.Confidence*100, c.Source)
		if c.Note != "" && len(c.Note) < 160 {
			fmt.Printf("     %s\n", c.Note)
		} else if c.Note != "" {
			fmt.Printf("     %s…\n", c.Note[:160])
		}
	}
	fmt.Printf("STREETS=%d\n", streetN)
	// tail research notes
	n := out.ResearchNotes
	if len(n) > 1500 {
		n = n[len(n)-1500:]
	}
	fmt.Println("--- notes tail ---\n", n)
	b, _ := json.MarshalIndent(map[string]any{"streets": streetN, "status": out.Status}, "", "  ")
	fmt.Println(string(b))
	if streetN == 0 {
		os.Exit(2)
	}
}
