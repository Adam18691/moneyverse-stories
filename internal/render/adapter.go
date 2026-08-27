package render

import (
	"fmt"
	"os"
	"os/exec"
)

// Build: توافق مع main.go — يلف RenderVideo
// story + voice → فيديو نهائي في outPath
func Build(id int, story, voicePath, outPath string) error {

	img := "assets/burning_money.jpg"
	if !fileExists(img) {
		// إنشاء صورة سوداء احتياطية عبر ffmpeg
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
		Scenes:    []Scene{{ImagePath: img, Text: "قصة", Duration: 0}},
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
