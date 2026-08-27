package render

import (
	"fmt"
	"os"
	"os/exec"
)

// Build: توافق مرن مع main.go — يقبل:
//   Build(story, voicePath, outPath)          ← 3 نصوص
//   Build(id, story, voicePath, outPath)      ← int + 3 نصوص
func Build(args ...interface{}) error {

	// تحليل المدخلات المرنة
	id := 0
	var story, voicePath, outPath string

	strs := []string{}
	for _, a := range args {
		switch v := a.(type) {
		case int:
			id = v
		case string:
			strs = append(strs, v)
		}
	}

	switch len(strs) {
	case 3: // story, voice, out
		story, voicePath, outPath = strs[0], strs[1], strs[2]
	case 2: // voice, out
		voicePath, outPath = strs[0], strs[1]
	default:
		return fmt.Errorf("render.Build: مدخلات غير متوقعة (%d)", len(args))
	}

	if voicePath == "" || !fileExists(voicePath) {
		return fmt.Errorf("ملف الصوت غير موجود: %s", voicePath)
	}

	// صورة المشهد — احتياطية إن لم توجد
	img := "assets/burning_money.jpg"
	if !fileExists(img) {
		_ = os.MkdirAll("assets", 0755)
		cmd := exec.Command("ffmpeg", "-y",
			"-f", "lavfi", "-i", "color=c=black:s=1280x720",
			"-frames:v", "1", img)
		if b, err := cmd.CombinedOutput(); err != nil {
			return fmt.Errorf("fallback image: %v: %s", err, lastLine(string(b)))
		}
	}

	in := Input{
		VideoID:   id,
		Scenes:    []Scene{{ImagePath: img, Text: cutRunes(story, 30), Duration: 0}},
		VoicePath: voicePath,
	}

	out, err := RenderVideo(in)
	if err != nil {
		return err
	}
	if out.VideoPath != outPath {
		if err := copyFile(out.VideoPath, outPath); err != nil {
			return err
		}
	}
	fmt.Printf("   🎬 BUILD: %s\n", outPath)
	return nil
}

// cutRunes: قص آمن يحترم الحروف العربية
func cutRunes(s string, n int) string {
	r := []rune(s)
	if len(r) > n {
		return string(r[:n])
	}
	return s
}
