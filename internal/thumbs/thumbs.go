package thumbs

import (
	"fmt"
	"os"
	"os/exec"
)

// Generate تنشئ صورة مصغرة احترافية 1280x720 بنص كبير
func Generate(id int, text string, bg string) error {
	if err := os.MkdirAll("thumbs", 0o755); err != nil {
		return err
	}
	out := fmt.Sprintf("thumbs/thumb_%d.jpg", id)

	// إن لم توجد خلفية، نستخدم لونًا متدرجًا
	background := bg
	if _, err := os.Stat(bg); err != nil {
		background = "xc:#1a1200"
	}

	args := []string{
		background,
		"-resize", "1280x720^",
		"-gravity", "center",
		"-extent", "1280x720",
		// طبقة تعتيم لوضوح النص
		"-fill", "rgba(0,0,0,0.55)",
		"-draw", "rectangle 0,400 1280,720",
		// النص: أبيض عريض بحد أحمر
		"-font", "DejaVu-Sans-Bold",
		"-pointsize", "72",
		"-stroke", "#ff0000", "-strokewidth", "3",
		"-fill", "#ffffff",
		"-gravity", "south",
		"-annotate", "+0+60", text,
		out,
	}

	cmd := exec.Command("convert", args...)
	cmd.Stderr = os.Stderr
	fmt.Printf("🖼️  THUMB %d → %s\n", id, out)
	return cmd.Run()
}
