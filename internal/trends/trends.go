package trends

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"strings"

	"tayyibat-money/internal/ai"
)

// ══════════════════════════════════════════
// 🌍 البيانات العالمية — 6 قارات / 28 دولة
// ══════════════════════════════════════════

type Country struct {
	Code    string  // ISO code لـ Google Trends
	Lang    string
	Name    string // اسم مع علم emoji
	RPM     float64
	PeakUTC int // ذروة المشاهدة الافتراضية UTC
}

type Continent struct {
	Name      string
	Countries []Country
}

var World = []Continent{
	{"🌎 أمريكا الشمالية", []Country{
		{"US", "en", "🇺🇸 أمريكا", 25.0, 22},
		{"CA", "en", "🇨🇦 كندا", 18.0, 22},
		{"MX", "es", "🇲🇽 المكسيك", 3.0, 23},
	}},
	{"🌵 أمريكا الجنوبية", []Country{
		{"BR", "pt", "🇧🇷 البرازيل", 2.5, 21},
		{"AR", "es", "🇦🇷 الأرجنتين", 2.0, 23},
		{"CO", "es", "🇨🇴 كولومبيا", 1.5, 0},
	}},
	{"🕌 آسيا", []Country{
		{"SA", "ar", "🇸🇦 السعودية", 8.0, 17},
		{"AE", "ar", "🇦🇪 الإمارات", 10.0, 16},
		{"EG", "ar", "🇪🇬 مصر", 3.0, 18},
		{"TR", "tr", "🇹🇷 تركيا", 4.0, 16},
		{"IN", "hi", "🇮🇳 الهند", .0, 13},
		{"ID", "id", "🇮🇩 إندونيسيا", 1.5, 10},
		{"PK", "ur", "🇵🇰 باكستان", 1.5, 14},
		{"MY", "ms", "🇲🇾 ماليزيا", 3.0, 11},
		{"PH", "tl", "🇵🇭 الفلبين", 2.0, 9},
		{"JP", "ja", "🇯🇵 اليابان", 12.0, 12},
	}},
	{"🦁 أفريقيا", []Country{
		{"NG", "en", "🇳🇬 نيجيريا", 2.0, 19},
		{"ZA", "en", "🇿🇦 جنوب أفريقيا", 4.0, 18},
		{"KE", "sw", "🇰🇪 كينيا", 1.5, 17},
		{"MA", "ar", "🇲🇦 المغرب", 2.0, 20},
	}},
	{"🏰 أوروبا", []Country{
		{"GB", "en", "🇬🇧 بريطانيا", 15.0, 20},
		{"DE", "de", "🇩🇪 ألمانيا", 15.0, 19},
		{"FR", "fr", "🇫🇷 فرنسا", 10.0, 19},
		{"ES", "es", "🇪🇸 إسبانيا", 7.0, 20},
		{"IT", "it", "🇮🇹 إيطاليا", 7.0, 19},
		{"NL", "nl", "🇳🇱 هولندا", 14.0, 20},
	}},
	{"🐨 أوقيانوسيا", []Country{
		{"AU", "en", "🇦🇺 أستراليا", 17.0, 12},
		{"NZ", "en", "🇳🇿 نيوزيلندا", 12.0, 9},
	}},
}

// AllCountries: دمج كل الدول (~28 دولة)
func AllCountries() []Country {
	var out []Country
	for _, cont := range World {
		out = append(out, cont.Countries...)
	}
	return out
}

// PrintCoverage: إحصائيات التغطية
func PrintCoverage() {
	fmt.Printf("🌍 WORLDWIDE COVERAGE: %d countries | %d continents\n",
		len(AllCountries()), len(World))
}

// ══════════════════════════════════════════
// 🔥 جلب الترندات الحقيقية — Google Trends RSS مجاني
// ══════════════════════════════════════════

func FetchDailyTrends(country string) ([]string, error) {
	url := fmt.Sprintf("https://trends.google.com/trending/rss?geo=%s", country)
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)

	var feed struct {
		Items []struct {
			Title string `xml:"title"`
		} `xml:"item"`
	}
	if xml.Unmarshal(body, &feed) != nil {
		return nil, fmt.Errorf("parse فشل")
	}
	var out []string
	for _, it := range feed.Items {
		out = append(out, it.Title)
	}
	return out, nil
}

// FetchWorldTrends: جلب متوازي لكل الدول ⚡
func FetchWorldTrends() map[string][]string {
	worldTrends := map[string][]string{}
	all := AllCountries()

	type result struct {
		Code   string
		Trends []string
	}
	results := make(chan result, len(all))

	for _, c := range all {
		go func(code string) {
			t, err := FetchDailyTrends(code)
			if err != nil {
				t = nil
			}
			results <- result{code, t}
		}(c.Code)
	}

	for i := 0; i < len(all); i++ {
		r := <-results
		if r.Trends != nil && len(r.Trends) > 0 {
			worldTrends[r.Code] = r.Trends
			fmt.Printf("   🔥 %s: %d trend\n", r.Code, len(r.Trends))
		}
	}
	close(results)
	fmt.Printf("🌍 TOTAL: %d/%d countries with active trends\n",
		len(worldTrends), len(all))
	return worldTrends
}

