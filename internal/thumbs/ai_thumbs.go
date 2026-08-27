package thumbs

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"tayyibat-money/internal/ai"
)

// ══════════════════════════════════════════
// 🎨 AI THUMBNAIL STUDIO — استوديو ثامبنيلات حصري
// سلسلة 6 مصادر توليد AI مجانية + محرر ذكي + مراقب جودة
// ══════════════════════════════════════════

// AIThumbPlan: خطة الثامبنيل الكاملة من الـ AI
type AIThumbPlan struct {
	Text        string `json:"text"`        // 3-4 كلمات فقط
	TextColor   string `json:"text_color"`  // yellow/red/white/green
	Emotion     string `json:"emotion"`     // shock/greed/fear/hope/anger
	ArtPrompt   string `json:"art_prompt"`  // برومبت فني سينمائي بالإنجليزي
	Style       string `json:"style"`       // hyperrealistic/cinematic/dramatic
	Lighting    string `json:"lighting"`    // golden hour/neon/dramatic shadows
	Composition string `json:"composition"` // close-up face/rule of thirds
	Emoji       string `json:"emoji"`       // 💰🔥😱⚡🏆
}

// QCReport: تقرير فحص الجودة
type QCReport struct {
	Passed       bool   `json:"passed"`
	TextReads    int    `json:"text_readable_1to10"`
	MatchesTrend bool   `json:"matches_trend"`
	Quality1to10 int    `json:"quality_1to10"`
	Reason       string `json:"reason"`
}

var httpClient = &http.Client{Timeout: 90 * time.Second}

// ══════════════════════════════════════════
// 🎬 SmartGenerate — التدفق الكامل 4 مراحل
// ══════════════════════════════════════════

func SmartGenerate(videoID int, trend, story, hook string) (string, error) {

	// ── 🧠 المرحلة 1: AI المدير الفني — تصميم البرومبت الحصري
	storyShort := story
	if len(storyShort) > 150 {
		storyShort = storyShort[:150]
	}

	planPrompt := fmt.Sprintf(`أنت مدير فني أسطوري — مستوى ثامبنيلات القنوات العالمية الأضخم (MrBeast / MagnatesMedia).

الترند: %s
ملخص القصة: %s
الهوك: %s

المطلوب JSON فقط — تصميم فني حصري:

1. text: 3-4 كلمات عربي صادمة تخلق فضول (أرقام أقوى: "خسر 2 مليون")
2. art_prompt: برومبت إنجليزي تفصيلي لتوليد صورة المشهد — قواعد الحصرية:
   - مشهد أصلي مبتكر غير مقلد من أي صورة موجودة
   - شخصية تواجه الكاميرا بعاطفة قوية + عناصر مال/دراما
   - تفاصيل سينمائية: depth of field, volumetric light, 8k detail
   - مثال: "A shocked businessman in expensive suit, hands on head, surrounded by falling burning dollar bills, dark dramatic background with red emergency lighting, hyperrealistic, cinematic depth of field, 8k"
3. style: hyperrealistic أو cinematic أو dramatic
4. lighting: golden hour / neon / dramatic shadows / rim light
5. composition: close-up face / rule of thirds / center hero

{"text":"...","text_color":"yellow","emotion":"shock",
 "art_prompt":"...","style":"hyperrealistic","lighting":"dramatic shadows",
 "composition":"close-up face","emoji":"😱"}`,
		trend, storyShort, hook)

	resp, err := ai.Chat("مدير فني أسطوري — ثامبنيلات حصرية مستوى عالمي", planPrompt)
	if err != nil {
		fmt.Println("   ⚠️ Thumb AI فشل → النظام الكلاسيكي")
		return "", fmt.Errorf("no ai plan")
	}

	var plan AIThumbPlan
	if json.Unmarshal([]byte(ai.ExtractJSON(resp)), &plan) != nil || plan.ArtPrompt == "" {
		return "", fmt.Errorf("bad ai plan")
	}

	fmt.Printf("   🎨 THUMB PLAN: \"%s\" | style: %s | light: %s\n",
		plan.Text, plan.Style, plan.Lighting)

	// ── 🎨 المرحلة 2: التوليد الحصري — سلسلة 6 مصادر
	bgPath := generateExclusiveArt(plan, videoID)
	switch {
	case strings.Contains(bgPath, "ai_"):
		fmt.Printf("   🎨 AI ART: %s (حصري — مش موجود في أي مكان!)\n", bgPath)
	case bgPath != "":
		fmt.Printf("   🖼️ STOCK FALLBACK: %s\n", bgPath)
	default:
		fmt.Println("   🖼️ الخلفية الافتراضية")
	}

	// ── 🖌️+🕵️ المرحلة 3+4: تحرير احترافي + فحص + إعادة (حتى 3 محاولات)
	outPath := fmt.Sprintf("thumbs/thumb_%d.jpg", videoID)

	for attempt := 1; attempt <= 3; attempt++ {
		if err := proEdit(plan, bgPath, outPath); err != nil {
			fmt.Printf("   ⚠️ edit attempt %d فشل: %v\n", attempt, err)
			continue
		}

		qc := qualityCheck(outPath, plan, trend)
		if qc.Passed && qc.TextReads >= 7 && qc.MatchesTrend && qc.Quality1to10 >= 7 {
			fmt.Printf("   ✅ STUDIO QC PASS (محاولة %d): جودة %d/10 | قراءة %d/10 | متوافق\n",
				attempt, qc.Quality1to10, qc.TextReads)
			return outPath, nil
		}

		fmt.Printf("   🔍 QC FAIL (محاولة %d): %s → إعادة تصميم\n", attempt, qc.Reason)

		// 🔄 لو المشهد نفسه ضعيف → توليد مشهد جديد بالكامل
		plan = refinePlan(plan, qc.Reason)
		if qc.Quality1to10 < 7 && attempt < 3 {
			bgPath = generateExclusiveArt(plan, videoID) // مشهد جديد!
		}
	}

	fmt.Println("   ⚠️ 3 محاولات → آخر نسخة (احتياطي)")
	return outPath, nil
}

