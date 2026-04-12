// Package theme classifies presentation subjects and generates deterministic
// design token systems for the dual render engine.
package theme

import (
	"strings"
)

// Mood represents a visual personality for a presentation deck.
type Mood string

const (
	MoodCorporate    Mood = "corporate"
	MoodCreative     Mood = "creative"
	MoodAcademic     Mood = "academic"
	MoodTechnology   Mood = "technology"
	MoodHealthcare   Mood = "healthcare"
	MoodFinance      Mood = "finance"
	MoodMarketing    Mood = "marketing"
	MoodSustainable  Mood = "sustainable"
	MoodStartup      Mood = "startup"
	MoodGovernment   Mood = "government"
)

// domainMoodMap maps domain keywords to a primary Mood.
var domainMoodMap = map[string]Mood{
	// Technology
	"tech":       MoodTechnology,
	"software":   MoodTechnology,
	"engineering": MoodTechnology,
	"ai":         MoodTechnology,
	"ml":         MoodTechnology,
	"data":       MoodTechnology,
	"cloud":      MoodTechnology,
	"saas":       MoodStartup,
	"devops":     MoodTechnology,
	"security":   MoodTechnology,
	"blockchain": MoodTechnology,

	// Finance
	"finance":    MoodFinance,
	"fintech":    MoodFinance,
	"banking":    MoodFinance,
	"investment": MoodFinance,
	"trading":    MoodFinance,
	"accounting": MoodFinance,
	"economy":    MoodFinance,

	// Healthcare
	"healthcare": MoodHealthcare,
	"medical":    MoodHealthcare,
	"pharma":     MoodHealthcare,
	"biotech":    MoodHealthcare,
	"clinical":   MoodHealthcare,
	"health":     MoodHealthcare,
	"hospital":   MoodHealthcare,

	// Academic
	"academic":   MoodAcademic,
	"research":   MoodAcademic,
	"education":  MoodAcademic,
	"university": MoodAcademic,
	"science":    MoodAcademic,
	"study":      MoodAcademic,

	// Marketing & Creative
	"marketing":  MoodMarketing,
	"brand":      MoodMarketing,
	"design":     MoodCreative,
	"creative":   MoodCreative,
	"media":      MoodCreative,
	"art":        MoodCreative,
	"advertising": MoodMarketing,

	// Startup
	"startup":    MoodStartup,
	"pitch":      MoodStartup,
	"venture":    MoodStartup,
	"founder":    MoodStartup,
	"growth":     MoodStartup,

	// Sustainable
	"sustainability": MoodSustainable,
	"environment": MoodSustainable,
	"climate":     MoodSustainable,
	"green":       MoodSustainable,
	"energy":      MoodSustainable,

	// Government
	"government": MoodGovernment,
	"policy":     MoodGovernment,
	"public":     MoodGovernment,
	"regulation": MoodGovernment,
	"compliance": MoodGovernment,
}

// ClassifyMood determines the Mood from domain string and topic tags.
// If no exact match is found, it falls back to MoodCorporate.
func ClassifyMood(domain string, topicTags []string) Mood {
	// Check domain first.
	if m, ok := lookupMood(domain); ok {
		return m
	}

	// Check topic tags.
	for _, tag := range topicTags {
		if m, ok := lookupMood(tag); ok {
			return m
		}
	}

	return MoodCorporate
}

func lookupMood(s string) (Mood, bool) {
	normalized := strings.ToLower(strings.TrimSpace(s))
	// Exact match.
	if m, ok := domainMoodMap[normalized]; ok {
		return m, true
	}
	// Substring match.
	for keyword, mood := range domainMoodMap {
		if strings.Contains(normalized, keyword) {
			return mood, true
		}
	}
	return "", false
}
