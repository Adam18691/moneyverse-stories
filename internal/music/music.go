package music

import (
	"fmt"
	"math/rand"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// ══════════════════════════════════════════
// 🎵 MUSIC ENGINE — موسيقى خلفية حرة 100%
// YouTube Audio Library (مرفوعة في assets) أو توليد صامت آمن
// ══════════════════════════════════════════

// مكتبة مقاطع حرة من YouTube Audio Library — ضعها في assets/music/
// (أو سيختار النظام تلقائياً مما وجد)
var library = []string{
	"assets/music/dramatic.mp3",
	"assets/music/suspense.mp3",
	"assets/music/epic.mp3",
	"assets/music/ambient.mp3",
	"assets/music/tension.mp3",
}

// Pick: يختار مقطعاً موسيقياً مناسباً — يعيد المسار أو "" (بدون موسيقى)
func Pick(emotion string) string {
	// فلترة حسب العاطفة لو أمكن
	var candidates []string
	for _, track := range library {
		if fileExists(track) {
			candidates = append(candidates, track)
		}
	}
	if len(candidates) == 0 {
		fmt.Println("   🎵 لا توجد موسيقى في assets/music → بدون موسيقى")
		return ""
	}

	// مطابقة تقريبية بالاسم مع العاطفة
	emotionTracks := map[string][]string{
		"shock":   {"dramatic", "tension"},
		"fear":    {"suspense", "tension"},
		"greed":   {"epic", "ambient"},
		"hope":    {"ambient", "epic"},
	}
	for _, pref := range emotionTracks[emotion] {
		for _, c := range candidates {
			if strings.Contains(c, pref) {
				fmt.Printf("   🎵 MUSIC: %s (%s)\n", c, emotion)
				return c
			}
		}
	}

	// عشوائي من المتوفر
	track := candidates[rand.Intn(len(candidates))]
	fmt.Printf("   🎵 MUSIC: %s\n", track)
	return track
}

// MixWithVoice: يدمج الموسيقى مع الصوت — موسيقى هادئة 15% + صوت رئيسي
// لو musicPath فاضي → يمرر الصوت كما هو
func MixWithVoice(voicePath, musicPath, outPath string, durationSec float64) error {
	if musicPath == "" || !fileExists(musicPath) {
		// بدون موسيقى — نسخ الصوت مباشرة
		cmd := exec.Command("cp", voicePath, outPath)
		if b, err := cmd.CombinedOutput(); err != nil {
			return fmt.Errorf("copy voice: %v: %s", err, string(b))
		}
		return nil
	}

	// ffmpeg: موسيقى loop خافتة 15% تحت الصوت
	fadeOut := ""
	if durationSec > 10 {
		fadeStart := fmt.Sprintf("%.0f", durationSec-3)
		fadeOut = fmt.Sprintf(",afade=t=out:st=%s:d=3", fadeStart)
	}

	args := []string{
		"-y",
		"-i", voicePath,
		"-stream_loop", "-1", "-i", musicPath,
		"-filter_complex", fmt.Sprintf(
			"[1:a]volume=0.15%s[bg];[0:a][bg]amix=inputs=2:duration=first[out]",
			fadeOut),
		"-map", "[out]",
		"-c:a", "aac", "-b:a", "192k",
		outPath,
	}
	cmd := exec.Command("ffmpeg", args...)
	if b, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("mix: %v: %s", err, lastLine(string(b)))
	}
	fmt.Printf("   🎵 MIXED: موسيقى 15%% + صوت → %s\n", outPath)
	return nil
}

// GenerateSilent: يولد مساراً صامتاً — احتياطي أخير لضمان استمرار الرندر
func GenerateSilent(outPath string, durationSec float64) error {
	cmd := exec.Command("ffmpeg", "-y",
		"-f", "lavfi", "-i", fmt.Sprintf("anullsrc=channel_layout=stereo:sample_rate=44100"),
		"-t", fmt.Sprintf("%.1f", durationSec),
		"-c:a", "aac", outPath)
	if b, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("silent: %v: %s", err, string(b))
	}
	return nil
}

// ── helpers ──

func fileExists(p string) bool {
	info, err := os.Stat(p)
	return err == nil && !info.IsDir()
}

func lastLine(s string) string {
	lines := strings.Split(strings.TrimSpace(s), "\n")
	if len(lines) == 0 {
		return ""
	}
	return lines[len(lines)-1]
}

// EnsureAssetsDir: ينشئ مجلد الموسيقى لو مش موجود
func EnsureAssetsDir() {
	_ = os.MkdirAll(filepath.Dir("assets/music/"), 0755)
}
