package seo

import (
	"encoding/json"
	"fmt"
	"strings"

	"tayyibat-money/internal/ai"
)

// ══════════════════════════════════════════
// 🔍 SEO ENGINE — عناوين ووصف وتاجات يوتيوب
// ══════════════════════════════════════════

// Meta: حزمة الـ SEO الكاملة للفيديو
type Meta struct {
	Title        string   `json:"title"`
	Description  string   `json:"description"`
	Tags         []string `json:"tags"`
	Hashtags     []string `json:"hashtags"`
}

// Generate: يبني حزمة SEO كاملة من الترند والقصة — سلاح الاكتشاف
func Generate(trend, story, hook string) Meta {

	prompt := fmt.Sprintf(`أنت خبير SEO يوتيوب أسطوري — فديوهاتك تتصدر الترند دايماً.

الترند: %s
القصة: %s
الهوك: %s

أخرج JSON فقط — عربي جذاب مصمم للنقر:

1. title: عنوان صادم 60-80 حرف — أرقام قوية + كلمات فضول (بدون كليك بيت منافق)
2. description: وصف 150-250 كلمة:
   - أول سطرين = الهوك (يظهر في البحث!)
   - ملخص القصة بتشويق بلا حرق
   - دعوة للإشترك والتفعيل
   - 3-5 هاشتاجات في النهاية
3. tags: 15-20 تاج — عربي + إنجليزي + طويل + قصير (كلمات البحث الفعلية)
4. hashtags: 3-5 هاشتاجات قصيرة جداً للعنوان

{"title":"...","description":"...","tags":["..."],"hashtags":["..."]}`,
		trend, cutStr(story, 300), hook)

	resp, err := ai.Chat("خبير SEO يوتيوب — عناوين تتصدر الترند", prompt)
	if err != nil {
		return fallbackMeta(trend, hook)
	}

	var meta Meta
	if json.Unmarshal([]byte(ai.ExtractJSON(resp)), &meta) != nil || meta.Title == "" {
		return fallbackMeta(trend, hook)
	}

	// 🛡️ تنظيف: عنوان ≤ 100 حرف (حد يوتيوب)
	if len(meta.Title) > 100 {
		meta.Title = meta.Title[:97] + "..."
	}
	if len(meta.Tags) == 0 {
		meta.Tags = fallbackTags(trend)
	}
	fmt.Printf("   🔍 SEO: \"%s\" | %d تاج\n", meta.Title, len(meta.Tags))
	return meta
}

// fallbackMeta: احتياطي لو الـ AI وقع
func fallbackMeta(trend, hook string) Meta {
	title := cutStr(trend+" | "+hook, 95)
	return Meta{
		Title:       title,
		Description: fmt.Sprintf("%s\n\n%s\n\n🔔 اشترك وفعّل الجرس ليصلك كل جديد!\n\n#ترند #قصص #حكايات", hook, trend),
		Tags:        fallbackTags(trend),
		Hashtags:    []string{"#ترند", "#قصص", "#حكايات"},
	}
}

func fallbackTags(trend string) []string {
	return []string{
		"ترند", "قصص", "حكايات", "قصة حقيقية", "قصة",
		"trend", "stories", "story", "قصص عربية", "أخبار",
		"ترند اليوم", "قصص واقعية", "قصة غريبة", trend,
	}
}

// cutStr: قص آمن للنصوص
func cutStr(s string, n int) string {
	if len(s) > n {
		return strings.TrimSpace(s[:n])
	}
	return s
}
