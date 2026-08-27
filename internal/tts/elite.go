package tts

import (
	"fmt"
	"os"
	"os/exec"
)

// ============================================================
//  أصوات Microsoft Neural النخبوية — بدون أي مفتاح API
//  مذيعون محترفون واقعيون بدفء سينمائي يشدّ المشاهد
// ============================================================

type EliteVoice struct {
	Name string
	Lang string
	Mood string
}

var arabicVoices = []EliteVoice{
	{Name: "ar-SA-HamedNeural", Lang: "ar", Mood: "مذيع سعودي عميق وقور"},
	{Name: "ar-EG-ShakirNeural", Lang: "ar", Mood: "مذيع مصري واثق حيوي"},
	{Name: "ar-SA-ZariyahNeural", Lang: "ar", Mood: "صوت نسائي فخم راقٍ"},
}

// PickElite: كل فيديو بصوت مختلف (حسب id) — تنويع يحفظ القناة من الملل
func PickElite(id int) EliteVoice {
	v := arabicVoices[id%len(arabicVoices)]
	fmt.Printf("🎙️ ELITE VOICE: %s — %s\n", v.Name, v.Mood)
	return v
}

// SmartSpeak: المحرك الذكي — أصوات النخبة أولًا، وتحويل WAV متوافق مع الرندر
// ⚠️ نفس توقيع الاستدعاء في main.go لسهولة التبديل
func SmartSpeak(id int, text, outPath string) error {
	v := PickElite(id)

	mp3 := outPath + ".elite.mp3"

	// أداة edge-tts المجانية الرسمية (مثبتة في الـ workflow)
	cmd := exec.Command("edge-tts",
		"--voice", v.Name,
		"--rate=-8%",
		"--pitch=-2Hz",
		"--text", text,
		"--write-media", mp3)
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("elite voice failed: %w", err)
	}

	// تحويل MP3 → WAV (16bit mono) ليتوافق مع خط الرندر الحالي
	conv := exec.Command("melt", mp3,
		"-consumer", "avformat:"+outPath,
		"acodec=pcm_s16le", "arate=44100")
	conv.Stderr = os.Stderr
	if err := conv.Run(); err != nil {
		// fallback ثانٍ: استخدم gmconvert الصوتي البسيط إن وُجد
		fmt.Printf("⚠️ melt wav convert failed: %v\n", err)
		return err
	}

	os.Remove(mp3) // تنظيف المؤقت
	fi, _ := os.Stat(outPath)
	fmt.Printf("🎙️ ELITE VOICE DONE → %s (%d KB)\n", outPath, fi.Size()/1024)
	return nil
}
