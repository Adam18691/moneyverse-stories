package meta

import (
	"fmt"
	"strings"
)

type DescriptionData struct {
	Hook   string
	Lesson string
}

func BuildDescription(d DescriptionData) string {
	var b strings.Builder

	b.WriteString("💰 " + d.Hook + "\n\n")
	b.WriteString("في هذا الفيديو نكشف لك قصة حقيقية ستغير نظرتك للمال والأعمال نهائياً.\n")
	b.WriteString("📚 الدرس الذهبي: " + d.Lesson + "\n\n")

	b.WriteString("⏱️ الفواصل الزمنية:\n")
	for _, c := range []string{
		"00:00 البداية الصادمة", "01:30 نقطة التحول",
		"03:00 القرار المصيري", "04:30 أول نجاح",
		"06:00 الأزمة الكبرى", "07:30 الصعود من جديد",
		"09:00 الدروس الذهبية 💰", "10:30 الخاتمة الملهمة",
	} {
		b.WriteString(c + "\n")
	}

	b.WriteString("\n─────────────────\n")
	b.WriteString("💰 He lost everything in one night... then became the richest man in his country\n\n")

	b.WriteString("\n🔔 SUBSCRIBE / اشترك:\n")
	for lang, sub := range subscribeLines {
		b.WriteString(sub + " [" + lang + "]\n")
	}

	b.WriteString("\n" + Hashtags() + "\n")
	b.WriteString(fmt.Sprintf("\n© MoneyVerse Stories — جميع الحقوق محفوظة\n"))
	return b.String()
}

var subscribeLines = map[string]string{
	"العربية":   "اشترك وفعّل الجرس 🔔",
	"English":   "Subscribe & hit the bell 🔔",
	"Français":  "Abonne-toi et active la cloche 🔔",
	"Español":   "Suscríbete y activa la campana 🔔",
	"Deutsch":   "Abonnieren & Glocke aktivieren 🔔",
	"Türkçe":    "Abone ol ve zili aç 🔔",
	"हिन्दी":     "सब्सक्राइब करें 🔔",
	"中文":        "订阅并打开铃铛🔔",
	"Indonesia": "Berlangganan & nyalakan lonceng 🔔",
	"اردو":      "سبسکرائب کریں 🔔",
}

func Hashtags() string {
	return "#قصص_نجاح #المال #الأعمال #مليونير #تجارة #ثروة #تحفيز " +
		"#success_story #money #business #motivation #millionaire #wealth"
}
