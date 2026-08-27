package thumbs

import (
	"encoding/json"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"strings"
	"time"
)

// ============================================================
//  أدوات
// ============================================================

func readKey(f, env string) string {
	if b, err := os.ReadFile(f); err == nil {
		return strings.TrimSpace(string(b))
	}
	return strings.TrimSpace(os.Getenv(env))
}

func get(u string, headers map[string]string) ([]byte, error) {
	req, _ := http.NewRequest("GET", u, nil)
	for k, v := range headers {
		req.Header.Set(k, v)
	}
	client := &http.ClientTimeout: 40 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	return io.ReadAll(resp.Body)
}

// ============================================================
//  صورة نخبوية + تقييم صارم
// ============================================================

type Photo struct {
	ID     int
	Width  int
	Height int
	URL    string
	Tags   string
}

// Score: الدقة أولًا ثم نسبة 16:9 ثم الطابع الدرامي
func (p Photo) Score() int {
	score := 0
	switch {
	case p.Width >= 3840:
		score += 600
	case p.Width >= 2560:
		score += 400
	case p.Width >= 1920:
		score += 300
	case p.Width >= 1280:
		score += 150
	default:
		score += 50
	}
	if p.Height > 0 {
		ratio := float64(p.Width) / float64(p.Height)
		if ratio > 1.5 && ratio < 1.9 {
			score += 200
		}
	}
	tags := strings.ToLower(p.Tags)
	for _, w := range []string{"dark", "night", "city", "luxury", "money",
		"gold", "business", "skyline", "silhouette", "neon", "success"} {
		if strings.Contains(tags, w) {
			score += 60
		}
	}
	return score
}

// ============================================================
//  الجلب من Pixabay + Pexels ← ثم النخبة فقط
// ============================================================

func FetchBestPhotos(queries []string, need int) []Photo {
	all := []Photo{}
	for _, q := range queries {
		all = append(all, pixabayPhotos(q, 20)...)
		all = append(all, pexelsPhotos(q, 20)...)
	}
	if len(all) == 0 {
		fmt.Println("❌ THUMB FETCH: no photos")
		return nil
	}

	// فرز بالتقييم
	for i := 0; i < len(all); i++ {
		for j := i + 1; j < len(all); j++ {
			if all[j].Score() > all[i].Score() {
				all[i], all[j] = all[j], all[i]
			}
		}
	}

	var elite []Photo
	seen := map[int]bool{}
	for _, p := range all {
		if p.Width < 1280 || seen[p.ID] { // ⛔ تحت HD ممنوع
			continue
		}
		seen[p.ID] = true
		elite = append(elite, p)
		if len(elite) >= need {
			break
		}
	}
	fmt.Printf("🖼️ THUMB PHOTOS: %d elite from %d candidates\n", len(elite), len(all))
	return elite
}

// ---- Pixabay ----

func pixabayPhotos(query string, count int) []Photo {
	key := readKey("secrets/pixabay_key.txt", "PIXABAY_API_KEY")
	if key == "" {
		return nil
	}
	u := fmt.Sprintf(
		"https://pixabay.com/api/?key=%s&q=%s&image_type=photo&orientation=horizontal&min_width=1600&per_page=%d&safesearch=true",
		key, url.QueryEscape(query), count)

	b, err := get(u, nil)
	if err != nil {
		return nil
	}
	var pr struct {
		Hits []struct {
			ID            int    `json:"id"`
			ImageWidth    int    `json:"imageWidth"`
			ImageHeight   int    `json:"imageHeight"`
			LargeImageURL string `json:"largeImageURL"`
			Tags          string `json:"tags"`
		} `json:"hits"`
	}
	if json.Unmarshal(b, &pr) != nil {
		return nil
	}
	var out []Photo
	for _, h := range pr.Hits {
		out = append(out, Photo{ID: h.ID, Width: h.ImageWidth,
			Height: h.ImageHeight, URL: h.LargeImageURL, Tags: h.Tags})
	}
	return out
}

// ---- Pexels ----

