package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
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

const DAILY_TARGET = 4 // فيديوهات يومياً

const pendingFile = "schedule/pending.json"

var plan map[int]time.Time

// ─────────────────────────────────────────────

// PendingEntry عنصر في جدول النشر (متوافق مع publish.go)
type PendingEntry struct {
	ID          int    `json:"id"`
	Filename    string `json:"filename"`
	Title       string `json:"title"`
	PublishAt   string `json:"publish_at"`
	Status      string `json:"status"`
}

func buildVideo(id int, wg *sync.WaitGroup) {
	defer wg.Done()
	start := time.Now()
	fmt.Printf("\n🎬 VIDEO %d START\n", id)

	// ── 1) القصة — جلب ترندات Google حسب يوم النشر
	story := ""
	if daily, err := trends.FetchDailyTrends("US"); err == nil {
		for _, tr := range daily {
			if s := trends.MatchStoryToTrend(tr); s != "" {
				story = s
				fmt.Printf("  🔥 TREND MATCHED: %s\n", tr)
				break
			}
		}
	} else {
		fmt.Printf("  ⚠️ Trends fetch failed (using fallback): %v\n", err)
	}
	p := prompts.Generate(id)
	if story == "" {
		story = p.Story // evergreen fallback
	}

	// ── 2) الهوك + timeline (هدف: 7 دقائق)
	h := hook.Generate(id)
	edit.BuildTimeline(h, 900)

	// ── 3) صوت عربي + دبلجة
	vo := fmt.Sprintf("audio/%d_ar.wav", id)
	tts.Narrate(h.VoiceLine+"\n"+story, "ar", vo)
	scriptLangs := map[string]string{"en": story, "es": story, "tr": story}
	dubs := tts.DubAllLanguages(scriptLangs, id)
	_ = dubs

	// ── 4) ترجمات VTT + ثامبنيل
	tracks := subs.GenerateSubtitles(id, scriptLangs)
	thumbs.Generate(id, h.ScreenText+" 💰", "assets/burning_money.jpg")

	// ── 5) بناء GStracer/melt
	out := fmt.Sprintf("output/money_%d.mp4", id)
	if err := render.Build(story, vo, out); err != nil {
		fmt.Printf("❌ VIDEO %d RENDER FAILED: %v\n", id, err)
		return // لا نكمل رفع فيديو غير موجود
	}

	// تأكد أن الملف موجود فعلاً قبل الرفع
	if _, err := os.Stat(out); os.IsNotExist(err) {
		fmt.Printf("❌ VIDEO %d: file not found at %s!\n", id, out)
		return
	}

	// ── 6) رفع + حفظ في الجدول
	pubTime := plan[id]
	fmt.Printf("  📅 Scheduled publish time: %s\n", pubTime.Format(time.RFC3339))

	videoID, err := youtube.Upload(out, youtube.Meta{
		Title:       fmt.Sprintf("💰 %s | قصص نجاح مالية تغير حياتك", h.ScreenText),
		Description: meta.BuildDescription(meta.DescriptionData{Hook: h.VoiceLine, Lesson: p.Angles[id%len(p.Angles)]}),
		Tags:        p.Tags,
		LangTracks:  tracks,
		ThumbPath:   fmt.Sprintf("thumbs/thumb_%d.jpg", id),
		PublishAt:   &pubTime,
	})

	// ✅ الإصلاح الأهم: فحص نتيجة الرفع فعلياً!
	if err != nil {
		fmt.Printf("❌ VIDEO %d YOUTUBE UPLOAD FAILED: %v\n", id, err)
		saveToPending(id, out, pubTime, "failed") // نسجله لإعادة المحاولة
		return
	}
	fmt.Printf("  ✅ Uploaded to YouTube, video ID: %s\n", videoID)

	saveToPending(id, out, pubTime, "pending")

	fmt.Printf("✅ VIDEO %d DONE in %.0fs 🔥\n", id, time.Since(start).Seconds())
}

// saveToPending يكتب الفيديو في schedule/pending.json ليقرأه publish.go
func saveToPending(id int, out string, pubTime time.Time, status string) {
	entries := []PendingEntry{}

	// اقرأ الجدول الحالي إن وجد
	if data, err := os.ReadFile(pendingFile); err == nil {
		if err := json.Unmarshal(data, &entries); err != nil {
			fmt.Printf("  ⚠️ pending.json corrupted, recreating: %v\n", err)
			entries = []PendingEntry{}
		}
	}

	entries = append(entries, PendingEntry{
		ID:        id,
		Filename:  filepath.Base(out),
		Title:     fmt.Sprintf("Money Story #%d", id),
		PublishAt: pubTime.Format(time.RFC3339),
		Status:    status,
	})

	data, err := json.MarshalIndent(entries, "", "  ")
	if err != nil {
		fmt.Printf("  ❌ Failed to marshal pending: %v\n", err)
		return
	}
	if err := os.WriteFile(pendingFile, data, 0644); err != nil {
		fmt.Printf("  ❌ Failed to write pending.json: %v\n", err)
		return
	}
	fmt.Printf("  📝 Saved to pending.json (status: %s)\n", status)
}

// ─────────────────────────────────────────────

func main() {
	for _, d := range []string{"output", "audio", "thumbs", "subs", "scenes", "schedule"} {
		os.MkdirAll(d, 0755)
	}

	fmt.Println("🚀 MONEYVERSE STORIES ENGINE — 4 videos/day — ALL LANGUAGES 🌍")

	// ✅ الإصلاح: 4 فيديوهات بدل 3
	ids := []int{1, 2, 3, 4}
	plan = schedule.PlanDay(ids, time.Now().UTC())

	// ✅ الإصلاح: اطبع الخطة للتشخيص
	for id, t := range plan {
		fmt.Printf("  📅 Plan: video %d → publish at %s\n", id, t.Format(time.RFC3339))
	}

	var wg sync.WaitGroup
	for _, id := range ids {
		wg.Add(1)
		go buildVideo(id, &wg) // parallel
	}
	wg.Wait()

	fmt.Println("🎬 ALL VIDEOS PROCESSED — CHECK LOGS FOR ACTUAL RESULTS 🎬")
}
