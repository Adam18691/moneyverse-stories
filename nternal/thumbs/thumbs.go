package thumbs

import (
	"fmt"
	"os"
	"os/exec"
)

// Generate: ثامبنيل CTR عالي — ذهبي + تباين قوي + نص ضخم
func Generate(videoID int, hookText, sceneImage string) error {
	os.MkdirAll("thumbs", 0755)
	out := fmt.Sprintf("thumbs/thumb_%d.jpg", videoID)
	bg := out + ".bg.jpg"
	defer os.Remove(bg)

	grad := exec.Command("magick", sceneImage,
		"-resize", "1280x720^", "-gravity", "center", "-extent", "1280x720",
		"-modulate", "105,120",
		"-level", "5%,95%",
		bg)
	if err := grad.Run(); err != nil {
		return err
	}

	text := exec.Command("magick", bg,
		"-gravity", "west",
		"-font", "DejaVu-Sans-Bold",
		"-pointsize", "72", "-fill", "#FFD700",
		"-stroke", "#000000", "-strokewidth", "4",
		"-annotate", "+30+80", hookText,
		"-quality", "95", out)
	return text.Run()
}
