package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"moneyverse-stories/internal/render"
	"moneyverse-stories/internal/thumbs"
	"moneyverse-stories/internal/trends"
	"moneyverse-stories/internal/tts"
	"moneyverse-stories/internal/youtube"
)

// ============================================================
//  💰 GOD Money Stories Engine
//  خط إنتاج سينمائي كامل:
//  نخبة مقاطع → صوت عصبي نخبوي → رندر سينمائي 15/30 دقيقة
//  → ثامبنيل 4K ذهبي → نشر تلقائي Public على يوتيوب
// ============================================================

// Hook الخطاف الافتتاحي — يحدد نجاح الفيديو
type Hook struct {
	VoiceLine string `json:"voice_line"`
	Title     string `json:"title"`
}

var hooks = []Hook{
	{
		VoiceLine: "توقف! قبل أن تكمل، خذ هذه المعاينة، وعد أحدهم أنك رأيت المستقبل اليوم.",
		Title:     "يقولون إنها حكاية تخييلية... إلا أنها تحدث الآن",
	},
	{
		VoiceLine: "لو أعطيتك مفتاحًا واحدًا ليفهم الأثرياء ثرواتهم، هل تهتم؟ هذا هو.",
		Title:     "المفتاح المخفي الذي لا يعرفه إلا الأثرياء",
	},
	{
		VoiceLine: "في عام 2008 انهارت أسواق المال... لكن شخصًا واحدًا وقف في الاتجاه المعاكس وربح المليارات.",
		Title:     "من الاحتكار إلى الملكية",
	},
	{
		VoiceLine: "هذه ليست قصة عن الحظ... هذه قصة عن قرار واحد غيّر كل شيء.",
		Title:     "قرار واحد... وحياة بأكملها تغيرت",
	},
}

// buildMeta تفاصيل الرفع لليوتيوب — عناوين وأوصاف محسّنة للنيتش
func buildMeta(id int, h Hook) youtube.Meta {
	title := fmt.Sprintf("%s 💰 | قصص وأسرار المال #%d", h.Title, id+1)

	desc := fmt.Sprintf(`%s

🎬 قصة سينمائية عن المال والنجاح بلغة يفهمها الجميع.
🔔 اشترك وفعّل الجرس لتصلك الحكاية الجديدة يوميًا.
💬 اكتب في التعليقات: ما أكثر سرٍّ أدهشك في هذه القصة؟

#قصص_مال #ثروة #تنمية_ذاتية #أسرار_النجاح #قصص_واقعية`, h.Title)

	return youtube.Meta{
		Title:       title,
		Description: desc,
		Tags: []string{
			"قصص المال", "الثروة", "أسرار النجاح", "قصص واقعية",
			"تنمية ذاتية", "استثمار", "money stories", "success story",
		},
	}
}

// ============================================================
//  produceVideo — خط الإنتاج الكامل لفيديو واحد
// ============================================================
func produceVideo(id int) error {
	h := hooks[id%len(hooks)]
	story := buildStoryText(id, h)

	// ============ 1) المقاطع السينمائية النخبوية ============
	fmt.Printf("🔍 VIDEO %d: fetching elite cinematic clips...\n", id)
	elite := trends.FetchBestCinematic(cinematicQueries(id), 10)

	var clipPaths []string
	for _, c := range elite {
		if p, err := trends.Download(c, "scenes"); err == nil {
			clipPaths = append(clipPaths, p)
		}
	}
	if len(clipPaths) == 0 {
		return fmt.Errorf("video %d: no elite clips available", id)
	}
	fmt.Printf("🎞️ VIDEO %d: %d elite clips ready\n", id, len(clipPaths))

	// ============ 2) الصوت — مذيع عصبي نخبوي (Edge Neural) ============
	os.MkdirAll("audio", 0o755)
	vo := filepath.Join("audio", fmt.Sprintf("%d_ar.wav", id))
	if err := tts.SmartSpeak(id, h.VoiceLine+" "+story, vo); err != nil {
		fmt.Printf("⚠️ VOICE %d failed: %v\n", id, err)
		return fmt.Errorf("voice %d: %w", id, err)
	}
	fmt.Printf("🎙️ VIDEO %d: voice ready\n", id)

	// ============ 3) الرندر السينمائي (مقدمة 7 ثوانٍ + جسم 15/30 دقيقة) ============
	out := filepath.Join("output", fmt.Sprintf("money_%d.mp4", id))
	if err := render.BuildStory(out, id); err != nil {
		return fmt.Errorf("render %d: %w", id, err)
	}
	fmt.Printf("🎬 VIDEO %d: render done → %s\n", id, out)

	// ============ 4) الثامبنيل — صور 4K دراماتيكية + نص ذهبي ثلاثي الأبعاد ============
	err := thumbs.Generate(id, shortTitle(h.Title), "")
	if err != nil {
		fmt.Printf("⚠️ THUMB %d failed (continuing): %v\n", id, err)
	} else {
		fmt.Printf("🖼️ VIDEO %d: thumb done\n", id)
	}

	// ============ 5) الرفع لليوتيوب — Public مباشرة ============
	meta := buildMeta(id, h)
	meta.ThumbPath = filepath.Join("thumbs", fmt.Sprintf("thumb_%d.jpg", id))

	vidID, err := youtube.Upload(out, meta)
	if err != nil {
		return fmt.Errorf("upload %d: %w", id, err)
	}

	fmt.Printf("🎉 VIDEO %d LIVE: https://youtu.be/%s\n", id, vidID)
	return nil
}