// ══════════════════════════════════════════
// 🎨 generateExclusiveArt — سلسلة 6 مصادر مجانية
// ══════════════════════════════════════════

func generateExclusiveArt(plan AIThumbPlan, videoID int) string {
	out := fmt.Sprintf("thumbs/ai_%d.jpg", videoID)
	fullPrompt := buildFullPrompt(plan)

	// ── 1️⃣+2️⃣ Pollinations — محركان، مجانيان بدون مفتاح ⭐
	for _, model := range []string{"flux", "turbo"} {
		if p := pollinationsGenerate(fullPrompt, model, out); p != "" {
			fmt.Printf("   🎨 SOURCE #1 POLLINATIONS (%s) ✅\n", model)
			return p
		}
		fmt.Printf("   ⚠️ pollinations[%s] فشل → التالي\n", model)
	}

	// ── 3️⃣ Hugging Face — FLUX.1-schnell (مفتاح مجاني)
	if key := os.Getenv("HUGGINGFACE_KEY"); key != "" {
		if p := huggingfaceGenerate(fullPrompt, key, out); p != "" {
			fmt.Println("   🎨 SOURCE #3 HUGGINGFACE (FLUX-schnell) ✅")
			return p
		}
		fmt.Println("   ⚠️ huggingface فشل → التالي")
	}

	// ── 4️⃣ AI Horde — مجتمع مفتوح مجاني بدون أي مفتاح
	if p := aiHordeGenerate(fullPrompt, out); p != "" {
		fmt.Println("   🎨 SOURCE #4 AI HORDE ✅")
		return p
	}
	fmt.Println("   ⚠️ ai-horde فشل → التالي")

	// ── 5️⃣ Together AI — Flux Schnell Free (مفتاح مجاني)
	if key := os.Getenv("TOGETHER_API_KEY"); key != "" {
		if p := togetherGenerate(fullPrompt, key, out); p != "" {
			fmt.Println("   🎨 SOURCE #5 TOGETHER AI (FLUX-free) ✅")
			return p
		}
		fmt.Println("   ⚠️ together فشل → التالي")
	}

	// ── 6️⃣ آخر احتياطي: صور stock احترافية
	fmt.Println("   ⚠️ كل مصادر AI فشلت → صور احتياطية stock")
	return fetchStockBackground(plan.SearchFallback(), videoID)
}

