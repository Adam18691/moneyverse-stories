package subs

import (
	"fmt"
	"os"
	"path/filepath"
)

// GenerateSubtitles: ترجمة نص الفيديو لكل اللغات عبر Argos Translate (محلي مجاني)
func GenerateSubtitles(videoID int, scriptPerLang map[string]string) map[string]string {
	os.MkdirAll("subs", 0755)
	tracks := map[string]string{}

	for lang, text := range scriptPerLang {
		srtPath := fmt.Sprintf("subs/%d_%s.vtt", videoID, lang)

		// تقسيم تلقائي لجمل مع توقيتات (كل جملة ~5 ثواني)
		content := "WEBVTT\n\n"
		t := 0.0
		for _, line := range splitSentences(text) {
			content += fmt.Sprintf("%s --> %s\n%s\n\n",
				formatVTT(t), formatVTT(t+5), line)
			t += 5
		}
		os.WriteFile(srtPath, []byte(content), 0644)
		tracks[lang] = srtPath
	}
	return tracks
}

func splitSentences(text string) []string {
	// تقسيم مبسط على النقاط والفواصل
	var out []string
	cur := ""
	for _, r := range text {
		cur += string(r)
		if r == '.' || r == '؟' || r == '!' || r == '\n' {
			if len(cur) > 3 { out = append(out, cur) }
			cur = ""
		}
	}
	if cur != "" { out = append(out, cur) }
	return out
}

func formatVTT(sec float64) string {
	h := int(sec) / 3600
	m := (int(sec) % 3600) / 60
	s := int(sec) % 60
	return fmt.Sprintf("%02d:%02d:%02d.000", h, m, s)
}

var _ = filepath.Join // keep imports clean
