package meal

import (
	"regexp"
	"strings"
	"time"
)

var (
	time24Pattern = regexp.MustCompile(`\b([01]?\d|2[0-3]):([0-5]\d)\b`)
	time12Pattern = regexp.MustCompile(`\b(1[0-2]|0?[1-9])\s*(am|pm)\b`)
)

func parseEatenAtFromText(raw string, now time.Time) (time.Time, bool) {
	text := strings.ToLower(strings.TrimSpace(raw))
	if text == "" {
		return time.Time{}, false
	}

	base := dateOnly(now)
	parsedSignal := false

	if strings.Contains(text, "yesterday") {
		base = base.AddDate(0, 0, -1)
		parsedSignal = true
	}
	if strings.Contains(text, "last night") {
		base = base.AddDate(0, 0, -1)
		parsedSignal = true
	}

	hour, minute := 12, 0
	if h, m, ok := parseClockFromText(text); ok {
		hour, minute = h, m
		parsedSignal = true
	} else if h, m, ok := parsePartOfDay(text); ok {
		hour, minute = h, m
		parsedSignal = true
	}

	if !parsedSignal {
		return time.Time{}, false
	}

	return time.Date(base.Year(), base.Month(), base.Day(), hour, minute, 0, 0, time.UTC), true
}

func parseClockFromText(text string) (int, int, bool) {
	if m := time24Pattern.FindStringSubmatch(text); len(m) == 3 {
		parsed, err := time.Parse("15:04", m[1]+":"+m[2])
		if err == nil {
			return parsed.Hour(), parsed.Minute(), true
		}
	}

	if m := time12Pattern.FindStringSubmatch(text); len(m) == 3 {
		parsed, err := time.Parse("3pm", strings.ReplaceAll(m[1]+m[2], " ", ""))
		if err == nil {
			return parsed.Hour(), parsed.Minute(), true
		}
	}

	return 0, 0, false
}

func parsePartOfDay(text string) (int, int, bool) {
	switch {
	case strings.Contains(text, "this morning"), strings.Contains(text, "morning"):
		return 8, 0, true
	case strings.Contains(text, "afternoon"):
		return 14, 0, true
	case strings.Contains(text, "this evening"), strings.Contains(text, "evening"):
		return 19, 0, true
	case strings.Contains(text, "tonight"):
		return 20, 0, true
	case strings.Contains(text, "night"):
		return 21, 0, true
	default:
		return 0, 0, false
	}
}
