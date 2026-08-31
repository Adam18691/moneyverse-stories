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

const (
	videosPerDay = 4

	defaultTimelineDuration = 900

	filePermission = 0755
)

var (
	trendForVideo []trends.TopTrend
	plan          map[int]time.Time

	uploadCount int
	uploadMu    sync.Mutex
)

// buildVideo يبني فيديو واحد بالكامل.
func buildVideo(
	id int,
	wg *sync.WaitGroup,
) {
	defer wg.Done()

	start := time.Now()

	// حماية إضافية حتى لا يؤدي panic في أحد المكونات
	// إلى إيقاف بقية عملية الإنتاج.
	defer func() {
		if recovered := recover(); recovered != nil {
			fmt.Printf(
				"   💥 VIDEO %d PANIC: %v\n",
				id,
				recovered,
			)
		}
	}()

	fmt.Printf(
		"\n🎬 VIDEO %d START\n",
		id,
	)

	// =========================
	// Quota Check
	// =========================

	if quotaReached() {
		fmt.Printf(
			"🛑 VIDEO %d — الكوتا اليومية (%d) مكتملة — تخطي\n",
			id,
			videosPerDay,
		)
		return
	}

	// =========================
	// Story / Trend
	// =========================

	story := ""
	hookText := ""

	if id >= 1 && id <= len(trendForVideo) {
		trend := trendForVideo[id-1]

		story = trend.Story
		hookText = trend.Hook

		fmt.Printf(
			"   🏆 TREND #%d [%s %s]: %s\n",
			trend.Rank,
			trend.Source,
			trend.Continent,
			trend.Trend,
		)

		fmt.Printf(
			"   💰 VIRAL SCORE: %d/10\n",
			trend.ViralScore,
		)
	}

	// =========================
	// Prompt
	// =========================

	prompt := prompts.Generate(id)

	if story == "" {
		story = prompt.Story
	}

	if story == "" {
		fmt.Printf(
			"   ❌ VIDEO %d — لم يتم إنشاء Story — تخطي\n",
			id,
		)
		return
	}

	// =========================
	// Hook
	// =========================

	h := hook.Generate(id)

	if hookText != "" {
		h.ScreenText = hookText
		h.VoiceLine = hookText
	}

	if h.ScreenText == "" {
		h.ScreenText = "Moneyverse Stories"
	}

	if h.VoiceLine == "" {
		h.VoiceLine = h.ScreenText
	}

	timeline := edit.BuildTimeline(
		h,
		defaultTimelineDuration,
	)

	fmt.Printf(
		"   🎯 HOOK: %s | cuts: %d\n",
		h.ScreenText,
		len(timeline),
	)

	// =========================
	// Arabic TTS
	// =========================

	audioPath := fmt.Sprintf(
		"audio/%d_ar.wav",
		id,
	)

	narration := h.VoiceLine + "\n" + story

	if err := tts.Narrate(
		narration,
		"ar",
		audioPath,
	); err != nil {
		fmt.Printf(
			"   ❌ TTS failed: %v — تخطي الفيديو\n",
			err,
		)
		return
	}

	if !fileExists(audioPath) {
		fmt.Printf(
			"   ❌ TTS output not found: %s — تخطي الفيديو\n",
			audioPath,
		)
		return
	}

	// =========================
	// Multi-language Dubbing
	// =========================

	// نحافظ على واجهة المشروع الحالية.
	// الترجمة الفعلية يمكن تطويرها داخل tts.DubAllLanguages.
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

	fmt.Printf(
		"   📝 SUBTITLES: %d tracks\n",
		len(tracks),
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

	// إذا لم يتم إنشاء الصورة فعليًا،
	// لا نرسل مسارًا غير موجود إلى YouTube.
	if thumbPath != "" && !fileExists(thumbPath) {
		fmt.Printf(
			"   ⚠️ Thumbnail output not found: %s\n",
			thumbPath,
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
		audioPath,
		out,
	); err != nil {
		fmt.Printf(
			"   ❌ RENDER FAILED: %v — تخطي\n",
			err,
		)
		return
	}

	if !fileExists(out) {
		fmt.Printf(
			"   ❌ OUTPUT FILE NOT FOUND: %s — تخطي\n",
			out,
		)
		return
	}

	// =========================
	// SEO Metadata
	// =========================

	seoMeta, err := seo.GenerateMetadata(
		story,
	)

	if err != nil ||
		seoMeta == nil ||
		seoMeta.Title == "" {

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

			Tags: prompt.Tags,
		}
	} else {
		lesson := ""

		if len(prompt.Angles) > 6 {
			lesson = prompt.Angles[6]
		}

		seoMeta.Description = meta.BuildDescription(
			meta.DescriptionData{
				Hook:   seoMeta.Description,
				Lesson: lesson,
			},
		)

		if len(seoMeta.Tags) == 0 {
			seoMeta.Tags = prompt.Tags
		}
	}

	// =========================
	// Publish Schedule
	// =========================

	pubTime, hasSchedule := plan[id]

	var publishAt *time.Time

	if hasSchedule && !pubTime.IsZero() {
		scheduledUTC := pubTime.UTC()

		publishAt = &scheduledUTC

		fmt.Printf(
			"   ⏰ PUBLISH AT: %s\n",
			scheduledUTC.Format(time.RFC3339),
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

	videoID, err := youtube.Upload(
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

	fmt.Printf(
		"   📺 YOUTUBE UPLOAD SUCCESS: %v\n",
		videoID,
	)

	// =========================
	// Confirm Upload
	// =========================

	currentUploads := incrementUploadCount()

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

// quotaReached يتحقق من الكوتا بشكل آمن.
func quotaReached() bool {
	uploadMu.Lock()
	defer uploadMu.Unlock()

	return uploadCount >= videosPerDay
}

// incrementUploadCount يزيد عداد الرفع بشكل آمن.
func incrementUploadCount() int {
	uploadMu.Lock()
	defer uploadMu.Unlock()

	if uploadCount < videosPerDay {
		uploadCount++
	}

	return uploadCount
}

// fileExists يتحقق من وجود الملف وأنه ليس مجلدًا.
func fileExists(path string) bool {
	info, err := os.Stat(path)

	if err != nil {
		return false
	}

	return !info.IsDir()
}

// prepareDirectories ينشئ جميع مجلدات المشروع المطلوبة.
func prepareDirectories() error {
	directories := []string{
		"output",
		"audio",
		"thumbs",
		"subs",
		"scenes",
		"schedule",
	}

	for _, directory := range directories {
		if err := os.MkdirAll(
			directory,
			filePermission,
		); err != nil {
			return fmt.Errorf(
				"cannot create %s: %w",
				directory,
				err,
			)
		}
	}

	return nil
}

// resetDailyState يضمن بداية نظيفة للعداد عند تشغيل البرنامج.
func resetDailyState() {
	uploadMu.Lock()
	uploadCount = 0
	uploadMu.Unlock()
}

func main() {
	// =========================
	// Project Initialization
	// =========================

	if err := prepareDirectories(); err != nil {
		fmt.Printf(
			"❌ PROJECT INITIALIZATION FAILED: %v\n",
			err,
		)

		os.Exit(1)
	}

	resetDailyState()

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
	// Daily Video IDs
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

		plan = make(
			map[int]time.Time,
			len(ids),
		)

		now := time.Now().UTC()

		for _, id := range ids {
			plan[id] = now
		}
	} else {
		plan = schedule.PlanDay(
			ids,
			time.Now().UTC(),
			regions,
		)
	}

	// =========================
	// Schedule Report
	// =========================

	fmt.Println(
		"\n⏰ DAILY PUBLISH PLAN",
	)

	for _, id := range ids {
		if publishAt, ok := plan[id]; ok &&
			!publishAt.IsZero() {

			fmt.Printf(
				"   🎬 VIDEO %d → %s\n",
				id,
				publishAt.UTC().Format(
					time.RFC3339,
				),
			)
		} else {
			fmt.Printf(
				"   ⚠️ VIDEO %d → NO SCHEDULE\n",
				id,
			)
		}
	}

	// =========================
	// PHASE 1
	// Parallel Production
	// =========================

	fmt.Println(
		"\n⚡ PHASE 1: PARALLEL PRODUCTION",
	)

	var wg sync.WaitGroup

	wg.Add(len(ids))

	for _, id := range ids {
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

	fmt.Println(
		"\n════════════════════════════════════",
	)

	fmt.Println(
		"🏁 MONEYVERSE STORIES FINAL REPORT",
	)

	fmt.Printf(
		"📺 UPLOADS CONFIRMED: %d/%d\n",
		finalUploadCount,
		videosPerDay,
	)

	if finalUploadCount == videosPerDay {
		fmt.Println(
			"✅ DAILY QUOTA COMPLETED",
		)
	} else {
		fmt.Printf(
			"⚠️ DAILY QUOTA NOT COMPLETED: %d/%d\n",
			finalUploadCount,
			videosPerDay,
		)
	}

	fmt.Println(
		"⏰ AUTO-PUBLISH AT PEAK HOURS 🌍💰",
	)

	fmt.Println(
		"════════════════════════════════════",
	)
}
```
