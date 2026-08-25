package hook

import (
	"fmt"
	"math/rand"
	"time"
)

// HookEngine: يبني هوك 7 ثواني بتركيبة نفسية مثالية
type Hook struct {
	VisualScene   string // برومبت المشهد البصري
	ScreenText    string // الكلمة الضخمة على الشاشة
	VoiceLine     string // جملة الراوي
	CuriosityLine string // "شاهد حتى النهاية"
	Duration      float64 // دايماً 7.0
}

var shockNumbers = []string{
	"2 مليون دولار", "48 ساعة", "1009 رفض", "3 إفلاسات",
	"10 دولارات", "900 مليون", "صفقة واحدة",
}

var paradoxLines = []string{
	"أفلس تماماً.. وهذا بالضبط ما صنعه مليونيراً",
	"طُرد من عمله.. فأسس إمبراطورية أكبر من الشركة نفسها",
	"قالوا له مستحيل.. والآن الجميع يقلّد فكرته",
	"خسر كل شيء في ليلة واحدة.. وبعد 3 سنوات اشترى المبنى الذي طُرد منه",
}

func Generate(seed int) *Hook {
	rnd := rand.New(rand.NewSource(time.Now().UnixNano() + int64(seed)*104729))
	num := shockNumbers[rnd.Intn(len(shockNumbers))]
	return &Hook{
		Duration: 7.0,
		VisualScene: fmt.Sprintf(
			"cinematic extreme close-up, burning banknotes and falling stock charts, "+
				"dark moody lighting, gold particles floating, 8k photorealistic, film grain --ar 16:9"),
		ScreenText:    num,
		VoiceLine:     fmt.Sprintf("خسر %s.. %s", num, pick(rnd, paradoxLines)),
		CuriosityLine: "والنهاية ستصدقك.. شاهد حتى الآخر",
	}
}

func pick(rnd *rand.Rand, s []string) string { return s[rnd.Intn(len(s))] }
