package render

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"

	"tayyibat-money/internal/music"
	"tayyibat-money/internal/thumbs"
)

// ══════════════════════════════════════════
// 🎬 RENDER ENGINE — تجميع الفيديو النهائي
// صور + صوت TTS + موسيقى خلفية → MP4 يوتيوب
// ══════════════════════════════════════════

// Scene: مشهد واحد من الفيديو
type Scene struct {
	ImagePath string  // صورة المشهد
	Text      string  // نص التعليق (اختياري للوج)
	Duration  float64 // مدة المشهد بالثواني
}

// Input: مدخلات الرندر الكاملة
type Input struct {
	VideoID   int
	Scenes    []Scene // مشاهد القصة
	VoicePath string  // ملف صوت TTS كامل
	MusicEmotion string // shock/greed/hope/fear
	ThumbPath string  // ثامبنيل جاهز (اختياري)
}

// Output: ناتج الرندر
type Output struct {
	VideoPath string
	ThumbPath string
	Duration  float64
}

// ══════════════════════════════════════════
// 🎬 RenderVideo — خط الإنتاج الكامل
// ══════════════════════════════════════════

func RenderVideo(in Input) (Output, error) {

	if in.VideoID <= 0 {
		return Output{}, fmt.Errorf("invalid video id")
	}
	if in.VoicePath == "" || !fileExists(in.VoicePath) {
		return Output{}, fmt.Errorf("ملف الصوت غير موجود: %s", in.VoicePath)
	}
	if len(in.Scenes) == 0 {
		return Output{}, fmt.Errorf("لا توجد مشاهد")
	}

	workDir := fmt.Sprintf("output/video_%d", in.VideoID)
	_ = os.MkdirAll(workDir, 0755)
	_ = os.MkdirAll("output", 0755)

	// ── 1️⃣ مدة الصوت الكلي
	voiceDur, err := audioDuration(in.VoicePath)
	if err != nil {
		voiceDur = float64(len(in.Scenes)) * 8.0 // تقدير احتياطي
	}
	fmt.Printf("   🎬 RENDER #%d: %d مشاهد | صوت %.0f ثانية\n",
		in.VideoID, len(in.Scenes), voiceDur)

	// ── 2️⃣ فيديو الصور المتتالية (slideshow) بنفس مدة الصوت
	concatPath, err := buildSlideshow(in.Scenes, voiceDur, workDir)
	if err != nil {
		return Output{}, fmt.Errorf("slideshow: %w", err)
	}

	// ── 3️⃣ دمج الصوت مع الصور (بدون موسيقى — مؤقتاً)
	avPath := filepath.Join(workDir, "av.m4a")
	silentVideo := filepath.Join(workDir, "silent.mp4")
	if err := muxAV(concatPath, in.VoicePath, silentVideo, avPath); err != nil {
		return Output{}, fmt.Errorf("mux: %w", err)
	}

	// ── 4️⃣ موسيقى خلفية هادئة 15%
	musicTrack := music.Pick(in.MusicEmotion)
	finalPath := fmt.Sprintf("output/final_%d.mp4", in.VideoID)

	if musicTrack != "" {
		mixedAudio := filepath.Join(workDir, "mixed.m4a")
		if err := music.MixWithVoice(avPath, musicTrack, mixedAudio, voiceDur); err == nil {
			if err := muxAV(concatPath, mixedAudio, silentVideo, finalPath); err == nil {
				fmt.Printf("   🎬 RENDERED (with music): %s\n", finalPath)
				return Output{VideoPath: finalPath, ThumbPath: in.ThumbPath, Duration: voiceDur}, nil
			}
		}
		// فشلت الموسيقى → النسخة بدونها
	}

	if err := copyFile(silentVideo, finalPath); err != nil {
		return Output{}, fmt.Errorf("finalize: %w", err)
	}
	fmt.Printf("   🎬 RENDERED: %s\n", finalPath)
	return Output{VideoPath: finalPath, ThumbPath: in.ThumbPath, Duration: voiceDur}, nil
}

// ══════════════════════════════════════════
// 🖼️ buildSlideshow — صور المشاهد → فيديو متدرج
// ══════════════════════════════════════════

