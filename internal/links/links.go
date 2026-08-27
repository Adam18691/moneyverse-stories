package links

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
)

// ══════════════════════════════════════════════════════
// 🔗 Internal Links — ربط الفيديوهات ببعضها (معادل End Screen)
// يوتيوب: راجعها يدوياً من استوديو → فيديو → شرائح النهاية
// هنا: نولد النصوص الجاهزة + ملف خريطة الربط لكل فيديو
// ══════════════════════════════════════════════════════

// LinkMap: خريطة الفيديوهات المنتجة — الفيديو الجديد يرتبط بـ 2 السابقين
type LinkMap struct {
	VideoID   int      `json:"video_id"`
	Title     string   `json:"title"`
	RelatedTo []string `json:"related_titles"`
}

// Register: سجل فيديو جديد واربطه بآخر 2 — يرجع نص التعليق التثبيت
func Register(videoID int, title string) string {
	os.MkdirAll("data", 0755)
	var links []LinkMap
	if b, err := os.ReadFile("data/links.json"); err == nil {
		json.Unmarshal(b, &links)
	}

	// 🔗 اقترح آخر عنوانين كـ"شاهد التالي"
	var related []string
	for i := len(links) - 1; i >= 0 && len(related) < 2; i-- {
		related = append(related, links[i].Title)
	}

	links = append(links, LinkMap{videoID, title, related})
	if b, _ := json.MarshalIndent(links, "", "  "); b != nil {
		os.WriteFile("data/links.json", b, 0644)
	}

	// 📌 نص التعليق المثبت — انسخه في تعليق مثبت تحت الفيديو
	pin := "📌 قصص أخرى ستätzlich أعجبك:\n"
	for i, r := range related {
		pin += fmt.Sprintf("%d️⃣ %s\n", i+1, r)
	}
	pin += "\n🔔 اشترك ليصلك كل جديد يومياً!"

	fmt.Printf("   🔗 روابط داخلية: %d ارتباطات (%s)\n", len(related), short(title))
	return pin
}

func short(s string) string {
	if len(s) > 50 {
		return s[:50] + "..."
	}
	return s
}
