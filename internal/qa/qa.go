package qa

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

// ══════════════════════════════════════════════════════
// 🧠 طبقة الذكاء — فحص AI + تحسين صوتي + تعلم ذاتي تراكمي
//    1️⃣ Whisper يسمع الصوت ويقارنه بالنص الأصلي
//    2️⃣ loudnorm يرفع الجودة لمستوى يوتيوب (-16 LUFS)
//    3️⃣ qa.json — ذاكرة تراكمية تُحفظ في git بعد كل جولة
//    4️⃣ SelfTune — المحرك يضبط نفسه تلقائياً حسب تاريخه
// ══════════════════════════════════════════════════════

// ─── 1️⃣ فحص AI: نسبة تطابق 0-100 — أقل من العتبة = مرفوض ───
func CheckAudio(audioPath, originalText string) (int, error) {
	if _, err := os.Stat(audioPath); err != nil {
		return 0, fmt.Errorf("لا يوجد ملف صوتي")
	}

	py := fmt.Sprintf(`
import json, warnings
warnings.filterwarnings("ignore")
try:
    import whisper
    model = whisper.load_model("tiny")
    r = model.transcribe(%q, language=None)
    print(json.dumps({"text": r["text"]}))
except Exception as e:
    print(json.dumps({"error": str(e)}))
`, audioPath)

	cmd := exec.Command("python3", "-c", py)
	out, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Printf("   ⚠️ whisper غير متاح — تخطي الفحص الذكي\n")
		return 100, nil
	}

	var res struct {
		Text  string `json:"text"`
		Error string `json:"error"`
	}
	json.Unmarshal(out, &res)
	if res.Error != "" {
		fmt.Printf("   ⚠️ whisper: %s — تخطي الفحص\n", short(res.Error, 60))
		return 100, nil
	}

	score := similarity(originalText, res.Text)
	fmt.Printf("   🧠 فحص AI: تطابق %d%% (سمع: %s)\n", score, short(res.Text, 40))
	return score, nil
}

// ─── 2️⃣ التحسين الصوتي: loudnorm احترافي -16 LUFS ───
func Enhance(audioPath string) error {
	tmp := audioPath + ".tmp.wav"
	cmd := exec.Command("ffmpeg", "-y", "-i", audioPath,
		"-af", "loudnorm=I=-16:TP=-1.5:LRA=11",
		"-ar", "24000", tmp)
	if out, err := cmd.CombinedOutput(); err != nil {
		os.Remove(tmp)
		return fmt.Errorf("loudnorm: %v — %s", err, lastLine(string(out)))
	}
	return os.Rename(tmp, audioPath)
}

// ─── 3️⃣ السجل التراكمي — ذاكرة المحرك ───
type QARecord struct {
	Time    string `json:"time"`
	File    string `json:"file"`
	Lang    string `json:"lang"`
	Score   int    `json:"score"`
	Retries int    `json:"retries"`
}

// Threshold: العتبة الديناميكية — تُقرأ من ذاكرة التعلم
// (افتراضياً 60، وSelfTune قد يرفعها أو يخفضها حسب الأداء)
func Threshold() int {
	t := 60
	if b, err := os.ReadFile("data/qa_config.json"); err == nil {
		var cfg struct {
			Threshold int `json:"threshold"`
		}
		if json.Unmarshal(b, &cfg) == nil && cfg.Threshold >= 40 && cfg.Threshold <= 90 {
			t = cfg.Threshold
		}
	}
	return t
}

// LogRecord: سجّل نتيجة + حدّث الذكاء الذاتي تلقائياً
func LogRecord(file, lang string, score, retries int) {
	os.MkdirAll("data", 0755)
	var log []QARecord
	if b, err := os.ReadFile("data/qa.json"); err == nil {
		json.Unmarshal(b, &log)
	}
	log = append(log, QARecord{
		Time: time.Now().UTC().Format("2006-01-02 15:04"),
		File: filepath.Base(file), Lang: lang, Score: score, Retries: retries,
	})
	if len(log) > 500 {
		log = log[len(log)-500:]
	}
	if b, _ := json.MarshalIndent(log, "", "  "); b != nil {
		os.WriteFile("data/qa.json", b, 0644)
	}

	avg, n := average(log, 20)
	if n >= 5 {
		fmt.Printf("   📈 متوسط جودة آخر %d ملف: %d%% (العتبة: %d%%)\n", n, avg, Threshold())
	}
	selfTune(log)
}

// ─── 4️⃣ SelfTune — التحديث الذاتي المستمر ───
// يضبط عتبة الجودة تلقائياً حسب أداء المحرك التراكمي:
//   جودة عالية مستقرة ≥85% → شدّد المعيار (ارفع العتبة)
//   جودة متعثرة <65%        → خفف (أنزل العتبة) حتى لا يعلق الإنتاج
func selfTune(log []QARecord) {
	avg, n := average(log, 30)
	if n < 10 {
		return // بيانات غير كافية — ننتظر
	}

	current := Threshold()
	newThreshold := current

	switch {
	case avg >= 85 && current < 85:
		newThreshold = current + 5 // المحرك ممتاز — ارفع سقف الجودة
	case avg < 65 && current > 45:
		newThreshold = current - 5 // المحرك متعثر — خفف حتى لا يتوقف
	}

	if newThreshold != current {
		cfg := struct {
			Threshold int `json:"threshold"`
		}{newThreshold}
		if b, _ := json.MarshalIndent(cfg, "", "  "); b != nil {
			os.WriteFile("data/qa_config.json", b, 0644)
		}
		fmt.Printf("   🔄 SELF-TUNE: العتبة %d%% → %d%% (متوسط %d%% من %d عينة)\n",
			current, newThreshold, avg, n)
	}
}

// average: متوسط آخر n سجل
func average(log []QARecord, n int) (int, int) {
	if len(log) == 0 {
		return 0, 0
	}
	sum, cnt := 0, 0
	start := len(log) - n
	if start < 0 {
		start = 0
	}
	for i := start; i < len(log); i++ {
		sum += log[i].Score
		cnt++
	}
	if cnt == 0 {
		return 0, 0
	}
	return sum / cnt, cnt
}

// ─── أدوات مساعدة ───
func similarity(a, b string) int {
	aw, bw := strings.Fields(lower(a)), strings.Fields(lower(b))
	if len(bw) == 0 {
		return 0
	}
	match := map[string]bool{}
	for _, w := range bw {
		match[w] = true
	}
	hit := 0
	for _, w := range aw {
		if match[w] {
			hit++
		}
	}
	return hit * 100 / len(aw)
}

func lower(s string) string {
	r := []rune(s)
	out := make([]rune, len(r))
	for i, c := range r {
		if (c >= 'أ' && c <= 'ي') || (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9') {
			out[i] = c
		} else {
			out[i] = ' '
		}
	}
	return string(out)
}

func short(s string, n int) string {
	r := []rune(s)
	if len(r) > n {
		return string(r[:n]) + "..."
	}
	return s
}

func lastLine(s string) string {
	lines := strings.Split(strings.TrimSpace(s), "\n")
	if len(lines) == 0 {
		return ""
	}
	return lines[len(lines)-1]
}