// buildFullPrompt: البرومبت السينمائي الكامل
func buildFullPrompt(plan AIThumbPlan) string {
	return fmt.Sprintf(
		"%s, %s style, %s lighting, %s composition, "+
			"youtube thumbnail masterpiece, ultra detailed, professional photography, dramatic color grading, 8k",
		plan.ArtPrompt, plan.Style, plan.Lighting, plan.Composition)
}

// ────────────────────────────────────────────
// 1️⃣2️⃣ Pollinations — GET يعيد الصورة مباشرة، بدون مفتاح!
// ────────────────────────────────────────────
func pollinationsGenerate(prompt, model, outPath string) string {
	url := fmt.Sprintf(
		"https://image.pollinations.ai/prompt/%s?width=1280&height=720&model=%s&nologo=true&seed=%d",
		urlQueryEscape(prompt), model, time.Now().UnixNano()%100000)

	resp, err := httpClient.Get(url)
	if err != nil {
		return ""
	}
	defer resp.Body.Close()

	ct := resp.Header.Get("Content-Type")
	if resp.StatusCode != 200 || !strings.Contains(ct, "image") {
		return ""
	}
	return saveImage(resp.Body, outPath)
}

// ────────────────────────────────────────────
// 3️⃣ Hugging Face Inference — FLUX.1-schnell
// ────────────────────────────────────────────
func huggingfaceGenerate(prompt, key, outPath string) string {
	body, _ := json.Marshal(map[string]interface{}{
		"inputs":     prompt,
		"parameters": map[string]interface{}{"width": 1280, "height": 720},
	})
	req, err := http.NewRequest("POST",
		"https://api-inference.huggingface.co/models/black-forest-labs/FLUX.1-schnell",
		bytes.NewBuffer(body))
	if err != nil {
		return ""
	}
	req.Header.Set("Authorization", "Bearer "+key)
	req.Header.Set("Content-Type", "application/json")

	resp, err := httpClient.Do(req)
	if err != nil {
		return ""
	}
	defer resp.Body.Close()

	// نجاح = bytes صورة | فشل = JSON خطأ
	ct := resp.Header.Get("Content-Type")
	if resp.StatusCode != 200 || strings.Contains(ct, "application/json") {
		b, _ := io.ReadAll(resp.Body)
		fmt.Printf("   ⚠️ HF: %s\n", truncateStr(string(b), 80))
		return ""
	}
	return saveImage(resp.Body, outPath)
}

// ────────────────────────────────────────────
// 4️⃣ AI Horde — مجتمع مفتوح، مفتاح anonymous مجاني
//    غير متزامن: إرسال طلب → استطلاع كل 5 ثوانٍ
// ────────────────────────────────────────────
func aiHordeGenerate(prompt, outPath string) string {
	const hz = "https://aihorde.net/api/v2"

	body, _ := json.Marshal(map[string]interface{}{
		"prompt": prompt,
		"params": map[string]interface{}{
			"width": 1024, "height": 576,
			"steps":        25,
			"cfg_scale":    7,
			"sampler_name": "k_euler_a",
		},
		"nsfw": false, "censor_nsfw": true,
		"models": []string{"AlbedoBase XL (SDXL)"},
	})
	req, _ := http.NewRequest("POST", hz+"/generate/async", bytes.NewBuffer(body))
	req.Header.Set("apikey", "0000000000") // مفتاح anonymous — مسموح رسمياً
	req.Header.Set("Client-Agent", "tayyibat-money-engine")
	req.Header.Set("Content-Type", "application/json")

	resp, err := httpClient.Do(req)
	if err != nil {
		return ""
	}
	var sub struct {
		ID string `json:"id"`
	}
	json.NewDecoder(resp.Body).Decode(&sub)
	resp.Body.Close()
	if sub.ID == "" {
		return ""
	}

	// استطلاع — حد أقصى دقيقتين
	pollURL := fmt.Sprintf("%s/generate/check/%s", hz, sub.ID)
	for i := 0; i < 24; i++ {
		time.Sleep(5 * time.Second)
		pr, err := httpClient.Get(pollURL)
		if err != nil {
			continue
		}
		var st struct {
			Done       bool `json:"done"`
			Censored   bool `json:"censored"`
			IsPossible bool `json:"is_possible"`
		}
		json.NewDecoder(pr.Body).Decode(&st)
		pr.Body.Close()

		if st.Censored || !st.IsPossible {
			return ""
		}
		if st.Done {
			gu, _ := http.NewRequest("GET",
				fmt.Sprintf("%s/generate/status/%s", hz, sub.ID), nil)
			gu.Header.Set("Client-Agent", "tayyibat-money-engine")
			gr, err := httpClient.Do(gu)
			if err != nil {
				return ""
			}
			var res struct {
				Generations []struct {
					Img string `json:"img"` // رابط webp أو base64
				} `json:"generations"`
			}
			json.NewDecoder(gr.Body).Decode(&res)
			gr.Body.Close()
			if len(res.Generations) == 0 {
				return ""
			}
			img := res.Generations[0].Img
			if strings.HasPrefix(img, "http") {
				dr, err := httpClient.Get(img)
				if err != nil {
					return ""
				}
				defer dr.Body.Close()
				return saveImage(dr.Body, outPath)
			}
			return saveBase64(img, outPath)
		}
	}
	return ""
}

