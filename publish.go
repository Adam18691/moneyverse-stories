// publish.go — ينشر الفيديوهات المجدولة عند وصول موعدها
// يعمل كل ساعتين عبر cron workflow
package main

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	"tayyibat-money/internal/youtube"
)

const pendingFile = "schedule/pending.json"

type PendingVideo = youtube.PendingVideo

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
		if v.Published || now.Before(v.PublishAt) {
			continue
		}
		if err := youtube.SetPublic(v.YouTubeID); err != nil {
			fmt.Printf("❌ FAILED %s: %v\n", v.YouTubeID, err)
			continue
		}
		v.Published = true
		changed = true
		fmt.Printf("✅ PUBLISHED: https://youtu.be/%s — %s\n", v.YouTubeID, v.Title)
	}

	if changed {
		savePending(pending)
	}
	fmt.Println("🏁 Publisher check complete")
}

func loadPending() []PendingVideo {
	b

> ⚠️ The response reached the length limit. Reply **continue** to get the rest.
