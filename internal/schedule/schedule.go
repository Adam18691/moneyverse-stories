package schedule

import (
	"fmt"
	"sort"
	"time"

	"tayyibat-money/internal/trends"
)

type Region struct {
	Code        string
	Name        string
	PeakHourUTC int
	RPM         float64
}

// BuildRegions: كل الدول مرتبة حسب RPM — الأعلى أولاً 💰
func BuildRegions() []Region {
	var out []Region
	for _, c := range trends.AllCountries() {
		out = append(out, Region{
			Code:        c.Code,
			Name:        c.Name,
			PeakHourUTC: c.PeakUTC,
			RPM:         c.RPM,
		})
	}
	sort.Slice(out, func(i, j int) bool {
		return out[i].RPM > out[j].RPM
	})
	return out
}

// PlanDay: 4 فيديوهات → أعلى 4 أسواق RPM تلقائياً
func PlanDay(videoIDs []int, now time.Time, regions []Region) map[int]time.Time {
	plan := map[int]time.Time{}
	for i, id := range videoIDs {
		r := regions[i%len(regions)]
		target := time.Date(now.Year(), now.Month(), now.Day(),
			r.PeakHourUTC-2, 0, 0, 0, time.UTC)
		if !target.After(now) {
			target = target.Add(24 * time.Hour)
		}
		plan[id] = target
		fmt.Printf("📅 VIDEO %d → %s (RPM $%.0f) at %s UTC\n",
			id, r.Name, r.RPM, target.Format("15:04"))
	}
	return plan
}