func buildSlideshow(scenes []Scene, totalDur float64, workDir string) (string, error) {

	// توزيع المدة بالتساوي لو المدد فارغة
	per := totalDur / float64(len(scenes))
	if per < 3 {
		per = 3
	}

	// تحويل كل صورة لمقطع بنفس القياس
	var parts []string
	for i, sc := range scenes {
		if sc.ImagePath == "" || !fileExists(sc.ImagePath) {
			continue
		}
		partPath := filepath.Join(workDir, fmt.Sprintf("part_%02d.mp4", i))
		dur := sc.Duration
		if dur <= 0 {
			dur = per
		}
		cmd := exec.Command("ffmpeg", "-y",
			"-loop", "1", "-i", sc.ImagePath,
			"-t", fmt.Sprintf("%.2f", dur),
			"-vf", "scale=1280:720:force_original_aspect_ratio=decrease,pad=1280:720:(ow-iw)/2:(oh-ih)/2,fps=25",
			"-c:v", "libx264", "-pix_fmt", "yuv420p", "-preset", "fast",
			partPath)
		if b, err := cmd.CombinedOutput(); err != nil {
			return "", fmt.Errorf("part %d: %v: %s", i, err, lastLine(string(b)))
		}
		parts = append(parts, partPath)
	}

	if len(parts) == 0 {
		return "", fmt.Errorf("لا توجد صور صالحة للمشاهد")
	}

	// دمج المقاطع concat
	listPath := filepath.Join(workDir, "parts.txt")
	lf, _ := os.Create(listPath)
	for _, p := range parts {
		lf.WriteString("file '" + p + "'\n")
	}
	lf.Close()

	concatPath := filepath.Join(workDir, "slideshow.mp4")
	cmd := exec.Command("ffmpeg", "-y",
		"-f", "concat", "-safe", "0", "-i", listPath,
		"-c", "copy", concatPath)
	if b, err := cmd.CombinedOutput(); err != nil {
		return "", fmt.Errorf("concat: %v: %s", err, lastLine(string(b)))
	}
	return concatPath, nil
}

// ══════════════════════════════════════════
// 🔊 muxAV — صور + صوت → MP4 نهائي
// ══════════════════════════════════════════

func muxAV(videoPath, audioPath, silentOut, finalOut string) error {

	// فيديو صامت (صور فقط) — يُنتج مرة
	if !fileExists(silentOut) {
		cmd := exec.Command("ffmpeg", "-y",
			"-i", videoPath, "-an",
			"-c:v", "copy", silentOut)
		if b, err := cmd.CombinedOutput(); err != nil {
			return fmt.Errorf("%v: %s", err, lastLine(string(b)))
		}
	}

	// دمج الصوت
	cmd := exec.Command("ffmpeg", "-y",
		"-i", silentOut, "-i", audioPath,
		"-map", "0:v", "-map", "1:a",
		"-c:v", "copy", "-c:a", "aac", "-b:a", "192k",
		"-shortest", finalOut)
	if b, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("%v: %s", err, lastLine(string(b)))
	}
	return nil
}

// ══════════════════════════════════════════
// helpers
// ══════════════════════════════════════════

// audioDuration: قراءة مدة الصوت عبر ffprobe
func audioDuration(path string) (float64, error) {
	out, err := exec.Command("ffprobe",
		"-v", "error", "-show_entries", "format=duration",
		"-of", "default=noprint_wrappers=1:nokey=1", path).Output()
	if err != nil {
		return 0, err
	}
	return strconv.ParseFloat(strings.TrimSpace(string(out)), 64)
}

func fileExists(p string) bool {
	info, err := os.Stat(p)
	return err == nil && !info.IsDir()
}

func copyFile(src, dst string) error {
	b, err := os.ReadFile(src)
	if err != nil {
		return err
	}
	return os.WriteFile(dst, b, 0644)
}

func lastLine(s string) string {
	lines := strings.Split(strings.TrimSpace(s), "\n")
	if len(lines) == 0 {
		return ""
	}
	return lines[len(lines)-1]
}

// LastVideoPath: آخر فيديو نهائي جاهز
func LastVideoPath() string {
	matches, _ := filepath.Glob("output/final_*.mp4")
	if len(matches) > 0 {
		return matches[len(matches)-1]
	}
	return ""
}

// _ استيراد thumbs لضمان التوافق (لو لم يُستخدم احذف هذا السطر والاستيراد)
var _ = thumbs.SmartGenerate
