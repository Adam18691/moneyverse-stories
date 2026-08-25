package trends

import (
	"fmt"
	"time"
)

// TrendWindow: أفضل وقت نشر لكل منطقة زمنية
// القاعدة: انشر قبل ذروة مشاهدة المنطقة المستهدفة بـ 1-2 ساعة
type TrendWindow struct {
	Region    string
	BestHourUTC int // ساعة UTC للنشر
	Audience  string
}

// GlobalTrendWindows: مواعيد الذروة العالمية للنيتش المالي
var GlobalTrendWindows = []TrendWindow{
	{"🇸🇦 الخليج", 3, "ذروة بعد العشاء 9-11 مساءً"},       // UTC 18-21 → نشر UTC 17
	{"🇪🇬 مصر والمغرب", 20, "ذروة 8-11 مساءً"},
	{"🇹🇷 تركيا", 17, "ذروة بعد العشاء"},
	{"🇺🇸 أمريكا", 22, "ذروة 12-2 ظهراً + 7-10 مساءً"}, // أعلى RPM عالمياً 💰
	{"🇮🇳 الهند", 14, "ذروة 8-11 مساءً"},                // أكبر عدد مشاهدات
	{"🇮🇩 إندونيسيا", 11, "ذروة 7-10 مساءً"},
}

// SeasonalTrends: مواسم الترند المالي العالمي 🔥
type Seasonal struct {
	Month     string
	Event     string
	Idea      string
	Power     int // 1-10 قوة الترند
}

var SeasonalTrends = []Seasonal{
	{"يناير", "قرارات العام الجديد New Year Resolutions", "قصص: كيف بدأت الثروات يوم 1 يناير؟", 10},
	{"فبراير", "Tax Season أمريكا 🇺🇸", "قصص تهرب ضريبي أسقطت ملايينير", 7},
	{"مارس", "رمضان — زكاة وصدقة وتجارة", "قصص تجارة الصحابة وزكاة المال", 10},
	{"أبريل", "عيد الفطر + shopping season", "قصص تجار المواسم", 8},
	{"يونيو", "تخرج ووظائف جديدة", "من أول راتب إلى أول مليون", 7},
	{"سبتمبر", "Back to Business — ربع أخير", "خطط الأثرياء للربع الأخير", 8},
	{"نوفمبر", "Black Friday + Cyber Monday", "قصة اختراع Black Friday وأرباحه الخيالية", 10},
	{"ديسمبر", "Budget Season + مراجعات العام", "أكبر صفقات أنهت العام بمليارات", 9},
}

// DailyTrendHooks: ترندات يومية متجددة (تُحدَّث من Google Trends API مجاني)
// https://trends.google.com/trending/rss?geo=US — بدون مفتاح API!
func FetchDailyTrends(country string) ([]string, error) {
	url := fmt.Sprintf(
		"https://trends.google.com/trending/rss?geo=%s", country)
	// parse XML → استخراج <title> لكل ترند
	// ثم مطابقته مع بنك القصص المالي
	return fetchAndMatch(url)
}

// MatchStoryToTrend: يربط الترند اليومي بقصة من بنكك
func MatchStoryToTrend(trend string) string {
	matches := map[string]string{
		"inflation":  "قصة الدولار الذي فقد نصف قيمته.. ومن ربح من الانهيار",
		"crypto":     "قصة الشاب الذي راهن آخر 100$ على عملة رقمية",
		"ai":         "قصة أول مليونير بالذكاء الاصطناعي",
		"housing":    "قصة العقار: من مستأجر إلى مالك 50 عقاراً",
		"stocks":     "قصة السهم الذي صعد 10000% في سنة",
	}
	for key, story := range matches {
		if contains(trend, key) { return story }
	}
	return "" // لا تطابق → استخدم قصة دائمة من البنك
}

func contains(s, sub string) bool {
	return len(s) >= len(sub) && strings.Contains(strings.ToLower(s), sub)
}
