package tts

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"tayyibat-money/internal/qa"
)

// ══════════════════════════════════════════════════════
// 🎙️ المحرك v2.0 النهائي — أقصى استغلال للـ RAM المجاني:
//    🐱 KittenTTS 25MB   — الإنجليزية الأولى (150MB)
//    🥇 Kokoro ONNX      — استوديو بحجم 88MB (300MB RAM)
//    🔵 Piper ×12 لغة    — الأساس (150-300MB)
//    🎭 Bark small       — ذخيرة درامية (بوابة ≥3GB)
//    💾 GC بعد كل فيديو + وضع طوارئ piper-only
// ══════════════════════════════════════════════════════

// ─── 🔌 الواجهة لـ main.go ───

func Narrate(text, lang, outPath string) error {
	return Generate(text, lang, outPath)
}

// NarrateDramatic: للمقدمة/الذروة — bark فقط عند RAM كافية
func NarrateDramatic(text, lang, outPath string) error {
	os.MkdirAll(filepath.Dir(outPath), 0755)

	if qa.RamFreeMB() >= 3000 {
		if err := barkTTS(text, lang, outPath); err == nil {
			if err := qa.Enhance(outPath); err == nil {
				fmt.Printf("   🎭 bark dramatic ⭐ %s\n", qa.RamStatus())
				qa.FreeMemory()
				return nil
			}
		}
		fmt.Println("   ⚠️ bark فشل → السلسلة العادية")
	} else {
		fmt.Printf("   💾 RAM %dMB < 3000 — تخطي bark (حماية)\n", qa.RamFreeMB())
	}
	return Generate(text, lang, outPath)
}

func DubAllLanguages(langs map[string]string, id int) map[string]string {
	dubs := make(map[string]string)
	for lang, script := range langs {
		path := fmt.Sprintf("audio/%d_%s.wav", id, lang)
		if err := Generate(script, lang, path); err != nil {
			fmt.Printf("   ⚠️ dub [%s] فشل — تخطي: %v\n", lang, err)
			continue
		}
		dubs[lang] = path
	}
	qa.FreeMemory()
	return dubs
}

// ─── 🔵 Piper — 12 لغة (الموجود يعمل، الغائب يقفز تلقائياً) ───
var piperModels = map[string]string{
	"ar": "models/ar_JO-kareem-medium",
	"en": "models/en_US-ryan-high",
	"es": "models/es_ES-carlfm-x-low",
	"tr": "models/tr_TR-fahrettin-medium",
	"fr": "models/fr_FR-siwis-low",
	"de": "models/de_DE-eva_k-x-low",
	"ja": "models/ja_JP-thorsten-high",
	"ru": "models/ru_RU-irina-medium",
	"id": "models/id_ID-ardhi-medium",
	"hi": "models/hi_IN-pratham-medium",
	"pt": "models/pt_BR-faber-medium",
	"it": "models/it_IT-riccardo-x_low",
}

var barkVoices = map[string]string{
	"ar": "v2/ar_speaker_4", "en": "v2/en_speaker_6",
	"es": "v2/es_speaker_0", "tr": "v2/tr_speaker_0",
	"fr": "v2/fr_speaker_1", "de": "v2/de_speaker_3",
}

// ══════════════════════════════════════════════════════
// 🎛️ Generate — سلسلة تقرر حسب RAM لحظياً:
//    <1.5GB : piper فقط بلا فحص (وضع الطوارئ)
//    ≥1.5GB : kitten/kokoro → piper → bark
// ══════════════════════════════════════════════════════
func Generate(text, lang, outPath string) error {
	os.MkdirAll(filepath.Dir(outPath), 0755)

	const maxRetries = 2
	for attempt := 1; attempt <= maxRetries; attempt++ {

		// ═══ وضع الطوارئ — صفر مخاطر ═══
		if qa.RamFreeMB() < 1500 {
			fmt.Printf("   💾 وضع الطوارئ %s — piper مباشر\n", qa.RamStatus())
			if err := piperGenerate(text, lang, outPath); err == nil {
				qa.LogRecord(outPath, lang, "piper-lite", 100, attempt-1)
				return nil
			}
			continue
		}

		// 🐱/🥇 الإنجليزية: kitten (أخف) → kokoro-onnx (أجمل)
		if strings.HasPrefix(lang, "en") {
			if err := kittenTTS(text, outPath); err == nil {
				if finalize(outPath, text, lang, "kitten", attempt) {
					return nil
				}
			} else if err := kokoroTTS(text, outPath); err == nil {
				if finalize(outPath, text, lang, "kokoro", attempt) {
					return nil
				}
			}
		}

		// 🔵 piper — الأساس لكل الـ 12 لغة
		if err := piperGenerate(text, lang, outPath); err != nil {
			fmt.Printf("   ⚠️ piper [%s]: %v → bark\n", lang, err)
		} else if finalize(outPath, text, lang, "piper", attempt) {
			return nil
		}

		// 🎭 bark — الذخيرة (بوابة RAM داخل الدالة)
		if err := barkTTS(text, lang, outPath); err == nil {
			if finalize(outPath, text, lang, "bark", attempt) {
				fmt.Println("   🎭 bark أنقذ الجولة!")
				return nil
			}
		}
	}
	return fmt.Errorf("كل المحركات فشلت [%s]", lang)
}

