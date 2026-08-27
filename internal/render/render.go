package render

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"tayyibat-money/internal/music"
	"tayyibat-money/internal/trends"
)

// ============================================================
//  BuildStory — الخط السينمائي الكامل (نفس توقيع main.go)
//  out: مسار الملف النهائي | id: رقم الفيديو
// ============================================================

func BuildStory(out string, id int) error {
	os.MkdirAll("scenes", 0o755)
	os.MkdirAll("music_cache", 0o755)
	os.MkdirAll("output", 0o755)

	start := time.Now()

	// ---- 1) مقاطع النخبة: الأطول + أعلى دقة + زوايا درون سينمائية ----
	queries := cinematicQueries(id)
	elite := trends.FetchBestCinematic(queries, 10)
	if len(elite) == 0 {
		fmt.Println("⚠️ RENDER: no elite clips, aborting story render")
		return fmt.Errorf("no clips fetched")
	}

	var clipPaths []string
	for _, c := range elite {
		p, err := trends.Download(c, "scenes")
		if err == nil {
			clipPaths = append(clipPaths, p)
		}
	}
	fmt.Printf("🎞️ VIDEO %d: %d elite clips ready\n", id, len(clipPaths))

	// ---- 2) الموسيقى: مقدمة حماسية + خلفية هادئة مريحة ----
	introTrack := music.Pick(true)
	mainTrack := music.Pick(false)
	introPath := fmt.Sprintf("music_cache/intro_%d.mp3", id)
	mainPath := fmt.Sprintf("music_cache/main_%d.mp3", id)
	music.Download(introTrack, introPath)
	music.Download(mainTrack, mainPath)

	// ---- 3) المقدمة: 7 ثوانٍ من أقوى لقطة + موسيقى حماسية ----
	introOut := fmt.Sprintf("output/intro_%d.mp4", id)
	buildIntro(clipPaths[0], introPath, introOut)

	// ---- 4) الجسم: 15 أو 30 دقيقة حسب رقم الفيديو ----
	targetMin := 15
	if id%2 == 0 {
		targetMin = 30 // نُوزّع: فيديوهات زوجية 30 دقيقة، فردية 15
	}

	err := buildMain(clipPaths, mainPath, out, targetMin)
	if err != nil {
		return err
	}

	fmt.Printf("🎬 VIDEO %d RENDERED (%d min target) in %.0fs\n",
		id, targetMin, time.Since(start).Seconds())
	return nil
}

// cinematicQueries: جمل بحث سينمائية — تتغير حسب id حتى يختلف كل فيديو
func cinematicQueries(id int) []string {
	pool := [][]string{
		{"money falling slow motion", "gold coins macro", "luxury watch closeup"},
		{"city skyline aerial night", "businessman walking city", "office skyscraper glass"},
		{"stock market screen", "cash counting machine", "credit card luxury"},
		{"drone ocean coast", "supercar driving night", "penthouse interior luxury"},
	}
	return pool[id%len(pool)]
}

// buildIntro: مقدمة 7 ثوانٍ (175 إطار × 25fps) بتكبير سينمائي بطيء
func buildIntro(bestClip, musicPath, outPath string) {
	cmd := exec.Command("melt",
		bestClip+" in=0 out=175",
		"audio="+musicPath,
		"-consumer", "avformat:"+outPath,
		"vcodec=libx264", "preset=fast", "crf=20", "threads=0",
		"acodec=aac", "ab=192k", "pix_fmt=yuv420p")
	cmd.Stdout, cmd.Stderr = os.Stdout, os.Stderr
	if err := cmd.Run(); err != nil {
		fmt.Printf("⚠️ INTRO skipped: %v\n", err)
	} else {
		fmt.Println("🎬 INTRO 7s DONE")
	}
}

// buildMain: جسم الفيديو — يلف المقاطع النخبوية دوريًا حتى يصل للمدة المطلوبة
func buildMain(clips []string, musicPath, outPath string, targetMin int) error {
	seconds := targetMin * 60
	framesPerClip := 12 * 25 // 12 ثانية لكل مقطع × 25fps
	totalNeeded := (seconds / 12) + 1

	args := []string{}

	// نكرر المقاطع دوريًا حتى نغطي المدة (15 دقيقة ≈ 75 مقطعًا / 30 دقيقة ≈ 150)
	for i := 0; i < totalNeeded; i++ {
		c := clips[i%len(clips)]
		in := (i / len(clips)) * framesPerClip // إزاحة داخل المقطع حتى لا يتكرر نفس الجزء
		args = append(args, fmt.Sprintf("%s in=%d out=%d", c, in, in+framesPerClip))
	}

	// موسيقى هادئة خافتة تحت الفيديو
	args = append(args, "track=1", "audio="+musicPath, "mix=-17dB")

	args = append(args,
		"-consumer", "avformat:"+outPath,
		"vcodec=libx264", "preset=medium", "crf=22", "threads=0",
		"acodec=aac", "ab=192k", "pix_fmt=yuv420p",
	)

	cmd := exec.Command("melt", args...)
	cmd.Stdout, cmd.Stderr = os.Stdout, os.Stderr
	fmt.Printf("🎬 MAIN BUILD: %d segments → %s (%d min)\n",
		totalNeeded, filepath.Base(outPath), targetMin)
	return cmd.Run()
}
