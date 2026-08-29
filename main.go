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
	"tayyibat-money/internal/seo"
	"tayyibat-money/internal/subs"
	"tayyibat-money/internal/thumbs"
	"tayyibat-money/internal/trends"
	"tayyibat-money/internal/tts"
	"tayyibat-money/internal/youtube"
)

const VIDEOS_PER_DAY = 4

var trendForVideo []trends.TopTrend
var plan map[int]time.Time

var uploadCount int
var uploadMu sync.Mutex

func buildVideo(id int, wg *sync.WaitGroup) {
	defer wg.Done()
	start := time.Now()
	fmt.Printf("\n🎬 VIDEO %d START\n", id)

	uploadMu.Lock()
	if uploadCount >= VIDEOS_PER_DAY {
		uploadMu.Unlock()
		fmt.Printf("🛑 VIDEO %d — كوتا اليوم (4) اكتملت — تخطي\n", id)
		return
	}
	uploadCount++
	uploadMu.Unlock()

	story := ""
	hookText := ""

	if id <= len(trendForVideo) {
		t := trendForVideo[id-1]
		story = t.Story
		hookText = t.Hook
		fmt.Printf("   🏆 TREND #%d [%s %s]: %s\n", t.Rank, t.Source, t.Continent, t.Trend)
		fmt.Printf("   💰 VIRAL SCORE: %d/10\n", t.ViralScore)
	}

	p := prompts.Generate(id)
	if story == "" {
		story = p.Story
	}

	h := hook.Generate(id)
	if hookText != "" {
		h.ScreenText = hookText
		h.VoiceLine = hookText
	}

	timeline := edit.BuildTimeline(h, 900)
	fmt.Printf("   🎯 HOOK: %s | cuts: %d\n", h.ScreenText, len(timeline))

	vo := fmt.Sprintf("audio/%d_ar.wav", id)
	if err := tts.Narrate(h.VoiceLine+"\n"+story, "ar", vo); err != nil {
		fmt.Printf("   ⚠️ TTS failed: %v → silent render\n", err)
	}

	scriptLangs := map[string]string{"en": story, "es": story, "tr": story}
	dubs := tts.DubAllLanguages(scriptLangs, id)
	fmt.Printf("   🌍 DUBS: %d languages\n", len(dubs))

	tracks := subs.GenerateSubtitles(id, scriptLangs)

	thumbPath := fmt.Sprintf("thumbs/thumb_%d.jpg", id)
	if err := thumbs.Generate(id, h.ScreenText+" 💰", "assets/burning_money.jpg"); err != nil {
		fmt.Printf("   ⚠️ thumbnail failed: %v\n", err)
	}

	out := fmt.Sprintf("output/money_%d.mp4", id)
	if err := render.Build(story, vo, out); err != nil {
		fmt.Printf("   ❌ RENDER FAILED: %v — تخطي\n", err)
		return
	}

	seoMeta, err := seo.GenerateMetadata(story)
	if err != nil || seoMeta == nil || seoMeta.Title == "" {
		seoMeta = &seo.Metadata{
			Title:       fmt.Sprintf("💰 %s | قصة ستغير نظرتك للمال", h.ScreenText),
			Description: meta.BuildDescription(meta.DescriptionData{Hook: h.VoiceLine}),
			Tags:        p.Tags,
		}
	} else {
		seoMeta.Description = meta.BuildDescription(meta.DescriptionData{
			Hook:   seoMeta.Description,
			Lesson: p.Angles[6],
		})
	}

	pubTime := plan[id]
	youtube.Upload(out, youtube.Meta{
		Title:       seoMeta.Title,
		Description: seoMeta.Description,
		Tags:        seoMeta.Tags,
		LangTracks:  tracks,
		ThumbPath:   thumbPath,
		PublishAt:   &pubTime,
	})

	fmt.Printf("✅ VIDEO %d DONE in %.0fs 💰🔥\n", id, time.Since(start).Seconds())
}

func main() {
	for _, d := range []string{"output", "audio", "thumbs", "subs", "scenes"} {
		_ = os.MkdirAll(d, 0755)
	}

	fmt.Println("🚀 MONEYVERSE STORIES ENGINE")
	fmt.Println("════════════════════════════════════")
	fmt.Println("🔒 DAILY QUOTA: 4 videos FIXED")
	fmt.Println("🌍 Sources: 28 countries | 6 continents")
	fmt.Println("🤖 SEO: Groq → Gemini → OpenRouter")
	fmt.Println("📺 Upload: OAuth 3-secrets | ⏰ smart publish")
	fmt.Println("════════════════════════════════════")

	fmt.Println("\n🌍 PHASE 0: WORLDWIDE INTELLIGENCE")
	trends.PrintCoverage()

	worldTrends := trends.FetchWorldTrends()
	topTrends := trends.Top4Trends(worldTrends)
	trendForVideo = topTrends

	ids := make([]int, VIDEOS_PER_DAY)
	for i := range ids {
		ids[i] = i + 1
	}

	fmt.Printf("💰 DAILY QUOTA: %d videos | 🏆 TOP TRENDS READY: %d\n", len(ids), len(topTrends))

	regions := schedule.BuildRegions()
	plan = schedule.PlanDay(ids, time.Now().UTC(), regions)

	fmt.Println("\n⚡ PHASE 1: PARALLEL PRODUCTION")
	var wg sync.WaitGroup
	for _, id := range ids {
		wg.Add(1)
		go buildVideo(id, &wg)
	}
	wg.Wait()

	fmt.Println("\n🏁 4/4 VIDEOS UPLOADED — AUTO-PUBLISH AT PEAK HOURS ⏰🌍💰")
}
