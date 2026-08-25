package schedule

import (
	"fmt"
	"time"
)

type Region struct {
	Name        string
	PeakHourUTC int
	RPM         float64
	FlagEmoji   string
}

var Regions = []Region{
	{"أمريكا 🇺🇸", 22, 25.0, "🇺🇸"},
	{"الخليج 🇸🇦", 17, 8.0, "🇸🇦"},
	{"تركيا 🇹🇷", 16, 4.0, "🇹🇷"},
	{"الهند 🇮🇳", 13, 2.0, "🇮🇳"},
	{"إندونيسيا 🇮🇩", 10, 1.5, "🇮🇩"},
}

func NextPublishTime(r Region, from time.Time) time.Time {
	target := time.Date(from.Year(), from.Month(), from.Day(),
		r.PeakHourUTC-2, 0, 0, 0, time.UTC)
	if !target.After(from) {
		target = target.Add(24 * time.Hour)
	}
	return target
}

// PlanDay: يوزع الفيديوهات على أفضل النوافذ (كل منطقة مرة)
func PlanDay(videoIDs []int, now time.Time) map[int]time.Time {
	plan := map[int]time.Time{}
	used := map[string]bool{}

	for _, id := range videoIDs {
		for _, r := range Regions {
			t := NextPublishTime(r, now)
			key := r.Name + t.Format("2006-01-02")
			if !used[key] {
				used[key] = true
				plan[id] = t
				fmt.Printf("📅 VIDEO %d → %s at %s (peak %d UTC)\n",
					id, r.FlagEmoji+r.Name, t.Format("15:04"), r.PeakHourUTC)
				break
			}
		}
	}
	return plan
}

func FormatRFC3339(t time.Time) string { return t.UTC().Format(time.RFC3339) }
