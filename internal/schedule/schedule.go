package schedule

import (
	"fmt"
	"time"
)

// Region: منطقة نشر مع وقت الذروة
type Region struct {
	Name       string        // 🇸🇦 الخليج
	PeakHourUTC int          // ساعة الذروة UTC
	RPM        float64       // قيمة RPM التقديرية
	FlagEmoji  string
}

// Regions: مرتبة حسب الأولوية (RPM × حجم الجمهور)
var Regions = []Region{
	{"أمريكا 🇺🇸", 22, 25.0, "🇺🇸"},   // أعلى RPM عالمياً 💰
	{"الخليج 🇸🇦", 17, 8.0, "🇸🇦"},
	{"تركيا 🇹🇷", 16, 4.0, "🇹🇷"},
	{"الهند 🇮🇳", 13, 2.0, "🇮🇳"},     // أكبر جمهور
	{"إندونيسيا 🇮🇩", 10, 1.5, "🇮🇩"},
}

// NextPublishTime: يحسب أقرب موعد ذروة قادم لأي منطقة
func NextPublishTime(r Region, from time.Time) time.Time {
	target := time.Date(from.Year(), from.Month(), from.Day(),
		r.PeakHourUTC-2, 0, 0, 0, time.UTC) // نشر قبل الذروة بساعتين (معالجة + ترند)
	if !target.After(from) {
		target = target.Add(24 * time.Hour) // ذروة اليوم انتهت → الغد
	}
	return target
}

// PlanDay: يوزع الفيديوهات الأربعة على أفضل 4 نوافذ يومية
//
// الاستراتيجية: كل فيديو لنافذة مختلفة = تغطية 24 ساعة
// الفيديو الأفضل جودة (hook الأقوى) يروح لأعلى RPM 🇺🇸
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

// FormatRFC3339: الصيغة المطلوبة من YouTube API
func FormatRFC3339(t time.Time) string { return t.UTC().Format(time.RFC3339) }
