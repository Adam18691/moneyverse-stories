package main

import (
	"fmt"
	"os"
	"sync"
	"time"

	"tayyibat-money/internal/edit"
	"tayyibat-money/internal/hook"
	"tayyibat-money/internal/meta"
	"tayyibat-money/internal/prompts"
	"tayyibat-money/internal/render"
	"tayyibat-money/internal/subs"
	"tayyibat-money/internal/thumbs"
	"tayyibat-money/internal/trends"
	"tayyibat-money/internal/tts"
	"tayyibat-money/internal/youtube"
)

const DAILY_TARGET = 4

func buildVideo(id int, wg *sync.WaitGroup) {
	defer wg.Done()
	start := time.Now()
	fmt.Printf("\n🎬 VIDEO %d START\n", id)

	// 1️⃣ 🔥 جلب ترندات Google اليومية ومطابقتها مع القصص
	dailyTrends, _ := trends.FetchDailyTrends("US")
	story := ""
	for _, tr := range dailyTrends {
		if s := trends.MatchStoryToTrend(tr); s != "" {
			story = s
			fmt.Printf("   🔥 TREND MATCHED: %s → %s\n", tr, story[:30])
			break
		}
	}
	if story == "" {
		p := prompts.Generate(id) // fallback: قصة evergreen
		story = p.Story
	}

	// 2️⃣ 🎯 هوك 7 ثواني نفسي مدروس
	h := hook.Generate(id)
	edit.BuildTimeline(h, 900) // طباعة الـ cuts + interrupts

	// 3️⃣ صوت + دبلجة + ثامبنيل + ترجمات
	vo := fmt.Sprintf("audio/%d_ar.wav", id)
	tts.Narrate(h.VoiceLine+"\n"+story, "ar", vo)
	scriptLangs := map[string]string{"en": story, "tr": story, "es": story}
	tracks := subs.GenerateSubtitles(id, scriptLangs)
	thumbs.Generate(id, h.ScreenText+" 💰", "assets/burning_money.jpg")

	// 4️⃣ رندر بالمونتاج السينمائي (cuts + zoom + interrupts)
	out := fmt.Sprintf("output/money_%d.mp4", id)
	render.Build(story, vo, out)

	// 5️⃣ 🌍 اختيار أفضل وقت نشر حسب الترندات العالمية
	bestTime := trends.GlobalTrendWindows[0] // الخليج افتراضياً
	for _, w := range trends.GlobalTrendWindows {
		if nowUTC().Hour()+2 == w.BestHourUTC { // نشر قبل الذروة بساعتين
			bestTime = w
		}
	}
	fmt.Printf("   ⏰ Publish optimized for: %s (%s)\n", bestTime.Region, bestTime.Audience)

	// 6️⃣ رفع Public فعلي
	youtube.Upload(out, youtube.Meta{
		Title:       fmt.Sprintf("💰 %s", h.ScreenText),
		Description: meta.BuildDescription(meta.DescriptionData{Hook: h.VoiceLine}),
		Tags:        []string{"قصص نجاح", "المال", "ترند", "مليونير"},
		LangTracks:  tracks,
		ThumbPath:   fmt.Sprintf("thumbs/thumb_%d.jpg", id),
	})

	fmt.Printf("✅ VIDEO %d LIVE in %.0fs 💰🔥\n", id, time.Since(start).Seconds())
}

func main() {
	for _, d := range []string{"output", "audio", "thumbs", "subs"} {
		os.MkdirAll(d, 0755)
	}
	fmt.Println("🚀 MONEYVERSE v16 — TREND-AWARE HOOK ENGINE — 4 videos/day")
	var wg sync.WaitGroup
	for i := 1; i <= DAILY_TARGET; i++ {
		wg.Add(1)
		go buildVideo(i, &wg)
	}
	wg.Wait()
	fmt.Println("🏁 DONE 💰🌍")
}

func nowUTC() time.Time { return time.Now().UTC() }