func pexelsPhotos(query string, count int) []Photo {
	key := readKey("secrets/pexels_key.txt", "PEXELS_API_KEY")
	if key == "" {
		return nil
	}
	u := fmt.Sprintf(
		"https://api.pexels.com/v1/search?query=%s&orientation=landscape&per_page=%d",
		url.QueryEscape(query), count)

	b, err := get(u, map[string]string{"Authorization": key})
	if err != nil {
		return nil
	}
	var pr struct {
		Photos []struct {
			ID     int    `json:"id"`
			Width  int    `json:"width"`
			Height int    `json:"height"`
			Src    struct {
				Large2x string `json:"large2x"`
				Large   string `json:"large"`
			} `json:"src"`
			Alt string `json:"alt"`
		} `json:"photos"`
	}
	if json.Unmarshal(b, &pr) != nil {
		return nil
	}
	var out []Photo
	for _, p := range pr.Photos {
		link := p.Src.Large2x
		if link == "" {
			link = p.Src.Large
		}
		out = append(out, Photo{ID: p.ID, Width: p.Width,
			Height: p.Height, URL: link, Tags: p.Alt})
	}
	return out
}

func downloadPhoto(p Photo, path string) error {
	client := &http.Client{Timeout: 60 * time.Second}
	resp, err := client.Get(p.URL)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = io.Copy(f, resp.Body)
	return err
}

// ============================================================
//  Generate — التركيب السينمائي النهائي
//  ⚠️ نفس التوقيع الذي يستدعيه main.go
// ============================================================

func Generate(id int, text string, bg string) error {
	os.MkdirAll("thumbs", 0o755)
	out := fmt.Sprintf("thumbs/thumb_%d.jpg", id)

	// ---- 1) خلفية نخبوية 4K — تُجلب تلقائيًا حسب النص ----
	bgPath := fmt.Sprintf("thumbs/bg_%d.jpg", id)
	defer os.Remove(bgPath)

	queries := []string{
		text, // الكلمات المفتاحية من العنوان نفسه
		"dark money luxury gold",
		"city night skyline cinematic",
	}
	found := false
	if elite := FetchBestPhotos(queries, 3); len(elite) > 0 {
		if err := downloadPhoto(elite[0], bgPath); err == nil {
			found = true
		}
	}

	background := bg // إن مرّرت خلفية خارجية نستخدمها
	if !found && background == "" {
		background = "gradient:#0d0d0d-#8a6d00" // ذهبي فخم كبديل
	}
	if found {
		background = bgPath
	}

	// ---- 2) التركيب: قص 1280x720 + تعتيم سفلي + شريط ذهبي + نص ثلاثي الأبعاد ----
	gold := "#c9a227"
	args := []string{
		background,
		"-resize", "1280x720^",
		"-gravity", "center", "-extent", "1280x720",

		// طبقة تعتيم سفلية لوضوح النص
		"-fill", "rgba(0,0,0,0.55)",
		"-draw", "rectangle 0,400 1280,720",

		// شريط ذهبي فخم أسفل الصورة
		"-fill", gold,
		"-draw", "rectangle 0,706 1280,720",

		// النص ثلاثي الأبعاد: ظل أسود سميك → حدود ذهبية → تعبئة بيضاء
		"-font", "DejaVu-Sans-Bold",
		"-pointsize", "72",
		"-gravity", "south",
		"-stroke", "#000000", "-strokewidth", "10",
		"-annotate", "+6+62", text,
		"-stroke", gold, "-strokewidth", "4",
		"-fill", "#ffffff",
		"-annotate", "+0+54", text,

		"-quality", "95",
		out,
	}

	cmd := exec.Command("convert", args...)
	cmd.Stderr = os.Stderr
	err := cmd.Run()

	if err != nil {
		// خطة بديلة مبسطة
		fallback := exec.Command("convert",
			background,
			"-resize", "1280x720^",
			"-gravity", "center", "-extent", "1280x720",
			"-fill", "rgba(0,0,0,0.6)", "-draw", "rectangle 0,420 1280,720",
			"-font", "DejaVu-Sans-Bold", "-pointsize", "72",
			"-gravity", "south",
			"-fill", "#ffffff", "-stroke", "#000000", "-strokewidth", "5",
			"-annotate", "+0+60", text,
			out)
		fallback.Stderr = os.Stderr
		if ferr := fallback.Run(); ferr != nil {
			return ferr
		}
	}

	fmt.Printf("🎬 THUMB %d ELITE CINEMATIC → %s\n", id, out)
	return nil
}
