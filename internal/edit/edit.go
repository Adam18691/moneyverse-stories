package edit

import (
	"fmt"

	"tayyibat-money/internal/hook"
)

type Cut struct {
	Start, End float64
	Type       string
	Effect     string
}

// BuildTimeline: مونتاج بقواعد retention العالمية
// cut كل 4 ثوانٍ + Ken Burns + Pattern Interrupt كل ~90 ثانية
func BuildTimeline(h *hook.Hook, totalDuration float64) []Cut {
	var tl []Cut

	tl = append(tl,
		Cut{0, 2, "hook_visual", h.VisualScene + " + zoom-in fast + shake"},
		Cut{2, 5, "hook_voice", h.VoiceLine + " + screen text GOLD 72pt"},
		Cut{5, 7, "hook_transition", "whoosh + flash to white → first scene"},
	)

	t := 7.0
	interrupt := 0
	for t < totalDuration {
		end := t + 4
		effect := "ken-burns-zoom-slow"
		if interrupt%23 == 22 {
			effect = "PATTERN-INTERRUPT: invert-color-flash + speed-ramp"
			fmt.Printf("   ⚡ Pattern Interrupt at %.0fs\n", t)
		}
		tl = append(tl, Cut{t, end, "scene_broll", effect})
		t = end
		interrupt++
	}
	return tl
}

var RetentionRules = []string{
	"✂️ Cut كل 3-5 ثواني",
	"🔍 Ken Burns zoom على كل صورة ثابتة",
	"🔊 Whoosh صوت عند كل انتقال",
	"⚡ Pattern interrupt كل 90 ثانية",
	"📊 أرقام تظهر مع ذكرها صوتياً",
	"🎯 CTA خفيف عند الدقيقة 3",
	"🎵 موسيقى BPM تصاعدي",
}