func finalize(outPath, text, lang, engine string, attempt int) bool {
	if err := qa.Enhance(outPath); err != nil {
		fmt.Printf("   ⚠️ تحسين: %v\n", err)
	} else {
		fmt.Printf("   ✨ enhanced -16 LUFS %s\n", qa.RamStatus())
	}
	score, _ := qa.CheckAudio(outPath, text)
	qa.LogRecord(outPath, lang, engine, score, attempt-1)
	return score >= qa.Threshold()
}

// ─── 🐱 KittenTTS — 25MB ───
func kittenTTS(text, outPath string) error {
	py := fmt.Sprintf(`
import warnings; warnings.filterwarnings("ignore")
from kittentts import KittenTTS
m = KittenTTS("KittenML/kitten-tts-nano-0.1")
audio = m.generate(%q)
import soundfile as sf
sf.write(%q, audio, 24000)
`, text, outPath)
	if out, err := exec.Command("python3", "-c", py).CombinedOutput(); err != nil {
		return fmt.Errorf("kitten: %s", lastLine(string(out)))
	}
	return fileOK(outPath)
}

// ─── 🥇 Kokoro ONNX — 88MB (بدل 300MB PyTorch) ───
func kokoroTTS(text, outPath string) error {
	py := fmt.Sprintf(`
import warnings; warnings.filterwarnings("ignore")
from kokoro_onnx import Kokoro
import soundfile as sf
k = Kokoro("models/kokoro-v1.0.onnx", "models/kokoro-v1.0-voices.bin")
audio, sr = k.create(%q, voice="af_heart", speed=1.0, lang="en-us")
sf.write(%q, audio, sr)
`, text, outPath)
	if out, err := exec.Command("python3", "-c", py).CombinedOutput(); err != nil {
		return fmt.Errorf("kokoro-onnx: %s", lastLine(string(out)))
	}
	return fileOK(outPath)
}

// ─── 🔵 Piper ───
func piperGenerate(text, lang, outPath string) error {
	model, ok := piperModels[lang]
	if !ok {
		return fmt.Errorf("لا موديل للغة %s", lang)
	}
	if _, err := os.Stat(model + ".onnx"); err != nil {
		return fmt.Errorf("موديل غير موجود: %s", model)
	}
	in, _ := os.CreateTemp("", "*.txt")
	in.WriteString(text)
	in.Close()
	defer os.Remove(in.Name())

	piper := os.Getenv("PIPER_PATH")
	if piper == "" {
		piper = "piper"
	}
	cmd := exec.Command(piper, "--model", model+".onnx", "--output_file", outPath)
	cmd.Stdin, _ = os.Open(in.Name())
	cmd.Env = append(os.Environ(), "LD_LIBRARY_PATH=/opt/piper")
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("%s", lastLine(string(out)))
	}
	return fileOK(outPath)
}

// ─── 🎭 Bark small — بوابة RAM ≥3GB ───
func barkTTS(text, lang, outPath string) error {
	if qa.RamFreeMB() < 3000 {
		return fmt.Errorf("RAM غير كافية لـ bark (%dMB)", qa.RamFreeMB())
	}
	voice, ok := barkVoices[lang]
	if !ok {
		voice = "v2/en_speaker_6"
	}
	py := fmt.Sprintf(`
import warnings; warnings.filterwarnings("ignore")
from bark import SAMPLE_RATE, generate_audio
from scipy.io.wavfile import write as wav_write
import gc
audio = generate_audio(%q, history_prompt=%q)
wav_write(%q, SAMPLE_RATE, audio)
del audio; gc.collect()
`, text, voice, outPath)
	cmd := exec.Command("python3", "-c", py)
	cmd.Env = append(os.Environ(), "SUNO_USE_SMALL_MODELS=1", "SUNO_OFFLOAD_CPU=1")
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("bark: %s", lastLine(string(out)))
	}
	return fileOK(outPath)
}

func fileOK(p string) error {
	st, err := os.Stat(p)
	if err != nil {
		return err
	}
	if st.Size() < 10000 {
		return fmt.Errorf("ملف صغير (%d bytes)", st.Size())
	}
	return nil
}

func lastLine(s string) string {
	lines := strings.Split(strings.TrimSpace(s), "\n")
	if len(lines) == 0 {
		return ""
	}
	return lines[len(lines)-1]
}
