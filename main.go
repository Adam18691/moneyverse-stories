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

const VIDEOS_PER_DAY = 4

var trendForVideo []trends.TopTrend
var plan map[int]time.Time

var uploadCount int
var uploadMu sync.Mutex

func buildVideo(id int, wg *sync.WaitGroup) {
	defer wg.Done()

	start := time.Now()
	fmt.Printf("\n🎬 VIDEO %d START\n", id)

	// لا نحسب الفيديو ضمن الكوتا إلا بعد نجاح الرفع إلى YouTube.
	uploadMu.Lock()

	if uploadCount >= VIDEOS_PER_DAY {
		uploadMu.Unlock()

		fmt.Printf(
			"🛑 VIDEO %d — كوتا اليوم (%d) اكتملت — تخطي\n",
			id,
			VIDEOS_PER_DAY,
		)

		return
	}

	uploadMu.Unlock()

	story := ""
	hookText := ""

	if id <= len(trendForVideo) {
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

	fmt.Printf(
		"   🎯 HOOK: %s | cuts: %d\n",
		h.ScreenText,
		len(timeline),
	)

	// Arabic TTS
	vo := fmt.Sprintf("audio/%d_ar.wav", id)

	if err := tts.Narrate(
		h.VoiceLine+"\n"+story,
		"ar",
		vo,
	); err != nil {

		fmt.Printf(
			"   ⚠️ TTS failed: %v → silent render\n",
			err,
		)
	}

	// اللغات المدعومة حاليًا.
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

	// Subtitles
	tracks := subs.GenerateSubtitles(
		id,
		scriptLangs,
	)

	// Thumbnail
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
			"   ⚠️ thumbnail failed: %v\n",
			err,
		)

		thumbPath = ""
	}

	// Render
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

	// SEO Metadata
	seoMeta, err := seo.GenerateMetadata(story)

	if err != nil ||
		seoMeta == nil ||
		seoMeta.Title == "" {

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
	}

	// وقت النشر
	pubTime := plan[id]

	// رفع الفيديو
	_, err = youtube.Upload(
		out,
		youtube.Meta{
			Title:       seoMeta.Title,
			Description: seoMeta.Description,
			Tags:        seoMeta.Tags,
			LangTracks:  tracks,
			ThumbPath:   thumbPath,
			PublishAt:   &pubTime,
		},
	)

	if err != nil {

		fmt.Printf(
			"   ❌ UPLOAD FAILED: %v\n",
			err,
		)

		return
	}

	// لا يتم احتساب الفيديو إلا بعد نجاح الرفع.
	uploadMu.Lock()

	if uploadCount < VIDEOS_PER_DAY {
		uploadCount++
	}

	uploadMu.Unlock()

	fmt.Printf(
		"✅ VIDEO %d DONE in %.0fs 💰🔥\n",
		id,
		time.Since(start).Seconds(),
	)
}

func main() {

	// إنشاء مجلدات المشروع.
	for _, d := range []string{
		"output",
		"audio",
		"thumbs",
		"subs",
		"scenes",
	} {

		if err := os.MkdirAll(
			d,
			0755,
		); err != nil {

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
		VIDEOS_PER_DAY,
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

	// Worldwide Intelligence
	fmt.Println(
		"\n🌍 PHASE 0: WORLDWIDE INTELLIGENCE",
	)

	trends.PrintCoverage()

	worldTrends := trends.FetchWorldTrends()

	topTrends := trends.Top4Trends(
		worldTrends,
	)

	trendForVideo = topTrends

	// IDs للفيديوهات اليومية.
	ids := make(
		[]int,
		VIDEOS_PER_DAY,
	)

	for i := range ids {
		ids[i] = i + 1
	}

	fmt.Printf(
		"💰 DAILY QUOTA: %d videos | 🏆 TOP TRENDS READY: %d\n",
		len(ids),
		len(topTrends),
	)

	// Smart schedule
	regions := schedule.BuildRegions()

	plan = schedule.PlanDay(
		ids,
		time.Now().UTC(),
		regions,
	)

	// Parallel production
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

	fmt.Printf(
		"\n🏁 UPLOADS CONFIRMED: %d/%d — AUTO-PUBLISH AT PEAK HOURS ⏰🌍💰\n",
		uploadCount,
		VIDEOS_PER_DAY,
	)
}
