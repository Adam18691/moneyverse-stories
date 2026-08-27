package tts

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
)

type Voice struct {
	Lang     string
	Model    string
	Speed    float64
	EdgeName string // صوت edge-tts العصبي البديل
	Kokoro   string // كود لغة kokoro ('a' عربي/إنجليزي، 'e' إسباني، ...)
}

var Voices = map[string]Voice{
	"ar": {"arabic",    "models/ar_JO-kareem-medium.onnx",    1.0, "ar-SA-HamedNeural", "a"},
	"en": {"english",   "models/en_US-ryan-high.onnx",        1.0, "en-US-GuyNeural",   "a"},
	"es": {"spanish",   "models/es_ES-carlfm-x-low.onnx",     1.0, "es-ES-AlvaroNeural","e"},
	"fr": {"french",    "models/fr_FR-siwis-low.onnx",        1.0, "fr-FR-HenriNeural", "f"},
	"de": {"german",    "models/de_DE-eva_k-x_low.onnx",      1.0, "de-DE-ConradNeural","d"},
	"tr": {"turkish",   "models/tr_TR-fahrettin-medium.onnx", 1.0, "tr-TR-AhmetNeural", ""},
	"id": {"indonesian","models/id_ID-arif-medium.onnx",      1.0, "id-ID-ArdiNeural",  ""},
	"ur": {"urdu",      "models/ur_PK-umair-medium.onnx",     1.0, "ur-PK-UzmaNeural",  ""},
	"hi": {"hindi",     "models/hi_IN-pratham-medium.onnx",   1.0, "hi-IN-MadhurNeural","h"},
	"zh": {"chinese",   "models/zh_CN-huayan-medium.onnx",    1.0, "zh-CN-YunxiNeural", "z"},
}

// kokoroScript: سكربت بايثون لتوليد الصوت عبر Kokoro-82M
func kokoroScript(text, outFile, langCode string) string {
	return fmt.Sprintf(`
from kokoro import KPipeline
import soundfile as sf
import numpy as np
p = KPipeline(lang_code='%s')
audio = []
for gs, ps, chunk in p(%q):
    audio.append(chunk.numpy())
sf.write(%q, np.concatenate(audio), 24000)
`, langCode, text, outFile)
}

// Narrate: توليد الصوت — 3 مستويات: kokoro 🥇 → piper 🗣️ → edge-tts 🌐
func Narrate(text, lang, outFile string) error {
	v, ok := Voices[lang]
	if !ok {
		v = Voices["en"]
	}

	// ─── 🥇 المستوى 1: Kokoro-82M (الممثل المميز — Apache 2.0 تجاري) ───
	if v.Kokoro != "" {
		cmd := exec.Command("python3", "-c", kokoroScript(text, outFile, v.Kokoro))
		cmd.Env = append(os.Environ(), "PYTHONUNBUFFERED=1")
		if b, err := cmd.CombinedOutput(); err == nil {
			if info, e := os.Stat(outFile); e == nil && info.Size() > 10000 {
				fmt.Printf("   🥇 kokoro-82M [%s] — مميز ⭐\n", lang)
				return nil
			}
		} else {
			fmt.Printf("   ⚠️ kokoro [%s]: %s — نزول لـ piper\n", lang, lastLine(string(b)))
		}
	}

	// ─── 🗣️ المستوى 2: Piper المحلي (إن وُجد الموديل) ───
	if _, err := os.Stat(v.Model); err == nil {
		cmd := exec.Command("piper",
			"--model", v.Model,
			"--length_scale", fmt.Sprintf("%.2f", v.Speed),
			"--output_file", outFile)
		cmd.Stdin = strings.NewReader(text)
		if err := cmd.Run(); err == nil {
			fmt.Printf("   🗣️ piper [%s]\n", lang)
			return nil
		} else {
			fmt.Printf("   ⚠️ piper [%s] فشل: %v — نزول لـ edge-tts\n", lang, err)
		}
	}

	// ─── 🌐 المستوى 3: edge-tts العصبي السحابي (Microsoft) ───
	if v.EdgeName != "" {
		cmd := exec.Command("edge-tts",
			"--voice", v.EdgeName,
			"--text", text,
			"--write-media", outFile)
		if b, err := cmd.CombinedOutput(); err != nil {
			return fmt.Errorf("edge-tts [%s]: %v: %s", lang, err, lastLine(string(b)))
		}
		if info, err := os.Stat(outFile); err == nil && info.Size() > 10000 {
			fmt.Printf("   🌐 edge-tts [%s] %s (عصبي)\n", lang, v.EdgeName)
			return nil
		}
		return fmt.Errorf("edge-tts [%s]: الملف الناتج فارغ", lang)
	}

	return fmt.Errorf("لا يوجد مولد صوت للغة %s", lang)
}

// DubAllLanguages: دبلجة كل اللغات
func DubAllLanguages(scriptPerLang map[string]string, videoID int) map[string]string {
	tracks := make(map[string]string)
	for lang, script := range scriptPerLang {
		out := fmt.Sprintf("audio/%d_%s.wav", videoID, lang)
		if err := Narrate(script, lang, out); err != nil {
			fmt.Printf("🔊 dub %s failed: %v\n", lang, err)
			continue
		}
		tracks[lang] = out
	}
	return tracks
}

// lastLine: آخر سطر من مخرجات الأمر
func lastLine(s string) string {
	s = strings.TrimSpace(s)
	if s == "" {
		return ""
	}
	lines := strings.Split(s, "\n")
	return lines[len(lines)-1]
}