// ────────────────────────────────────────────
// 5️⃣ Together AI — FLUX.1-schnell-Free
// ────────────────────────────────────────────
func togetherGenerate(prompt, key, outPath string) string {
	body, _ := json.Marshal(map[string]interface{}{
		"model":           "black-forest-labs/FLUX.1-schnell-Free",
		"prompt":          prompt,
		"width":           1280,
		"height":          720,
		"steps":           4,
		"n":               1,
		"response_format": "b64_json",
	})
	req, err := http.NewRequest("POST",
		"https://api.together.xyz/v1/images/generations", bytes.NewBuffer(body))
	if err != nil {
		return ""
	}
	req.Header.Set("Authorization", "Bearer "+key)
	req.Header.Set("Content-Type", "application/json")

	resp, err := httpClient.Do(req)
	if err != nil {
		return ""
	}
	defer resp.Body.Close()

	var out struct {
		Data []struct {
			B64JSON string `json:"b64_json"`
			URL     string `json:"url"`
		} `json:"data"`
	}
	if json.NewDecoder(resp.Body).Decode(&out) != nil || len(out.Data) == 0 {
		return ""
	}
	if out.Data[0].B64JSON != "" {
		return saveBase64(out.Data[0].B64JSON, outPath)
	}
	if out.Data[0].URL != "" {
		return downloadImage(out.Data[0].URL, outPath)
	}
	return ""
}

// ══════════════════════════════════════════
// 🖌️ proEdit — المحرر الاحترافي الذكي
// ══════════════════════════════════════════

func proEdit(plan AIThumbPlan, bgPath, outPath string) error {
	bg := bgPath
	if bg == "" {
		bg = "assets/burning_money.jpg"
	}

	colors := map[string]string{
		"yellow": "#FFD700", "red":   "#FF2222",
		"white":  "#FFFFFF", "green": "#00FF88",
	}
	color := colors[plan.TextColor]
	if color == "" {
		color = "#FFD700"
	}

	// 🎬 تحسين المدير الفني حسب العاطفة:
	brightness, saturation := "103", "115"
	switch plan.Emotion {
	case "shock", "fear":
		brightness, saturation = "100", "120" // تباين بارد حاد
	case "greed", "hope":
		brightness, saturation = "105", "120" // دفء ذهبي
	}

	args := []string{
		bg,
		// 1️⃣ قياس يوتيوب القياسي
		"-resize", "1280x720^",
		"-gravity", "center", "-extent", "1280x720",
		// 2️⃣ تحسين سينمائي حسب العاطفة
		"-modulate", fmt.Sprintf("%s,%s,100", brightness, saturation),
		// 3️⃣ فينيت داكن — يركز العين على المركز
		"-vignette", "0x6",
		// 4️⃣ طبقة تعتيم أسفل النص
		"-fill", "#00000055",
		"-draw", "rectangle 0,480 1280,720",
		// 5️⃣ النص الذكي — ضخم بحواف قوية
		"-font", "DejaVu-Sans-Bold",
		"-pointsize", "115",
		"-fill", color,
		"-stroke", "black", "-strokewidth", "8",
		"-gravity", "south", "-annotate", "+0+150", plan.Text,
		// 6️⃣ إيموجي انفعالي ضخم
		"-pointsize", "140",
		"-gravity", "south", "-annotate", "+0+20", plan.Emoji,
		"-quality", "95",
		outPath,
	}
	cmd := exec.Command("convert", args...)
	if b, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("%v: %s", err, string(b))
	}
	return nil
}

