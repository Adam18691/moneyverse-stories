# 💰 MoneyVerse Stories Engine

> مصنع فيديوهات قصص المال والأعمال — Go 100% PURE — بدون FFmpeg

![Go](https://img.shields.io/badge/Go-1.22-00ADD8)
![Status](https://img.shields.io/badge/videos-4%2Fday-gold)

## ✨ المميزات
- 🔥 4 فيديوهات يومياً تلقائياً (GitHub Actions Cron)
- 🎯 هوك نفسي 7 ثواني + مونتاج cut كل 4 ثوانٍ + Pattern Interrupt كل 90 ثانية
- 🌍 دبلجة وترجمة لـ 10+ لغات (Piper TTS مفتوح المصدر)
- 🔥 ترندات Google اليومية → مطابقة تلقائية للقصص
- 🖼️ ثامبنيل ذهبي CTR عالي (ImageMagick)
- ⏰ نشر مجدول في ذروة كل منطقة (أمريكا/الخليج/الهند/تركيا/إندونيسيا)
- 📤 رفع حقيقي على يوتيوب Public + Chapters + Captions + Thumbnails
- ♾️ توليد برومبتات لا نهائي — ملايين التركيبات الفريدة
- ⚡ GStreamer + MLT Melt — Zero FFmpeg

## 🚀 التشغيل المحلي (أول مرة — لتوليد token.json)
```bash
mkdir credentials && cp client_secret.json credentials/
go mod tidy && go run main.go
