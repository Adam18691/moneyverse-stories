// ══════════════════════════════════════════
// 🏆 TOP 4: أقوى 4 ترندات عالمياً — أساس الفيديوهات الأربعة
// ══════════════════════════════════════════

type TopTrend struct {
	Rank       int    `json:"rank"`         // 1-4
	Trend      string `json:"trend"`        // الترند نفسه
	Source     string `json:"source"`       // الدولة المصدر "US"
	Continent  string `json:"continent"`    // القارة
	Story      string `json:"story"`        // القصة المالية المرتبطة
	Hook       string `json:"hook"`         // هوك صادم
	Angle      string `json:"angle"`        // زاوية السرد
	ViralScore int    `json:"viral_score"`  // 1-10
	BestRegion string `json:"best_region"`  // أفضل سوق للنشر US/SA/...
}

// Top4Trends: AI يرتب أقوى 4 ترندات من كل العالم ويصنع قصة لكل واحد
func Top4Trends(worldTrends map[string][]string) []TopTrend {
	// نجمّع الترندات مع مصدرها وقارتها
	codeToContinent := map[string]string{}
	for _, cont := range World {
		for _, c := range cont.Countries {
			codeToContinent[c.Code] = cont.Name
		}
	}

	var b strings.Builder
	for code, trs := range worldTrends {
		limit := len(trs)
		if limit > 3 {
			limit = 3 // أول 3 من كل دولة — توفير tokens
		}
		for i := 0; i < limit; i++ {
			b.WriteString(fmt.Sprintf("%s|قارة:%s|%s\n",
				trs[i], codeToContinent[code], code))
		}
	}

	prompt := fmt.Sprintf(`ترندات Google الحقيقية اليوم من 28 دولة (6 قارات)، بصيغة:
الترند|القارة|كود الدولة
%s

المطلوب JSON فقط: اختر أقوى 4 ترندات عالمياً يمكن تحويلها لقصص مالية/أعمال ملهمة.
شرط مهم: الترندات الأربعة من دول/قارات مختلفة قدر الإمكان + لكل واحدة أفضل سوق نشر.
{"trends": [
 {"rank":1,"trend":"...","source":"US","continent":"🌎 أمريكا الشمالية",
  "story":"قصة مالية ملهمة 200 كلمة بالعربي","hook":"جملة صادمة <15 كلمة",
  "angle":"زاوية السرد","viral_score":9,"best_region":"US"},
 {"rank":2,...},{"rank":3,...},{"rank":4,...}
]}`, b.String())

	resp, err := ai.Chat("خبير فيروسية محتوى قصص المال العالمي", prompt)
	if err != nil {
		fmt.Println("   ⚠️ AI Top4 فشل → fallback للترندات الخام")
		return fallbackTop4(worldTrends)
	}

	var out struct {
		Trends []TopTrend `json:"trends"`
	}
	if json.Unmarshal([]byte(extractJSON(resp)), &out) == nil && len(out.Trends) > 0 {
		// 🔒 حماية: دائماً بالضبط 4 أو أقل
		if len(out.Trends) > 4 {
			out.Trends = out.Trends[:4]
		}
	.Println("\n🏆 TOP 4 GLOBAL TRENDS:")
		for _, t := range out.Trends {
			fmt.Printf("   #%d (%d/10) [%s %s] %s → %s\n",
				t.Rank, t.ViralScore, t.Source, t.Continent, t.Trend, t.Hook)
		}
		return out.Trends
	}
	return fallbackTop4(worldTrends)
}

// fallbackTop4: لو الـ AI فشل — أول 4 ترندات خام من أقوى الأسواق
func fallbackTop4(worldTrends map[string][]string) []TopTrend {
	priority := []string{"US", "GB", "DE", "SA", "AE", "IN", "BR", "JP"}
	var out []TopTrend
	rank := 1
	for _, code := range priority {
		trs, ok := worldTrends[code]
		if !ok || len(trs) == 0 {
			continue
		}
		out = append(out, TopTrend{
			Rank: rank, Trend: trs[0], Source: code,
			Story: "قصة مالية ملهمة عن " + trs[0],
			Hook:  "ما حدسدهاش حد!", Angle: "دراما مالية",
			ViralScore: 5, BestRegion: code,
		})
		rank++
		if rank > 4 { // 🔒 أصلاً 4
			break
		}
	}
	return out
}