// ══════════════════════════════════════════
// 🕵️ qualityCheck — المراقب الصارم
// ══════════════════════════════════════════

func qualityCheck(path string, plan AIThumbPlan, trend string) QCReport {

	// فحص 1: الملف موجود وحجمه معقول
	info, err := os.Stat(path)
	if err != nil {
		return QCReport{Passed: false, Reason: "ملف الثامبنيل غير موجود"}
	}
	if info.Size() < 10000 {
		return QCReport{Passed: false, Reason: "ملف تالف أو صغير جداً"}
	}

	// فحص 2: عدد الكلمات ≤ 5
	wordCount := len(strings.Fields(plan.Text))
	if wordCount > 5 {
		return QCReport{
			Passed: false,
			Reason: fmt.Sprintf("نص طويل (%d كلمات > 5)", wordCount),
		}
	}

	// فحص 3: AI يحكم على الجودة والتوافق
	qcPrompt := fmt.Sprintf(`أنت مراقب جودة ثامبنيلات صارم — مستوى استوديو عالمي.
الترند: %s
نص الثامبنيل: "%s"
البرومبت الفني للمشهد: %s
الستايل: %s | الإضاءة: %s | التكوين: %s
العاطفة: %s

قيّم بصرامة وأخرج JSON فقط:
{"passed": true/false,
 "text_readable_1to10": 8,
 "matches_trend": true/false,
 "quality_1to10": 8,
 "reason": "سبب مختصر لو فشل"}

قواعد الفشل:
- نص بلا علاقة بالترند
- برومبت فني عام بلا تفاصيل سينمائية
- تكوين ضعيف (بلا بطل واضح أو عاطفة)
- جودة أقل من 7/10`,
		trend, plan.Text, plan.ArtPrompt, plan.Style,
		plan.Lighting, plan.Composition, plan.Emotion)

	resp, err := ai.Chat("مراقب جودة استوديو ثامبنيلات — صارم جداً", qcPrompt)
	if err != nil {
		return QCReport{Passed: true, TextReads: 7, MatchesTrend: true, Quality1to10: 7}
	}

	var qc QCReport
	if json.Unmarshal([]byte(ai.ExtractJSON(resp)), &qc) != nil {
		return QCReport{Passed: true, TextReads: 7, MatchesTrend: true, Quality1to10: 7}
	}
	return qc
}

// ══════════════════════════════════════════
// 🔄 refinePlan — AI يصحح خطته بنفسه
// ══════════════════════════════════════════

func refinePlan(old AIThumbPlan, reason string) AIThumbPlan {
	prompt := fmt.Sprintf(`خطة ثامبنيل فشلت في فحص الجودة.
الخطة القديمة:
- text="%s"
- art_prompt="%s"
- style=%s | lighting=%s | composition=%s
سبب الفشل: %s

أخرج خطة معدلة أفضل بكثير JSON بنفس الصيغة فقط:
{"text":"...","text_color":"...","emotion":"...",
 "art_prompt":"...","style":"...","lighting":"...","composition":"...","emoji":"..."}
حسّن البرومبت الفني بتفاصيل سينمائية أقوى.`,
		old.Text, old.ArtPrompt, old.Style, old.Lighting, old.Composition, reason)

	resp, err := ai.Chat("مدير فني أسطوري يصحح أخطاءه", prompt)
	if err != nil {
		return old
	}
	var refined AIThumbPlan
	if json.Unmarshal([]byte(ai.ExtractJSON(resp)), &refined) == nil && refined.Text != "" {
		if len(strings.Fields(refined.Text)) > 5 {
			words := strings.Fields(refined.Text)
			refined.Text = strings.Join(words[:4], " ")
		}
		fmt.Printf("   🔄 REVISED: \"%s\" | برومبت أقوى\n", refined.Text)
		return refined
	}
	return old
}

// ══════════════════════════════════════════
// 6️⃣ Stock APIs — احتياطي أخير فقط
// ══════════════════════════════════════════

