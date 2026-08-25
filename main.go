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
	"tayyibat-money/internal/schedule"
	"tayyibat-money/internal/subs"
	"tayyibat-money/internal/thumbs"
	"tayyibat-money/internal/trends"
	"tayyibat-money/internal/tts"
	"tayyibat-money/internal/youtube"
)

const DAILY_TARGET = 4 // 4 فيديوهات يومياً

var plan map[int]time.Time

func buildVideo(id int, wg *sync.WaitGroup) {
	defer wg.Done()
	start := time.Now()
	fmt.Printf("\n🎬 VIDEO %d START\n", id)

	// 1️⃣ ترندات Google اليومية → مطابقة قصة
	story := ""
	if daily, err := trends.FetchDailyTrends("US"); err == nil {
		for _, tr := range daily {
			if s := trends.MatchStoryToTrend(tr); s != "" {
				story = s
				fmt.Printf("   🔥 TREND MATCHED: %s\n", tr)
				break
			}
		}
	}
	p := prompts.Generate(id)
	if story == "" {
		story = p.Story // evergreen fallback
	}

	// 2️⃣ هوك 7 ثواني نفسي + timeline المونتاج
	h := hook.Generate(id)
	edit.BuildTimeline(h, 900)

	// 3️⃣ صوت عربي + دبلجة
	vo := fmt.Sprintf("audio/%d_ar.wav", id)
	tts.Narrate(h.VoiceLine+"\n"+story, "ar", vo)
	scriptLangs := map[string]string{"en": story, "es": story, "tr": story}
	dubs := tts.DubAllLanguages(scriptLangs, id)
	_ = dubs

	// 4️⃣ ترجمات VTT + ثامبنيل ذهبي
	tracks := subs.GenerateSubtitles(id, scriptLangs)
	thumbs.Generate(id, h.ScreenText+" 💰", "assets/burning_money.jpg")

	// 5️⃣ رندر GStreamer/melt
	out := fmt.Sprintf("output/money_%d.mp4", id)
	render.Build(story, vo, out)

	// 6️⃣ رفع + جدولة النشر الذكي
	pubTime := plan[id]
	youtube.Upload(out, youtube.Meta{
		Title:       fmt.Sprintf("💰 %s | قصة ستغير نظرتك للمال", h.ScreenText),
		Description: meta.BuildDescription(meta.DescriptionData{Hook: h.VoiceLine, Lesson: p.Angles[6]}),
		Tags:        p.Tags,
		LangTracks:  tracks,
		ThumbPath:   fmt.Sprintf("thumbs/thumb_%d.jpg", id),
		PublishAt:   &pubTime,
	})

	fmt.Printf("✅ VIDEO %d DONE in %.0fs 💰🔥\n", id, time.Since(start).Seconds())
}

func main() {
	for _, d := range []string{"output", "audio", "thumbs", "subs", "scenes"} {
		os.MkdirAll(d, 0755)
	}

	fmt.Println("🚀 MONEYVERSE STORIES ENGINE — 4 videos/day — ALL LANGUAGES ⏰🌍")

	ids := []int{1, 2, 3, }
	plan = schedule.PlanDay(ids, time.Now().UTC())

	var wg sync.WaitGroup
	for _, id := range ids {
		wg.Add(1)
		go buildVideo(id, &wg) // parallel حقيقي
	}
	wg.Wait()
	fmt.Println("🏁 4 VIDEOS UPLOADED — AUTO-PUBLISH AT PEAK HOURS ⏰💰")
}