// ════════════════════════════════════════
// 🏆 TOP 4: أقوى 4 ترندات عالمياً — أساس الفيديوهات الأربعة
// ══════════════════════════════════════════

type TopTrend struct {
	Rank       int    `json:"rank"`
	Trend      string `json:"trend"`
	Source     string `json:"source"`
	Continent  string `json:"continent"`
	Story      string `json:"story"`
	Hook       string `json:"hook"`
	Angle      string `json:"angle"`
	ViralScore int    `json:"viral_score"`
	BestRegion string `json:"best_region"`
}

// Top4Trends: AI يرتب أقوى 4 ترندات من كل العالم ويصنع قصة لكل واحد
func Top4Trends(worldTrends map[string][]string) []TopTrend {
	codeToContinent := map[string]string{}
	for _, cont := range World {
		for _, c := range cont.Countries {
			codeToContinent[c.Code] = cont.Name
		}
	}

	var b strings.Builder
	for code, trs := range worldTrends {
		limit := len(trs)
		if limit > 3 {
			limit = 3
		}
		for i := 0; i < limit; i++ {
			b.WriteString(fmt.Sprintf("%s|قارة:%s|%s\n",
				trs[i], codeToContinent[code], code))
		}
	}

	prompt := fmt.Sprintf(`ترندات Google الحقيقية اليوم من 28 دولة (6 قارات)، بصيغة:
الترند|القارة|كود الدولة
%s

المطلوب JSON فقط: اختر أقوى 4 ترندات عالمياً يمكن تحويلها لقصص مالية/أعمال ملهمة.
شرط مهم: الترندات الأربعة من دول/قارات مختلفة قدر الإمكان + لكل واحدة أفضل سوق نشر.
{"trends": [
 {"rank":1,"trend":"...","source":"US","continent":"🌎 أمريكا الشمالية",
  "story":"قصة مالية ملهمة 200 كلمة بالعربي","hook":"جملة صادمة أقل من 15 كلمة",
  "angle":"زاوية السرد","viral_score":9,"best_region":"US"},
 {"rank":2,...},{"rank":3,...},{"rank":4,...}
]}`, b.String())

	resp, err := ai.Chat("خبير فيروسية محتوى قصص المال العالمي", prompt)
	if err != nil {
		fmt.Println("   ⚠️ AI Top4 فشل → fallback للترندات الخام")
		return fallbackTop4(worldTrends)
	}

	var out struct {
		Trends []TopTrend `json:"trends"`
	}
	if json.Unmarshal([]byte(extractJSONLocal(resp)), &out) == nil && len(out.Trends) > 0 {
		if len(out.Trends) > 4 {
			out.Trends = out.Trends[:4] // 🔒 دائماً 4 بالضبط أو أقل
		}
		fmt.Println("\n🏆 TOP 4 GLOBAL TRENDS:")
		for _, t := range out.Trends {
			fmt.Printf("   #%d (%d/10) [%s %s] %s\n      → %s\n",
				t.Rank, t.ViralScore, t.Source, t.Continent, t.Trend, t.Hook)
		}
		return out.Trends
	}
	return fallbackTop4(worldTrends)
}

// fallbackTop4: لو الـ AI فشل — أول ترند من أقوى الأسواق
func fallbackTop4(worldTrends map[string][]string) []TopTrend {
	priority := []string{"US", "GB", "DE", "SA", "AE", "IN", "BR", "JP"}
	var out []TopTrend
	rank := 1
	for _, code := range priority {
		trs, ok := worldTrends[code]
		if !ok || len(trs) == 0 {
			continue
		}
		out = append(out, TopTrend{
			Rank: rank, Trend: trs[0], Source: code,
			Story:      "قصة مالية ملهمة عن: " + trs[0],
			Hook:       "ما حدش توقع ده!",
			Angle:      "دراما مالية",
			ViralScore: 5,
			BestRegion: code,
		})
		rank++
		if rank > 4 { // 🔒 أصلاً 4
			break
		}
	}
	return out
}

// extractJSONLocal: تنظيف رد النموذج → JSON نقي
func extractJSONLocal(raw string) string {
	s := strings.ReplaceAll(raw, "```json", "")
	s = strings.ReplaceAll(s, "```", "")
	start := strings.Index(s, "{")
	end := strings.LastIndex(s, "}")
	if start >= 0 && end > start {
		return s[start : end+1]
	}
	return s
}

// MatchStoryToTrend: مطابقة سريعة بدون AI (fallback)
func MatchStoryToTrend(trend string) string {
	matches := map[string]string{
		"inflation": "قصة الدولار الذي فقد نصف قيمته.. ومن ربح من الانهيار",
		"crypto":    "قصة الشاب الذي راهن آخر 100$ على عملة رقمية",
		"ai":        "قصة أول مليونير بالذكاء الاصطناعي",
		"stocks":    "قصة السهم الذي صعد 10000% في سنة واحدة",
		"gold":      "قصة الذهب: الرجل الذي باعه في الوقت الخطأ",
		"dollar":    "قصة العملات: صناعة الثروة من الانهيارات",
	}
	low := strings.ToLower(trend)
	for k, story := range matches {
		if strings.Contains(low, k) {
			return story
		}
	}
	return ""
}
