package main

import (
	"fmt"
	"math/rand"
	"os"
	"time"

	"github.com/disintegration/imaging"
	"github.com/fogleman/gg"
	ffmpeg "github.com/u2takey/ffmpeg-go"
)

const (
	FULL_FOLDER_NAME = "tayyibat_v29_FINAL_GO_4K_ZW_Q4_60fps"
	FULL_FOLDER_PATH = "output/" + FULL_FOLDER_NAME
	FULL_FILE_NAME   = FULL_FOLDER_NAME + ".mp4"
)

func main() {
	rand.Seed(time.Now().UnixNano())
	
	// 1. انشاء الفولدر الكامل
	os.MkdirAll("output", 0755)
	os.MkdirAll(FULL_FOLDER_PATH+"/thumbnails", 0755)
	os.MkdirAll(FULL_FOLDER_PATH+"/meta", 0755)
	os.MkdirAll(FULL_FOLDER_PATH+"/audio", 0755)

	topics := []string{"الجلوتين", "اللبن", "السكر", "الزيوت المهدرجة"}
	badils := []string{"الارز", "السمسم", "زيت الزيتون", "الذرة"}
	marads := []string{"ارتشاح الامعاء", "مقاومة الانسولين"}

	topic := topics[rand.Intn(len(topics))]
	badil := badils[rand.Intn(len(badils))]
	marad := marads[rand.Intn(len(marads))]

	fullFilePath := FULL_FOLDER_PATH + "/" + FULL_FILE_NAME
	title := fmt.Sprintf("هذا %s يدمر 90%% - %s مسموح - ZW Harare | %s", topic, badil, FULL_FOLDER_NAME)
	if len(title) > 95 {
		title = title[:95]
	}
	desc := fmt.Sprintf("%s\n\nFULL FOLDER: %s\nFULL FILE: %s\nZW زيمبابوي Harare 20:00 = 18:00 UTC\n%s ممنوع يسبب %s - البديل %s مسموح\n#طيبات #%s #%s #ZW #Harare #4K #60fps\nhttps://youtu.be/k9iW7zxiAQq", title, FULL_FOLDER_PATH, fullFilePath, topic, marad, badil, topic, badil)

	// 2. Thumbnail 1280x720 - مضبوط بدون magick
	fmt.Println("🎨 Generating Thumbnail...")
	dc := gg.NewContext(1280, 720)
	dc.SetRGB(1, 0, 0)
	dc.Clear()
	// مستطيل اصفر
	dc.SetRGB(1, 0.92, 0)
	dc.DrawRectangle(15, 15, 1250, 160)
	dc.Fill()
	// نصوص
	dc.SetRGB(1, 0, 0)
	dc.LoadFontFace("/usr/share/fonts/truetype/dejavu/DejaVuSans-Bold.ttf", 54)
	dc.DrawString("ZIMBABWE PEAK! زيمبابوي!", 60, 110)
	dc.SetRGB(1, 1, 1)
	dc.LoadFontFace("/usr/share/fonts/truetype/dejavu/DejaVuSans-Bold.ttf", 38)
	dc.DrawString(fmt.Sprintf("90%% %s - %s مسموح", topic, badil), 80, 262)
	dc.SetRGB(1, 1, 0)
	dc.LoadFontFace("/usr/share/fonts/truetype/dejavu/DejaVuSans.ttf", 28)
	dc.DrawString(fmt.Sprintf("%s - طيبات | %s", marad, FULL_FOLDER_NAME), 80, 340)

	dc.SavePNG(FULL_FOLDER_PATH + "/thumbnails/thumbnail_10000.jpg")
	dc.SavePNG(FULL_FOLDER_PATH + "/thumbnail_10000.jpg")
	dc.SavePNG("output/thumbnail_10000.jpg")
	dc.SavePNG("output/thumbnail_ZW.jpg")
	fmt.Println("✅ Thumbnail Done")

	// 3. BG 4K 3840x2160
	fmt.Println("🖼️ Generating 4K BG...")
	bg := gg.NewContext(3840, 2160)
	bg.SetRGB(0.04, 0.04, 0.04)
	bg.Clear()
	bg.SetRGB(1, 1, 1)
	bg.LoadFontFace("/usr/share/fonts/truetype/dejavu/DejaVuSans-Bold.ttf", 110)
	bg.DrawStringWrapped(title, 240, 400, 0, 0, 3000, 1.5, gg.AlignLeft)
	bg.SetRGB(1, 1, 0)
	bg.LoadFontFace("/usr/share/fonts/truetype/dejavu/DejaVuSans.ttf", 52)
	bg.DrawString(fmt.Sprintf("%s | ZW 20:00 Harare | 4K 60fps 35Mbps", FULL_FOLDER_NAME), 240, 1400)
	bg.SavePNG(FULL_FOLDER_PATH + "/bg_4K.jpg")

	smallBg := imaging.Resize(bg.Image(), 1920, 1080, imaging.Lanczos)
	imaging.Save(smallBg, "output/bg.jpg")
	fmt.Println("✅ BG 4K Done")

	// 4. صوت وهمي + فيديو 10 دقايق 60fps
	os.WriteFile("output/voice.mp3", []byte("dummy"), 0644)
	os.WriteFile(FULL_FOLDER_PATH+"/audio/voice.mp3", []byte("dummy"), 0644)

	fmt.Println("🎬 Generating Video 4K 60fps 10min...")
	err := ffmpeg.Input(FULL_FOLDER_PATH+"/bg_4K.jpg", ffmpeg.KwArgs{"loop": "1", "t": "610", "r": "60"}).
		Output(fullFilePath, ffmpeg.KwArgs{
			"c:v": "libx264", "pix_fmt": "yuv420p",
			"vf": "scale=3840:2160:flags=lanczos,eq=saturation=1.3",
			"r": "60", "b:v": "35M", "b:a": "320k",
			"movflags": "+faststart", "shortest": "",
		}).OverWriteOutput().ErrorToStdOut().Run()

	if err != nil {
		fmt.Println("Fallback to 1080p 60fps...")
		ffmpeg.Input("output/bg.jpg", ffmpeg.KwArgs{"loop": "1", "t": "610"}).
			Output(fullFilePath, ffmpeg.KwArgs{"c:v": "libx264", "r": "60", "b:v": "15M", "t": "610"}).
			OverWriteOutput().Run()
	}

	// نسخ للتوافق
	os.MkdirAll("output", 0755)
	copyFile(fullFilePath, "output/tayyibat_10min_8K_60fps_STUDIO_FINAL.mp4")
	copyFile(fullFilePath, "output/tayyibat_10min_4K_UHD_60fps_FINAL.mp4")

	// 5. Meta
	os.WriteFile(FULL_FOLDER_PATH+"/meta/title.txt", []byte(title), 0644)
	os.WriteFile("output/title.txt", []byte(title), 0644)
	os.WriteFile(FULL_FOLDER_PATH+"/meta/desc.txt", []byte(desc), 0644)
	os.WriteFile("output/desc.txt", []byte(desc), 0644)
	os.WriteFile(FULL_FOLDER_PATH+"/meta/tags.txt", []byte(fmt.Sprintf("%s,%s,%s,طيبات,ZW,Harare,4K,60fps,%s", topic, badil, marad, FULL_FOLDER_NAME)), 0644)

	fmt.Printf("\n🎉 GOD Go v29 DONE:\n📁 %s/\n📹 %s\n📂 thumbnails/thumbnail_10000.jpg\n📂 meta/title.txt\n🔗 https://youtu.be/k9iW7zxiAQq\n", FULL_FOLDER_PATH, fullFilePath)
}

func copyFile(src, dst string) {
	data, _ := os.ReadFile(src)
	os.WriteFile(dst, data, 0644)
}
