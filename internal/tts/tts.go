package tts

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// ══════════════════════════════════════════════════════
// 🎙️ نظام الصوت الثلاثي — الأولوية الجديدة:
//    1️⃣ kokoro-82M (إنجليزي — بصوت af_heart)
//    2️⃣ piper أول أولاً لكل اللغات (محلي — لا إنترنت — لا 403!)
//    3️⃣ edge-tts أحدث نسخة — احتياط أخير فقط
// ══════════════════════════════════════════════════════

// ─── 🔌 الدوال التي يستدعيها main.go — توقيعات مطابقة 100% ───

// Narrate: يولّد الصوت الرئيسي للقصة — (النص، اللغة، مسار الإخراج)
func Narrate(text, lang, outPath string) error {
	return Generate(text, lang, outPath)
}

// DubAllLanguages: يولّد دبلجة القصة لكل اللغات —
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
	return dubs
}

// ══════════════════════════════════════════════════════
// 🎛️ المحرك الرئيسي — Generate
// ══════════════════════════════════════════════════════

// Generate: piper أولاً لكل اللغات (محلي لا 403) — ثم edge-tts احتياطاً
func Generate(text, lang, outPath string) error {
	os.MkdirAll(filepath.Dir(outPath), 0755)

	// 1️⃣ kokoro — للإنجليزية فقط (الصوت المميز)
	if strings.HasPrefix(lang, "en") {
		if err := kokoroTTS(text, outPath); err == nil {
			fmt.Printf("   🥇 kokoro-82M [en] ⭐ voice=af_heart\n")
			return nil
		}
		fmt.Println("   ⚠️ kokoro فشل → piper")
	}

	// 2️⃣ piper — الأولوية الأولى لكل اللغات (محلي، لا إنترنت، لا 403!)
	if err := piperTTS(text, lang, outPath); err == nil {
		fmt.Printf("   🔵 piper [%s] ⭐\n", lang)
		return nil
	}
	fmt.Printf("   ⚠️ piper [%s] فشل → edge-tts\n", lang)

	// 3️⃣ edge-tts أحدث نسخة — شبكة الأمان الأخيرة فقط
	if err := edgeTTS(text, lang, outPath); err != nil {
		return fmt.Errorf("❌ كل أنظمة الصوت فشلت [%s]: %v", lang, err)
	}
	fmt.Printf("   🌐 edge-tts [%s]\n", lang)
	return nil
}

// ─── 1️⃣ kokoro بصوت محدد ───
func kokoroTTS(text, outPath string) error {
	py := fmt.Sprintf(`
from kokoro import KPipeline
import soundfile as sf
p = KPipeline(lang_code='a')
chunks = []
for _, _, audio in p(%q, voice='af_heart', speed=1.0):
    chunks.append(audio)
sf.write(%q, sum(chunks), 24000)
print("ok")
`, text, outPath)
	cmd := exec.Command("python3", "-c", py)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("kokoro: %v — %s", err, strings.TrimSpace(string(out)))
	}
	return fileOK(outPath)
}

// ─── 2️⃣ piper — موديلات onnx المحلية ───
var piperModels = map[string]string{
	"ar": "models/ar_JO-kareem-medium",
	"es": "models/es_ES-carlfm-x-low",
	"tr": "models/tr_TR-fahrettin-medium",
	"fr": "models/fr_FR-siwis-low",
	"de": "models/de_DE-eva_k-x-low",
}

func piperTTS(text, lang, outPath string) error {
	model, ok := piperModels[lang]
	if !ok {
		return fmt.Errorf("لا موديل piper لـ %s", lang)
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
	cmd := exec.Command(piper,
		"--model", model+".onnx",
		"--output_file", outPath)
	cmd.Stdin, _ = os.Open(in.Name())
	cmd.Env = append(os.Environ(), "LD_LIBRARY_PATH=/opt/piper")
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("piper: %v — %s", err, strings.TrimSpace(string(out)))
	}
	return fileOK(outPath)
}

// ─── 3️⃣ edge-tts — شبكة الأمان الأخيرة ───
var edgeVoices = map[string]string{
	"ar": "ar-SA-HamedNeural",
	"en": "en-US-GuyNeural",
	"es": "es-ES-AlvaroNeural",
	"tr": "tr-TR-AhmetNeural",
	"fr": "fr-FR-HenriNeural",
	"de": "de-DE-ConradNeural",
}

func edgeTTS(text, lang, outPath string) error {
	voice, ok := edgeVoices[lang]
	if !ok {
		voice = "en-US-GuyNeural"
	}
	cmd := exec.Command("edge-tts", "--voice", voice,
		"--text", text, "--write-media", outPath)
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("edge-tts: %v — %s", err, strings.TrimSpace(string(out)))
	}
	return fileOK(outPath)
}

// fileOK: تحقق أن الملف الصوتي أُنشئ فعلاً وحجمه معقول
func fileOK(p string) error {
	st, err := os.Stat(p)
	if err != nil {
		return err
	}
	if st.Size() < 10000 {
		return fmt.Errorf("ملف صوتي صغير جداً (%d bytes)", st.Size())
	}
	return nil
}
