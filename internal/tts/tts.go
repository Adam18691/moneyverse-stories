package tts

import (
	"fmt"
	"os/exec"
	"strings"
)

type Voice struct {
	Lang  string
	Model string
	Speed float64
}

var Voices = map[string]Voice{
	"ar": {"arabic", "models/ar_JO-kareem-medium.onnx", 1.0},
	"en": {"english", "models/en_US-ryan-high.onnx", 1.0},
	"fr": {"french", "models/fr_FR-siwis-medium.onnx", 1.0},
	"es": {"spanish", "models/es_ES-carlfm-x_low.onnx", 1.0},
	"de": {"german", "models/de_DE-eva_k-x_low.onnx", 1.0},
	"tr": {"turkish", "models/tr_TR-fahrettin-medium.onnx", 1.0},
	"id": {"indonesian", "models/id_ID-arif-medium.onnx", 1.0},
	"ur": {"urdu", "models/ur_PK-umair-medium.onnx", 1.0},
	"hi": {"hindi", "models/hi_IN-pratham-medium.onnx", 1.0},
	"zh": {"chinese", "models/zh_CN-huayan-medium.onnx", 1.0},
}

// Narrate: صوت راوي عميق عبر Piper TTS (مفتوح المصدر)
func Narrate(text, lang, outFile string) error {
	v, ok := Voices[lang]
	if !ok {
		v = Voices["en"]
	}
	cmd := exec.Command("piper",
		"--model", v.Model,
		"--length_scale", fmt.Sprintf("%.2f", v.Speed),
		"--output_file", outFile)
	cmd.Stdin = strings.NewReader(text)
	return cmd.Run()
}

// DubAllLanguages: دبلجة لكل اللغات
func DubAllLanguages(scriptPerLang map[string]string, videoID int) map[string]string {
	tracks := map[string]string{}
	for lang, script := range scriptPerLang {
		out := fmt.Sprintf("audio/%d_%s.wav", videoID, lang)
		if err := Narrate(script, lang, out); err != nil {
			fmt.Printf("⚠️ dub %s failed: %v\n", lang, err)
			continue
		}
		tracks[lang] = out
	}
	return tracks
}