// ============================================================
//  نصوص القصص — توسّع تلقائيًا حسب مدة الفيديو المستهدفة
// ============================================================
func buildStoryText(id int, h Hook) string {
	base := []string{
		`في مدينة لا تنام، كان هناك شاب يعدّ القطع النقدية على مكتب خشبي قديم، يراقب من النافذة أضواء الأبراج التي لا تخصه بعد... فجأة جاءه عرض غريب من مجهول: عنوان واحد فقط، وشروط لن يصدقها أحد.`,
		`قال له جدّه ذات مساء: «النقود لا تنام يا بني، لكنها تختار أصحابها بعناية.» ثم وضع في يده عملة مؤرخة بعام صعب، وقال: احفظ هذه حتى تأتي اللحظة.`,
		`بدأت الحكاية برقم قياسي خرج عن كل التوقعات... ثم انكشف تدريجيًا ذلك العقل المدبر الذي أعاد رسم سماء المدينة خلال ليلة واحدة دون أن ينام.`,
		`كان يجلس في آخر مقعد بالقاعة يستمع لمن يقولون إن الثروة تعني الحظ... وبعد عشر سنوات، صار نفسوه هم الدرس الأول في كل جامعة اقتصاد بالمدينة.`,
	}

	txt := base[id%len(base)]
	txt += " " + h.VoiceLine + " "

	// طبقات سرد — توسع حقيقي يبلغ قراءة ~15 دقيقة
	extensions := []string{
		" وفي تلك اللحظة نفسها، بدأ العالم الخارجي يتحرك على نحوٍ لم يتوقعه أحد...",
		" لكن الحقيقة التي لم يعرفها أحد أنها كانت البداية وليست النهاية إطلاقًا.",
		" وبعد سنوات طويلة، أدرك الجميع أن تلك اللحظة حددت مصير المدينة بأسرها.",
		" الصفقة كانت تبدو مستحيلة على الورق... لكنها بُنيت على فهم أعمق لطبيعة الناس والمال.",
		" وكلما توغل في التفاصيل، أدرك أن السر ليس في الكسب... بل في الشراكة الذكية.",
		" خلف كل رقم ضخم، كان هناك إنسان عادي اتخذ قرارًا غير عادي في الوقت المناسب تمامًا.",
	}

	for i := 0; i < 40; i++ {
		txt += extensions[i%len(extensions)] + " "
	}
	return txt
}

// cinematicQueries جمل بحث سينمائية — تتغير لكل فيديو (درون/بطيء/فاخر)
func cinematicQueries(id int) []string {
	pool := [][]string{
		{"money falling slow motion", "gold coins macro", "luxury watch closeup"},
		{"city skyline aerial night", "businessman walking city", "office skyscraper glass"},
		{"stock market screen", "cash counting machine", "credit card luxury"},
		{"drone ocean coast", "supercar driving night", "penthouse interior luxury"},
	}
	return pool[id%len(pool)]
}

// shortTitle نسخة قصيرة للنص على الثامبنيل
func shortTitle(t string) string {
	r := []rune(t)
	if len(r) > 18 {
		return string(r[:18])
	}
	return t
}

// ============================================================
//  main — تشغيل متوازٍ ذكي (4 عمال)
// ============================================================
func main() {
	count := 4
	if n := os.Getenv("VIDEOS_COUNT"); n != "" {
		fmt.Sscanf(n, "%d", &count)
	}
	if count <= 0 {
		count = 4
	}

	fmt.Printf("🔥 GOD ENGINE START | videos=%d | parallel=4\n", count)

	// حفظ حالة التشغيل
	os.MkdirAll("schedule", 0o755)
	state, _ := json.MarshalIndent(map[string]interface{}{
		"run":   os.Getenv("GITHUB_RUN_NUMBER"),
		"count": count,
	}, "", "  ")
	os.WriteFile(filepath.Join("schedule", "run_state.json"), state, 0o644)

	jobs := make(chan int)
	var wg sync.WaitGroup

	for w := 0; w < 4; w++ { // 4 عمليات متوازية
		wg.Add(1)
		go func(worker int) {
			defer wg.Done()
			for id := range jobs {
				fmt.Printf("\n🚀 WORKER %d → VIDEO %d =================\n", worker, id)
				if err := produceVideo(id); err != nil {
					fmt.Printf("❌ VIDEO %d FAILED: %v\n", id, err)
				}
			}
		}(w)
	}

	for i := 0; i < count; i++ {
		jobs <- i
	}
	close(jobs)
	wg.Wait()

	fmt.Println("\n🏁 GOD ENGINE FINISHED")
}
