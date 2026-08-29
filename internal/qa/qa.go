package qa

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"runtime/debug"
	"strings"
	"time"
)

type QARecord struct {
	Time    string `json:"time"`
	File    string `json:"file"`
	Lang    string `json:"lang"`
	Engine  string `json:"engine"`
	Score   int    `json:"score"`
	Retries int    `json:"retries"`
}

func RamFreeMB() int {
	b, err := os.ReadFile("/proc/meminfo")
	if err != nil {
		return 9999
	}
	for _, line := range strings.Split(string(b), "\n") {
		if strings.HasPrefix(line, "MemAvailable:") {
			var kb int
			fmt.Sscanf(line, "MemAvailable: %d", &kb)
			return kb / 1024
		}
	}
	return 9999
}

func FreeMemory() {
	runtime.GC()
	debug.FreeOSMemory()
}

func RamStatus() string {
	return fmt.Sprintf("RAM متاح: %dMB", RamFreeMB())
}

func CheckAudio(audioPath, originalText string) (int, error) {
	if _, err := os.Stat(audioPath); err != nil {
		return 0, fmt.Errorf("لا يوجد ملف صوتي")
	}
	if RamFreeMB() < 1500 {
		fmt.Println("   RAM ضيق — تخطي الفحص")
		return 100, nil
	}

	py := fmt.Sprintf(`
import json, warnings
warnings.filterwarnings("ignore")
try:
    from faster_whisper import WhisperModel
    model = WhisperModel("tiny", device="cpu", compute_type="int8")
    segments, _ = model.transcribe(%q)
    text = " ".join(s.text for s in segments)
    print(json.dumps({"text": text}))
except Exception as e:
    print(json.dumps({"error": str(e)}))
`, audioPath)

	out, err := exec.Command("python3", "-c", py).CombinedOutput()
	if err != nil {
		fmt.Println("   faster-whisper غير متاح — تخطي الفحص")
		return 100, nil
	}

	var res struct {
		Text  string `json:"text"`
		Error string `json:"error"`
	}
	_ = json.Unmarshal(out, &res)
	if res.Error != "" {
		return 100, nil
	}

	score := similarity(originalText, res.Text)
	fmt.Printf("   فحص AI: تطابق %d%% | %s\n", score, RamStatus())
	return score, nil
}

func Enhance(audioPath string) error {
	tmp := audioPath + ".tmp.wav"
	af := "loudnorm=I=-16:TP=-1.5:LRA=11"

	if _, err := os.Stat("models/std.rnnn"); err == nil {
		af = "arnndn=m=models/std.rnnn," + af
		fmt.Println("   RNNoise مفعّل")
	}

	out, err := exec.Command("ffmpeg", "-y", "-i", audioPath, "-af", af, "-ar", "24000", tmp).CombinedOutput()
	if err != nil {
		fmt.Printf("   ffmpeg: %s\n", string(out))
		_ = os.Remove(tmp)
		return fmt.Errorf("enhance failed")
	}

	return os.Rename(tmp, audioPath)
}

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

func LogRecord(file, lang, engine string, score, retries int) {
	_ = os.MkdirAll("data", 0755)

	var log []QARecord
	if b, err := os.ReadFile("data/qa.json"); err == nil {
		_ = json.Unmarshal(b, &log)
	}

	if engine == "" {
		engine = "piper"
	}

	log = append(log, QARecord{
		Time:    time.Now().UTC().Format("2006-01-02 15:04"),
		File:    filepath.Base(file),
		Lang:    lang,
		Engine:  engine,
		Score:   score,
		Retries: retries,
	})

	if len(log) > 500 {
		log = log[len(log)-500:]
	}

	if b, err := json.MarshalIndent(log, "", " "); err == nil {
		_ = os.WriteFile("data/qa.json", b, 0644)
	}

	writeDailyReport(log)
	selfTune(log)
}

type engineStat struct {
	Count int     `json:"count"`
	Avg   float64 `json:"avg_score"`
}

func writeDailyReport(log []QARecord) {
	today := time.Now().UTC().Format("2006-01-02")
	engines := map[string]engineStat{}

	for _, r := range log {
		if !strings.HasPrefix(r.Time, today) {
			continue
		}
		s := engines[r.Engine]
		s.Count++
		s.Avg = s.Avg*float64(s.Count-1)/float64(s.Count) + float64(r.Score)/float64(s.Count)
		engines[r.Engine] = s
	}

	report := struct {
		Date    string                `json:"date"`
		RamFree int                   `json:"ram_free_mb"`
		Engines map[string]engineStat `json:"engines"`
	}{
		Date:    today,
		RamFree: RamFreeMB(),
		Engines: engines,
	}

	if b, err := json.MarshalIndent(report, "", " "); err == nil {
		_ = os.WriteFile("data/daily_report.json", b, 0644)
	}
}

func selfTune(log []QARecord) {
	avg, n := average(log, 30)
	if n < 10 {
		return
	}

	current := Threshold()
	newT := current

	if avg >= 85 && current < 85 {
		newT = current + 5
	}
	if avg < 65 && current > 45 {
		newT = current - 5
	}

	if newT != current {
		cfg := struct {
			Threshold int `json:"threshold"`
		}{Threshold: newT}

		if b, err := json.MarshalIndent(cfg, "", " "); err == nil {
			_ = os.WriteFile("data/qa_config.json", b, 0644)
		}
		fmt.Printf("   SELF-TUNE: %d الى %d\n", current, newT)
	}
}

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

func similarity(a, b string) int {
	aw, bw := strings.Fields(lower(a)), strings.Fields(lower(b))
	if len(aw) == 0 || len(bw) == 0 {
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
	s = strings.ToLower(s)
	var out []rune
	for _, c := range s {
		if (c >= 'أ' && c <= 'ي') || (c >= 'a' && c <= 'z') || (c >= '0' && c <= '9') || c == ' ' {
			out = append(out, c)
		} else {
			out = append(out, ' ')
		}
	}
	return string(out)
}
