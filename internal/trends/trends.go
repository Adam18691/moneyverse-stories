package trends

import (
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

type TrendWindow struct {
	Region      string
	BestHourUTC int
	Audience    string
}

var GlobalTrendWindows = []TrendWindow{
	{"🇺🇸 أمريكا", 22, "ذروة 12-2 ظهراً + 7-10 مساءً — أعلى RPM"},
	{"🇸🇦 الخليج", 17, "ذروة بعد العشاء 9-11 مساءً"},
	{"🇹🇷 تركيا", 16, "ذروة بعد العشاء"},
	{"🇮🇳 الهند", 13, "ذروة 8-11 مساءً — أكبر جمهور"},
	{"🇮🇩 إندونيسيا", 10, "ذروة 7-10 مساءً"},
}

type Seasonal struct {
	Month string
	Event string
	Idea  string
	Power int
}

var SeasonalTrends = []Seasonal{
	{"يناير", "قرارات العام الجديد", "كيف بدأت الثروات يوم 1 يناير؟", 10},
	{"فبراير", "Tax Season أمريكا", "قصص تهرب ضريبي أسقطت مليونيرات", 7},
	{"مارس", "رمضان — زكاة وتجارة", "قصص تجارة الصحابة وزكاة المال", 10},
	{"يونيو", "تخرج ووظائف جديدة", "من أول راتب إلى أول مليون", 7},
	{"سبتمبر", "Back to Business", "خطط الأثرياء للربع الأخير", 8},
	{"نوفمبر", "Black Friday", "قصة اختراع Black Friday وأرباحه", 10},
	{"ديسمبر", "مراجعات العام", "أكبر صفقات أنهت العام بمليارات", 9},
}

// FetchDailyTrends: ترندات Google المجانية بدون API key
func FetchDailyTrends(country string) ([]string, error) {
	url := fmt.Sprintf("https://trends.google.com/trending/rss?geo=%s", country)
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var feed struct {
		Items []struct {
			Title string `xml:"title"`
		} `xml:"item"`
	}
	if err := xml.Unmarshal(body, &feed); err != nil {
		return nil, err
	}

	var out []string
	for _, it := range feed.Items {
		out = append(out, it.Title)
	}
	return out, nil
}

// MatchStoryToTrend: يربط الترند بقصة مالية
func MatchStoryToTrend(trend string) string {
	matches := map[string]string{
		"inflation": "قصة الدولار الذي فقد نصف قيمته.. ومن ربح من الانهيار",
		"crypto":    "قصة الشاب الذي راهن آخر 100$ على عملة رقمية",
		"ai":        "قصة أول مليونير بالذكاء الاصطناعي",
		"housing":   "قصة العقار: من مستأجر إلى مالك 50 عقاراً",
		"stocks":    "قصة السهم الذي صعد 10000% في سنة",
		"bank":      "قصة البنك الذي انهار.. ومن توقّع الأمر قبل الجميع",
		"economy":   "قصة اقتصاد انهار ونهض بقرار واحد",
		"money":     "قصة الرجل الذي ضاعف ثروته في أزمة الجميع",
	}
	low := strings.ToLower(trend)
	for k, story := range matches {
		if strings.Contains(low, k) {
			return story
		}
	}
	return ""
}

var _ = time.Now
