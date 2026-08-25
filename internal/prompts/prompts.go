package prompts

import (
	"fmt"
	"math/rand"
	"time"
)

var hooks = []string{
	"هذا الرجل خسر كل شيء في ليلة واحدة... ثم أصبح أغنى رجل في بلاده",
	"رفضوه 1009 مرات... وبنى إمبراطورية بقيمة 30 مليار دولار",
	"بدأ بعشرة دولارات فقط... اليوم شركته في كل بلاد العالم",
	"طُرد من الشركة التي أسسها بيده...",
	"قالوا له إن فكرته مجنونة... الآن الجميع يقلدونه",
	"أفلس ثلاث مرات... والمرة الرابعة غيّرت الاقتصاد العالمي",
}

var storyBases = []string{
	"قصة تاجر بدأ من الصفر ووصل إلى القمة",
	"قصة شركة كانت على وشك الإفلاس وأنقذها قرار واحد",
	"قصة رجل راهن بكل ما يملك على فكرة واحدة",
	"قصة امرأة بنبت إمبراطورية تجارية في زمن ما كان يسمح",
	"قصة صفقة واحدة غيرت مصير عائلة كاملة",
	"قصة اختراع بسيط حقق مليارات الدولارات",
}

var lessons = []string{
	"الصبر على التجارة قبل الربح منها",
	"لا تستثمر كل ما تملك في صفقة واحدة أبداً",
	"السمعة التجارية أهم من الربح السريع",
	"اقرأ السوق قبل أن تدخل فيه",
	"الفشل ليس النهاية بل درس مدفوع الأجر",
	"من يملك المهارة يملك المال أينما ذهب",
}

var anglesScenes = []string{
	"man standing at crossroads of a glowing city at night",
	"close-up of hands signing a life-changing contract",
	"empty wallet on rain-soaked street, dramatic lighting",
	"skyline office window overlooking financial district sunrise",
	"old ledger book with gold coins on wooden desk",
	"huge stock market board reflecting in determined eyes",
}

type MoneyPrompt struct {
	ID     int
	Hook   string
	Story  string
	Angles [10]string
	Title  string
	Tags   []string
	ImagePrompts []string
}

func Generate(id int) *MoneyPrompt {
	rnd := rand.New(rand.NewSource(time.Now().UnixNano() + int64(id)*7919))

	p := &MoneyPrompt{
		ID:    id,
		Hook:  hooks[rnd.Intn(len(hooks))],
		Story: fmt.Sprintf("%s. الدرس الذهبي: %s", pick(rnd, storyBases), pick(rnd, lessons)),
	}
	p.Angles = [10]string{
		fmt.Sprintf("Angle 1 البداية: %s", p.Story),
		"Angle 2 نقطة التحول: اللحظة التي غيّرت كل شيء",
		"Angle 3 القرار المصيري: المخاطرة الكبرى بالأرقام",
		"Angle 4 أول نجاح: أول ألف دولار وأول درس",
		"Angle 5 الأزمة الكبرى: الانهيار والخيانة والديون",
		"Angle 6 الصعود: الاستراتيجية الذكية خطوة بخطوة",
		fmt.Sprintf("Angle 7 الدروس الذهبية: %s", pick(rnd, lessons)),
		"Angle 8 الأرقام: إحصائيات مذهلة عن الثروة والنجاح",
		"Angle 9 الخاتمة الملهمة: لو فعل هو.. أنت تقدر أيضاً",
		"Angle 10 CTA: اشترك الآن + القصة التالية أغرب",
	}
	p.Title = fmt.Sprintf("💰 قصة ستغير نظرتك للمال | الحلقة %d", id)
	p.Tags = []string{"قصص نجاح", "المال", "الأعمال", "مليونير", "تجارة",
		"ثروة", "قصص واقعية", "تحفيز", "استثمار", "تطوير ذات"}

	for _, a := range anglesScenes {
		p.ImagePrompts = append(p.ImagePrompts,
			fmt.Sprintf("cinematic shot, %s, golden hour lighting, 8k, photorealistic, film grain --ar 16:9", a))
	}
	return p
}

func pick(rnd *rand.Rand, s []string) string { return s[rnd.Intn(len(s))] }
