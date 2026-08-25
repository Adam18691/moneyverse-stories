package edit

import (
	"fmt"

	"tayyibat-money/internal/hook"
)

// Timeline: خط زمني سينمائي كامل للفيديو
type Cut struct {
	Start, End float64
	Type       string // hook / scene / broll / text / transition
	Effect     string
}

// BuildTimeline: المونتاج بقواعد الـ retention العالمية
//
// أسرار المونتاج المخفية المستخدمة هنا:
// 1. Cut كل 3-5 ثواني = الدماغ لا يمل (قاعدة MrBeast)
// 2. Zoom-in بطيء دائم (Ken Burns) = حياة في الصورة الثابتة
// 3. Whoosh sound عند كل cut = إحساس بالحركة
// 4. B-Roll فوق كل جملة = مشاهدات لا تنزل
// 5. Pattern Interrupt كل 90 ثانية = منع الملل (تغيير لون/زاوية)
func BuildTimeline(h *hook.Hook, totalDuration float64) []Cut {
	var tl []Cut

	// 🎯 HOOK — أول 7 ثواني (الأهم في الفيديو كله)
	tl = append(tl,
		Cut{0, 2, "hook_visual", "burning-money-scene + zoom-in fast + shake"},
		Cut{2, 5, "hook_voice", h.VoiceLine + " + screen text GOLD 72pt"},
		Cut{5, 7, "hook_transition", "whoosh + flash to white → first scene"},
	)

	// 🎬 الجسم — cut كل 4 ثواني، pattern interrupt كل 90 ثانية
	t := 7.0
	interrupt := 0
	for t < totalDuration {
		end := t + 4
		effect := "ken-burns-zoom-slow"
		if interrupt%23 == 22 { // كل ~90 ثانية
			effect = "PATTERN-INTERRUPT: invert-color-flash + speed-ramp"
			fmt.Printf("   ⚡ Pattern Interrupt at %.0fs\n", t)
		}
		tl = append(tl, Cut{t, end, "scene_broll", effect})
		t = end
		interrupt++
	}
	return tl
}

// RetentionRules: ملخص القواعد الذهبية
var RetentionRules = []string{
	"✂️ Cut كل 3-5 ثواني — لا مشهد ثابت أبداً",
	"🔍 Ken Burns zoom على كل صورة ثابتة",
	"🔊 Whoosh/Swoosh صوت عند كل انتقال",
	"⚡ Pattern interrupt كل 90 ثانية (وميض/تسريع)",
	"📊 أرقام تظهر على الشاشة مع ذكرها صوتياً",
	"🎯 CTA نصي خفيف عند الدقيقة 3 وليس النهاية فقط",
	"🎵 موسيقى تصاعدية BPM يزيد مع القصة",
}
