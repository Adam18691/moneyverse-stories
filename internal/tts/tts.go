package tts

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"tayyibat-money/internal/qa"
)

// 🎙️ السلسلة: kitten → kokoro → piper → bark (ببوابات RAM)

func Narrate(text, lang, outPath string) error {
	return Generate(text, lang, outPath)
}

func NarrateDramatic(text, lang, outPath string) error {
	_ = os.MkdirAll(filepath.Dir(outPath), 0755)

	if qa.RamFreeMB() >= 3000 {
		if err := barkTTS(text, lang, outPath); err == nil {
			if err := qa.Enhance(outPath); err == nil {
				fmt.Printf("   bark dramatic | %s\n", qa.RamStatus())
				qa.FreeMemory()
				return nil
			}
		}
	}
	return Generate(text, lang, outPath)
}

func DubAllLanguages(langs map[string]string, id int) map[string]string {
	_ = os.MkdirAll("audio", 0755)

	dubs := make(map[string]string)
	for lang, script := range langs {
		path := fmt.Sprintf("audio/%d_%s.wav", id, lang)
		if err := Generate(script, lang, path); err != nil {
			fmt.Printf("   dub %s فشل — تخطي\n", lang)
			continue
		}
		dubs[lang] = path
	}
	qa.FreeMemory()
	return dubs
}

// 8 لغات مؤكدة — أسماء قصيرة مطابقة لليامل
var piperModels = map[string]string{
	"ar": "models/ar",
	"en": "models/en",
	"es": "models/es",
	"tr": "models/tr",
	"fr": "models/fr",
	"de": "models/de",
	"pt": "models/pt",
	"it": "models/it",
}

var barkVoices = map[string]string{
	"ar": "v2/ar_speaker_4",
	"en": "v2/en_speaker_6",
	"es": "v2/es_speaker_0",
	"tr": "v2/tr_speaker_0",
	"fr": "v2/fr_speaker_1",
	"de": "v2/de_speaker_3",
}

func Generate(text, lang, outPath string) error {
	_ = os.MkdirAll(filepath.Dir(outPath), 0755)

	for attempt := 1; attempt <= 2; attempt++ {
		if qa.RamFreeMB() < 1500 {
			fmt.Printf("   وضع الطوارئ | %s — piper مباشر\n", qa.RamStatus())
			if err := piperGenerate(text, lang, outPath); err == nil {
				qa.LogRecord(outPath, lang, "piper-lite", 100, attempt-1)
				return nil
			}
			continue
		}

		if strings.HasPrefix(lang, "en") {
			if err := kittenTTS(text, outPath); err == nil {
				if finalize(outPath, text, lang, "kitten", attempt) {
					return nil
				}
			}
			if err := kokoroTTS(text, outPath); err == nil {
				if finalize(outPath, text, lang, "kokoro", attempt) {
					return nil
				}
			}
		}

		if err := piperGenerate(text, lang, outPath); err == nil {
			if finalize(outPath, text, lang, "piper", attempt) {
				return nil
			}
		} else {
			fmt.Printf("   piper %s فشل → bark\n", lang)
		}

		if err := barkTTS(text, lang, outPath); err == nil {
			if finalize(outPath, text, lang, "bark", attempt) {
				return nil
			}
		}
	}

	return fmt.Errorf("كل المحركات فشلت %s", lang)
}

// finalize v2.2 — يتحقق من صحة الملف قبل التحسين، ويرفض التالف فوراً
func finalize(outPath, text, lang, engine string, attempt int) bool {
	st, err := os.Stat(outPath)
	if err != nil || st.Size() < 10000 {
		fmt.Printf("   ملف تالف/صغير — رفض [%s]\n", lang)
		_ = os.Remove(outPath)
		return false
	}

	chk, _ := exec.Command("ffprobe", "-v", "error", outPath).CombinedOutput()
	if len(chk) > 0 {
		fmt.Printf("   ملف غير صالح — رفض [%s]: %s\n", lang, lastLine(string(chk)))
		_ = os.Remove(outPath)
		return false
	}

	if err := qa.Enhance(outPath); err != nil {
		fmt.Printf("   تحسين فشل (الملف الأصلي محفوظ): %v\n", err)
	} else {
		fmt.Printf("   enhanced | %s\n", qa.RamStatus())
	}

	score, _ := qa.CheckAudio(outPath, text)
	qa.LogRecord(outPath, lang, engine, score, attempt-1)
	return score >= qa.Threshold()
}

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

func kokoroTTS(text, outPath string) error {
	if _, err := os.Stat("models/kokoro.onnx"); err != nil {
		return fmt.Errorf("kokoro غير مثبت")
	}
	py := fmt.Sprintf(`
import warnings; warnings.filterwarnings("ignore")
from kokoro_onnx import Kokoro
import soundfile as sf
k = Kokoro("models/kokoro.onnx", "models/kokoro-voices.bin")
audio, sr = k.create(%q, voice="af_heart", speed=1.0, lang="en-us")
sf.write(%q, audio, sr)
`, text, outPath)
	if out, err := exec.Command("python3", "-c", py).CombinedOutput(); err != nil {
		return fmt.Errorf("kokoro: %s", lastLine(string(out)))
	}
	return fileOK(outPath)
}

func piperGenerate(text, lang, outPath string) error {
	model, ok := piperModels[lang]
	if !ok {
		return fmt.Errorf("لا موديل للغة %s", lang)
	}
	if _, err := os.Stat(model + ".onnx"); err != nil {
		return fmt.Errorf("موديل غير موجود: %s", model)
	}

	in, err := os.CreateTemp("", "*.txt")
	if err != nil {
		return err
	}
	defer os.Remove(in.Name())

	if _, err := in.WriteString(text); err != nil {
		_ = in.Close()
		return err
	}
	_ = in.Close()

	piper := os.Getenv("PIPER_PATH")
	if piper == "" {
		piper = "piper"
	}

	f, err := os.Open(in.Name())
	if err != nil {
		return err
	}
	defer f.Close()

	cmd := exec.Command(piper, "--model", model+".onnx", "--output_file", outPath)
	cmd.Stdin = f
	cmd.Env = append(os.Environ(), "LD_LIBRARY_PATH=/opt/piper")

	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("%s", lastLine(string(out)))
	}
	return fileOK(outPath)
}

func barkTTS(text, lang, outPath string) error {
	if qa.RamFreeMB() < 3000 {
		return fmt.Errorf("RAM غير كافية لـ bark")
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
		return fmt.Errorf("ملف صغير")
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
