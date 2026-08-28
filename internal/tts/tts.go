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
// 🎙️ محرك piper الموحد + طبقة الذكاء:
//    توليد → تحسين loudnorm → فحص whisper → إعادة تلقائية
//    العتبة ديناميكية — SelfTune يضبطها تلقائياً مع الوقت
// ══════════════════════════════════════════════════════

// ─── 🔌 الدوال التي يستدعيها main.go ───

// Narrate: يولّد الصوت الرئيسي للقصة
func Narrate(text, lang, outPath string) error {
	return Generate(text, lang, outPath)
}

// DubAllLanguages: دبلجة القصة لكل اللغات
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

// ─── الموديلات المحلية ───
var piperModels = map[string]string{
	"ar": "models/ar_JO-kareem-medium",
	"en": "models/en_US-ryan-high",
	"es": "models/es_ES-carlfm-x-low",
	"tr": "models/tr_TR-fahrettin-medium",
	"fr": "models/fr_FR-siwis-low",
	"de": "models/de_DE-eva_k-x-low",
}

// ══════════════════════════════════════════════════════
// 🎛️ Generate — توليد + تحسين + فحص AI + إعادة تلقائية
// ══════════════════════════════════════════════════════
func Generate(text, lang, outPath string) error {
	os.MkdirAll(filepath.Dir(outPath), 0755)

	threshold := qa.Threshold() // عتبة ديناميكية من ذاكرة التعلم
	const maxRetries = 2

	for attempt := 1; attempt <= maxRetries; attempt++ {
		if err := piperGenerate(text, lang, outPath); err != nil {
			fmt.Printf("   ⚠️ piper [%s] فشل: %v\n", lang, err)
			continue
		}

		// ✨ التحسين الصوتي — مستوى يوتيوب الاحترافي
		if err := qa.Enhance(outPath); err != nil {
			fmt.Printf("   ⚠️ تحسين صوتي فشل: %v — الأصل كافٍ\n", err)
		} else {
			fmt.Printf("   ✨ enhanced -16 LUFS\n")
		}

		// 🧠 فحص AI — هل الصوت فعلاً يقول النص؟
		score, _ := qa.CheckAudio(outPath, text)
		qa.LogRecord(outPath, lang, score, attempt-1)

		if score >= threshold {
			return nil // ✅ صوت سليم ومحسّن ومعتمد
		}
		fmt.Printf("   🔄 جودة منخفضة (%d%% < %d%%) — إعادة توليد %d/%d\n",
			score, threshold, attempt, maxRetries)
	}
	return fmt.Errorf("piper [%s]: فشل فحص AI بعد %d محاولات", lang, maxRetries)
}

// piperGenerate: توليد piper الخام
func piperGenerate(text, lang, outPath string) error {
	model, ok := piperModels[lang]
	if !ok {
		return fmt.Errorf("لا موديل piper للغة %s", lang)
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
		return fmt.Errorf("%v — %s", err, strings.TrimSpace(string(out)))
	}
	return fileOK(outPath)
}

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
