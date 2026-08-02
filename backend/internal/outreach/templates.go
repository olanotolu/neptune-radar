package outreach

import (
	"fmt"
	"strings"
)

// GreetingTemplate is a curated postcard greeting style.
type GreetingTemplate struct {
	ID               string `json:"id"`
	Name             string `json:"name"`
	Tone             string `json:"tone"` // warm, formal, casual
	CustomerFacing   string `json:"customer_facing"`
	InternalNote     string `json:"internal_note"`
	ConfidenceFloor  float64 `json:"confidence_floor"` // min address confidence for this template
}

// TemplateLibrary returns the curated greeting templates.
func TemplateLibrary() []GreetingTemplate {
	return []GreetingTemplate{
		{
			ID:   "warm_ohio",
			Name: "Warm Ohio",
			Tone: "warm",
			CustomerFacing: `Dear {{.NameA}} & {{.NameB}},

Congratulations on your engagement! {{if .Location}}We're so happy for you both here in {{.Location}}.{{else}}What an exciting season!{{end}}

May this time of planning bring you closer together, surrounded by the people who love you most.

With warm regards,
Neptune`,
			InternalNote:    "Warm, location-aware greeting. Best for verified Ohio addresses.",
			ConfidenceFloor: 0.65,
		},
		{
			ID:   "bright_casual",
			Name: "Bright & Casual",
			Tone: "casual",
			CustomerFacing: `Dear {{.NameA}} & {{.NameB}},

The news made our day — congratulations! {{if .Location}}Cannot wait to see what you plan for {{.Location}}.{{else}}This is going to be amazing.{{end}}

Here's to the next chapter. 🥂

With care,
Neptune`,
			InternalNote:    "Bright, upbeat tone. Good for social-media-savvy couples.",
			ConfidenceFloor: 0.50,
		},
		{
			ID:   "elegant_formal",
			Name: "Elegant & Formal",
			Tone: "formal",
			CustomerFacing: `Dear {{.NameA}} & {{.NameB}},

It is with great pleasure that we extend our warmest congratulations on your engagement. {{if .Location}}What a beautiful journey lies ahead for you in {{.Location}}.{{end}}

May your wedding planning be filled with joy, elegance, and the love of family and friends.

Warmly,
Neptune`,
			InternalNote:    "Formal, classic tone. Best for高端 couples or when address source is premium.",
			ConfidenceFloor: 0.75,
		},
	}
}

// TemplateData is the data available for template rendering.
type TemplateData struct {
	NameA    string
	NameB    string
	Location string
}

// RenderTemplate renders a greeting template with the given data.
func RenderTemplate(tpl GreetingTemplate, data TemplateData) string {
	result := tpl.CustomerFacing
	result = strings.ReplaceAll(result, "{{.NameA}}", data.NameA)
	result = strings.ReplaceAll(result, "{{.NameB}}", data.NameB)
	if data.Location != "" {
		result = strings.ReplaceAll(result, "{{if .Location}}We're so happy for you both here in {{.Location}}.{{else}}What an exciting season!{{end}}",
			fmt.Sprintf("We're so happy for you both here in %s.", data.Location))
		result = strings.ReplaceAll(result, "{{if .Location}}Cannot wait to see what you plan for {{.Location}}.{{else}}This is going to be amazing.{{end}}",
			fmt.Sprintf("Cannot wait to see what you plan for %s.", data.Location))
		result = strings.ReplaceAll(result, "{{if .Location}}What a beautiful journey lies ahead for you in {{.Location}}.{{else}}{{end}}",
			fmt.Sprintf("What a beautiful journey lies ahead for you in %s.", data.Location))
	} else {
		result = strings.ReplaceAll(result, "{{if .Location}}We're so happy for you both here in {{.Location}}.{{else}}What an exciting season!{{end}}",
			"What an exciting season!")
		result = strings.ReplaceAll(result, "{{if .Location}}Cannot wait to see what you plan for {{.Location}}.{{else}}This is going to be amazing.{{end}}",
			"This is going to be amazing.")
		result = strings.ReplaceAll(result, "{{if .Location}}What a beautiful journey lies ahead for you in {{.Location}}.{{else}}{{end}}",
			"")
	}
	return strings.TrimSpace(result)
}
