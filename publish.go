// publish.go — MoneyVerse Scheduled Publisher
// يشغَّل كل ساعتين عبر GitHub Actions cron
// يفحص الفيديوهات المجدولة وينشر اللي جه موعدها

package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"tayyibat-money/internal/schedule"
	"tayyibat-money/internal/youtube"
)

// PendingVideo: سجل فيديو مجدول (يُكتب عند الرفع)
type PendingVideo struct {
	YouTubeID string    `json:"youtube_id"`
	Title     string    `json:"title"`
	PublishAt time.Time `json:"publish_at"`
	Published bool      `json:"published"`
}

const pendingFile = "schedule/pending.json"

func main() {
	fmt.Println("⏰ PUBLISHER CHECK —", time.Now().UTC().Format("15:04 UTC"))

	pending := loadPending()
	if len(pending) == 0 {
		fmt.Println("📭 لا توجد فيديوهات مجدولة")
		return
	}

	now := time.Now().UTC()
	changed := false

	for i := range pending {
		v := &pending[i]

		if v.Published {
			continue // نُشر بالفعل
		}

		if now.Before(v.PublishAt) {
			fmt.Printf("⏳ %s → ينشر بعد %.1f ساعة\n",
				v.YouTubeID, v.PublishAt.Sub(now).Hours())
			continue
		}

		// 🔥 الموعد وصل → انشر الآن
		if err := youtube.SetPublic(v.YouTubeID); err != nil {
			fmt.Printf("❌ FAILED %s: %v\n", v.YouTubeID, err)
			continue
		}
		v.Published = true
		changed = true
		fmt.Printf("✅ PUBLISHED NOW: https://youtu.be/%s — %s\n",
			v.YouTubeID, v.Title)
	}

	if changed {
		savePending(pending)
	}
	fmt.Println("🏁 Publisher check complete")
}

func loadPending() []PendingVideo {
	b, err := os.ReadFile(pendingFile)
	if err != nil {
		return []PendingVideo{}
	}
	var out []PendingVideo
	json.Unmarshal(b, &out)
	return out
}

func savePending(v []PendingVideo) {
	os.MkdirAll("schedule", 0755)
	b, _ := json.MarshalIndent(v, "", "  ")
	os.WriteFile(pendingFile, b, 0644)
}
