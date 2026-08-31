```go
package main

import (
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/Adam18691/moneyverse-stories/internal/edit"
	"github.com/Adam18691/moneyverse-stories/internal/hook"
	"github.com/Adam18691/moneyverse-stories/internal/meta"
	"github.com/Adam18691/moneyverse-stories/internal/prompts"
	"github.com/Adam18691/moneyverse-stories/internal/render"
	"github.com/Adam18691/moneyverse-stories/internal/schedule"
	"github.com/Adam18691/moneyverse-stories/internal/seo"
	"github.com/Adam18691/moneyverse-stories/internal/subs"
	"github.com/Adam18691/moneyverse-stories/internal/thumbs"
	"github.com/Adam18691/moneyverse-stories/internal/trends"
	"github.com/Adam18691/moneyverse-stories/internal/tts"
	"github.com/Adam18691/moneyverse-stories/internal/youtube"
)

const videosPerDay = 4

var trendForVideo []trends.TopTrend
var plan map[int]time.Time

var uploadCount int
var uploadMu sync.Mutex

func buildVideo(id int, wg *sync.WaitGroup) {
	defer wg.Done()

	start := time.Now()

	fmt.Printf("\n🎬 VIDEO %d START\n", id)

	// نتحقق من الكوتا قبل بدء العمل المكلف.
	uploadMu.Lock()
	quotaReached := uploadCount >= videosPerDay
	uploadMu.Unlock()

	if quotaReached {
		fmt.Printf(
			"🛑 VIDEO %d — الكوتا اليومية (%d) مكتملة — تخطي\n",
			id,
			videosPerDay,
		)
		return
	}

	story := ""
	hookText := ""

	// استخدام الترند المقابل للفيديو إن وجد.
	if id >= 1 && id <= len(trendForVideo) {
		t := trendForVideo[id-1]

		story = t.Story
		hookText = t.Hook

		fmt.Printf(
			"   🏆 TREND #%d [%s %s]: %s\n",
			t.Rank,
			t.Source,
			t.Continent,
			t.Trend,
		)

		fmt.Printf(
			"   💰 VIRAL SCORE: %d/10\n",
			t.ViralScore,
		)
	}

	// توليد Prompt احتياطي إذا لم توجد قصة من الترند.
	p := prompts.Generate(id)

	if story == "" {
		story = p.Story
	}

	if story == "" {
		fmt.Printf(
			"   ❌ VIDEO %d — لم يتم إنشاء Story — تخطي\n",
			id,
		)
		return
	}

	// إنشاء الـ Hook.
	h := hook.Generate(id)

	if hookText != "" {
		h.ScreenText = hookText
		h.VoiceLine = hookText
	}

	timeline := edit.BuildTimeline(h, 900)

	fmt.Printf(
		"   🎯 HOOK: %s | cuts: %d\n",
		h.ScreenText,
		len(timeline),
	)

	// =========================
	// Arabic TTS
	// =========================

	vo := fmt.Sprintf(
		"audio/%d_ar.wav",
		id,
	)

	if err := tts.Narrate(
		h.VoiceLine+"\n"+story,
		"ar",
		vo,
	); err != nil {
		fmt.Printf(
			"   ❌ TTS failed: %v — تخطي الفيديو\n",
			err,
		)
		return
	}

	// =========================
	// Multi-language Dubbing
	// =========================

	scriptLangs := map[string]string{
		"en": story,
		"es": story,
		"tr": story,
	}

	dubs := tts.DubAllLanguages(
		scriptLangs,
		id,
	)

	fmt.Printf(
		"   🌍 DUBS: %d languages\n",
		len(dubs),
	)

	// =========================
	// Subtitles
	// =========================

	tracks := subs.GenerateSubtitles(
		id,
		scriptLangs,
	)

	// =========================
	// Thumbnail
	// =========================

	thumbPath := fmt.Sprintf(
		"thumbs/thumb_%d.jpg",
		id,
	)

	if err := thumbs.Generate(
		id,
		h.ScreenText+" 💰",
		"assets/burning_money.jpg",
	); err != nil {
		fmt.Printf(
			"   ⚠️ Thumbnail failed: %v\n",
			err,
		)

		thumbPath = ""
	}

	// =========================
	// Render
	// =========================

	out := fmt.Sprintf(
		"output/money_%d.mp4",
		id,
	)

	if err := render.Build(
		story,
		vo,
		out,
	); err != nil {
		fmt.Printf(
			"   ❌ RENDER FAILED: %v — تخطي\n",
			err,
		)
		return
	}

	// التأكد من أن ملف الفيديو موجود فعلاً.
	if _, err := os.Stat(out); err != nil {
		fmt.Printf(
			"   ❌ OUTPUT FILE NOT FOUND: %v — تخطي\n",
			err,
		)
		return
	}

	// =========================
	// SEO Metadata
	// =========================

	seoMeta, err := seo.GenerateMetadata(story)

	if err != nil || seoMeta == nil || seoMeta.Title == "" {
		fmt.Printf(
			"   ⚠️ SEO generation failed — using fallback metadata\n",
		)

		seoMeta = &seo.Metadata{
			Title: fmt.Sprintf(
				"💰 %s | قصة ستغير نظرتك للمال",
				h.ScreenText,
			),

			Description: meta.BuildDescription(
				meta.DescriptionData{
					Hook: h.VoiceLine,
				},
			),

			Tags: p.Tags,
		}
	} else {
		lesson := ""

		if len(p.Angles) > 6 {
			lesson = p.Angles[6]
		}

		seoMeta.Description = meta.BuildDescription(
			meta.DescriptionData{
				Hook:   seoMeta.Description,
				Lesson: lesson,
			},
		)

		// إذا لم تُرجع خدمة SEO Tags، استخدم Tags الخاصة بالـ Prompt.
		if len(seoMeta.Tags) == 0 {
			seoMeta.Tags = p.Tags
		}
	}

	// =========================
	// Publish Schedule
	// =========================

	pubTime, hasSchedule := plan[id]

	var publishAt *time.Time

	if hasSchedule && !pubTime.IsZero() {
		publishAt = &pubTime

		fmt.Printf(
			"   ⏰ PUBLISH AT: %s\n",
			pubTime.Format(time.RFC3339),
		)
	} else {
		fmt.Printf(
			"   ⚠️ No valid publish schedule for VIDEO %d\n",
			id,
		)
	}

	// =========================
	// YouTube Upload
	// =========================

	_, err = youtube.Upload(
		out,
		youtube.Meta{
			Title:       seoMeta.Title,
			Description: seoMeta.Description,
			Tags:        seoMeta.Tags,
			LangTracks:  tracks,
			ThumbPath:   thumbPath,
			PublishAt:   publishAt,
		},
	)

	if err != nil {
		fmt.Printf(
			"   ❌ UPLOAD FAILED: %v\n",
			err,
		)
		return
	}

	// =========================
	// Confirm successful upload
	// =========================

	uploadMu.Lock()

	if uploadCount < videosPerDay {
		uploadCount++
	}

	currentUploads := uploadCount

	uploadMu.Unlock()

	fmt.Printf(
		"   📺 UPLOAD CONFIRMED: %d/%d\n",
		currentUploads,
		videosPerDay,
	)

	fmt.Printf(
		"✅ VIDEO %d DONE in %.0fs 💰🔥\n",
		id,
		time.Since(start).Seconds(),
	)
}

func main() {
	// =========================
	// Project directories
	// =========================

	directories := []string{
		"output",
		"audio",
		"thumbs",
		"subs",
		"scenes",
	}

	for _, d := range directories {
		if err := os.MkdirAll(d, 0755); err != nil {
			fmt.Printf(
				"❌ cannot create %s: %v\n",
				d,
				err,
			)
			return
		}
	}

	fmt.Println(
		"🚀 MONEYVERSE STORIES ENGINE",
	)

	fmt.Println(
		"════════════════════════════════════",
	)

	fmt.Printf(
		"🔒 DAILY QUOTA: %d videos FIXED\n",
		videosPerDay,
	)

	fmt.Println(
		"🌍 Sources: 28 countries | 6 continents",
	)

	fmt.Println(
		"🤖 SEO: Groq → Gemini → OpenRouter",
	)

	fmt.Println(
		"📺 Upload: OAuth 3-secrets | ⏰ smart publish",
	)

	fmt.Println(
		"════════════════════════════════════",
	)

	// =========================
	// PHASE 0
	// Worldwide Intelligence
	// =========================

	fmt.Println(
		"\n🌍 PHASE 0: WORLDWIDE INTELLIGENCE",
	)

	trends.PrintCoverage()

	worldTrends := trends.FetchWorldTrends()

	topTrends := trends.Top4Trends(
		worldTrends,
	)

	trendForVideo = topTrends

	// =========================
	// Daily video IDs
	// =========================

	ids := make(
		[]int,
		videosPerDay,
	)

	for i := range ids {
		ids[i] = i + 1
	}

	fmt.Printf(
		"💰 DAILY QUOTA: %d videos | 🏆 TOP TRENDS READY: %d\n",
		len(ids),
		len(topTrends),
	)

	// =========================
	// Smart Schedule
	// =========================

	regions := schedule.BuildRegions()

	if len(regions) == 0 {
		fmt.Println(
			"⚠️ No regions returned by scheduler — continuing without smart regional scheduling",
		)

		plan = make(map[int]time.Time)

		for _, id := range ids {
			plan[id] = time.Now().UTC()
		}
	} else {
		plan = schedule.PlanDay(
			ids,
			time.Now().UTC(),
			regions,
		)
	}

	// =========================
	// PHASE 1
	// Parallel Production
	// =========================

	fmt.Println(
		"\n⚡ PHASE 1: PARALLEL PRODUCTION",
	)

	var wg sync.WaitGroup

	for _, id := range ids {
		wg.Add(1)

		go buildVideo(
			id,
			&wg,
		)
	}

	wg.Wait()

	// =========================
	// Final Report
	// =========================

	uploadMu.Lock()
	finalUploadCount := uploadCount
	uploadMu.Unlock()

	fmt.Printf(
		"\n🏁 UPLOADS CONFIRMED: %d/%d — AUTO-PUBLISH AT PEAK HOURS ⏰🌍💰\n",
		finalUploadCount,
		videosPerDay,
	)
}
```