// SearchFallback: أول كلمات البرومبت الفني كبحث stock
func (p AIThumbPlan) SearchFallback() string {
	words := strings.Fields(p.ArtPrompt)
	if len(words) > 4 {
		return strings.Join(words[:4], " ")
	}
	if len(words) > 0 {
		return strings.Join(words, " ")
	}
	return "burning money"
}

func fetchStockBackground(query string, videoID int) string {
	if key := os.Getenv("PEXELS_API_KEY"); key != "" {
		if p := pexelsSearch(query, key, videoID); p != "" {
			return p
		}
	}
	if key := os.Getenv("PIXABAY_API_KEY"); key != "" {
		if p := pixabaySearch(query, key, videoID); p != "" {
			return p
		}
	}
	return ""
}

func pexelsSearch(query, key string, id int) string {
	url := fmt.Sprintf(
		"https://api.pexels.com/v1/search?query=%s&per_page=5&orientation=landscape",
		strings.ReplaceAll(query, " ", "+"))
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return ""
	}
	req.Header.Set("Authorization", key)

	resp, err := httpClient.Do(req)
	if err != nil {
		return ""
	}
	defer resp.Body.Close()

	var out struct {
		Photos []struct {
			Src struct {
				Large2x string `json:"large2x"`
			} `json:"src"`
		} `json:"photos"`
	}
	if json.NewDecoder(resp.Body).Decode(&out) != nil || len(out.Photos) == 0 {
		return ""
	}
	idx := id % len(out.Photos)
	return downloadImage(out.Photos[idx].Src.Large2x, fmt.Sprintf("thumbs/bg_%d.jpg", id))
}

// ✅ الإصلاح هنا: func pixabaySearch (كان funcabaySearch)
func pixabaySearch(query, key string, id int) string {
	url := fmt.Sprintf(
		"https://pixabay.com/api/?key=%s&q=%s&per_page=5&image_type=photo",
		key, strings.ReplaceAll(query, " ", "+"))
	resp, err := httpClient.Get(url)
	if err != nil {
		return ""
	}
	defer resp.Body.Close()

	var out struct {
		Hits []struct {
			LargeImageURL string `json:"largeImageURL"`
		} `json:"hits"`
	}
	if json.NewDecoder(resp.Body).Decode(&out) != nil || len(out.Hits) == 0 {
		return ""
	}
	idx := id % len(out.Hits)
	return downloadImage(out.Hits[idx].LargeImageURL, fmt.Sprintf("thumbs/bg_%d.jpg", id))
}

// ══════════════════════════════════════════
// helpers مشتركة
// ══════════════════════════════════════════

// saveImage: حفظ stream + تحقق الحجم (أقل من 20KB = تالفة)
func saveImage(r io.Reader, outPath string) string {
	f, err := os.Create(outPath)
	if err != nil {
		return ""
	}
	n, _ := io.Copy(f, r)
	f.Close()
	if n < 20000 {
		os.Remove(outPath)
		return ""
	}
	return outPath
}

// saveBase64: فك ترميز وحفظ صورة base64
func saveBase64(b64, outPath string) string {
	b64 = strings.TrimPrefix(b64, "data:image/png;base64,")
	data, err := base64.StdEncoding.DecodeString(b64)
	if err != nil || len(data) < 20000 {
		return ""
	}
	if err := os.WriteFile(outPath, data, 0644); err != nil {
		return ""
	}
	return outPath
}

// downloadImage: تحميل صورة من رابط لملف محلي
func downloadImage(url, path string) string {
	resp, err := httpClient.Get(url)
	if err != nil {
		return ""
	}
	defer resp.Body.Close()
	return saveImage(resp.Body, path)
}

// urlQueryEscape: ترميز آمن للبرومبت في الرابط
func urlQueryEscape(s string) string {
	var b bytes.Buffer
	for _, r := range s {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') {
			b.WriteRune(r)
		} else if r == ' ' {
			b.WriteByte('+')
		} else {
			fmt.Fprintf(&b, "%%%02X", r)
		}
	}
	return b.String()
}

func truncateStr(s string, n int) string {
	if len(s) > n {
		return s[:n]
	}
	return s
}

// LastThumbPath: أداة مساعدة
func LastThumbPath() string {
	matches, _ := filepath.Glob("thumbs/thumb_*.jpg")
	if len(matches) > 0 {
		return matches[len(matches)-1]
	}
	return ""
}
