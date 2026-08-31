```go
// cmd/publisher/publish.go
//
// Publisher for Moneyverse Stories.
// ينشر الفيديوهات التي حان موعد نشرها على YouTube.
//
// ملاحظات:
// - يعتمد على internal/youtube.SetPublic()
// - يدعم pending.json بصيغة object أو array القديمة
// - يمنع تشغيل Publisher مرتين في نفس الوقت
// - يحفظ الحالة بطريقة Atomic Write
// - يسجل عدد المحاولات الفاشلة
// - لا يعتبر الفيديو Published إلا بعد نجاح YouTube API

package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/Adam18691/moneyverse-stories/internal/youtube"
)

const (
	pendingFile = "schedule/pending.json"
	lockFile    = "schedule/publisher.lock"

	maxAttempts = 3
)

var fileMu sync.Mutex

// ============================================================
// Pending Video
// ============================================================

type PendingVideo struct {
	YouTubeID string    `json:"youtube_id"`
	Title     string    `json:"title"`
	PublishAt time.Time `json:"publish_at"`
	Published bool      `json:"published"`
	Attempts  int       `json:"attempts"`
}

type state struct {
	Videos []PendingVideo `json:"videos"`
}

// ============================================================
// Main
// ============================================================

func main() {
	fmt.Println("🗓️ MONEYVERSE PUBLISHER")
	fmt.Println("════════════════════════════════════")
	fmt.Printf(
		"⏰ CHECK: %s\n",
		time.Now().UTC().Format(time.RFC3339),
	)
	fmt.Printf("📁 STATE: %s\n", pendingFile)
	fmt.Printf("🔁 MAX ATTEMPTS: %d\n", maxAttempts)
	fmt.Println("════════════════════════════════════")

	// --------------------------------------------------------
	// Prevent duplicate publisher processes
	// --------------------------------------------------------

	unlock, err := acquireLock()
	if err != nil {
		fmt.Printf("🛑 Publisher already running: %v\n", err)
		return
	}

	defer unlock()

	// --------------------------------------------------------
	// Load queue
	// --------------------------------------------------------

	pending, err := loadPending()
	if err != nil {
		fmt.Printf("❌ LOAD FAILED: %v\n", err)
		return
	}

	if len(pending) == 0 {
		fmt.Println("📭 لا توجد فيديوهات مجدولة")
		return
	}

	fmt.Printf(
		"📋 PENDING VIDEOS: %d\n",
		len(pending),
	)

	// --------------------------------------------------------
	// Processing
	// --------------------------------------------------------

	now := time.Now().UTC()

	changed := false

	publishedCount := 0
	failedCount := 0
	waitingCount := 0
	skippedCount := 0

	for i := range pending {
		video := &pending[i]

		// ====================================================
		// Normalize YouTube ID
		// ====================================================

		video.YouTubeID = strings.TrimSpace(
			video.YouTubeID,
		)

		// ====================================================
		// Validate ID
		// ====================================================

		if video.YouTubeID == "" {
			fmt.Printf(
				"⚠️ ENTRY %d — missing YouTube ID — skip\n",
				i+1,
			)

			skippedCount++
			continue
		}

		// ====================================================
		// Already Published
		// ====================================================

		if video.Published {
			fmt.Printf(
				"   ✓ %s — already published\n",
				video.YouTubeID,
			)

			skippedCount++
			continue
		}

		// ====================================================
		// Validate PublishAt
		// ====================================================

		if video.PublishAt.IsZero() {
			fmt.Printf(
				"   ⚠️ %s — publish time missing — skip\n",
				video.YouTubeID,
			)

			skippedCount++
			continue
		}

		// Normalize schedule to UTC.
		video.PublishAt = video.PublishAt.UTC()

		// ====================================================
		// Waiting
		// ====================================================

		if now.Before(video.PublishAt) {
			waitingCount++

			fmt.Printf(
				"   ⏳ %s — waiting until %s\n",
				video.YouTubeID,
				video.PublishAt.Format(time.RFC3339),
			)

			continue
		}

		// ====================================================
		// Maximum Attempts
		// ====================================================

		if video.Attempts >= maxAttempts {
			fmt.Printf(
				"   🛑 %s — maximum attempts reached (%d/%d)\n",
				video.YouTubeID,
				video.Attempts,
				maxAttempts,
			)

			// مهم:
			// لا نضع Published=true.
			// الفيديو فشل فعليًا ويحتاج مراجعة.
			skippedCount++
			continue
		}

		// ====================================================
		// Publish
		// ====================================================

		fmt.Printf(
			"   🚀 PUBLISHING: %s",
			video.YouTubeID,
		)

		if strings.TrimSpace(video.Title) != "" {
			fmt.Printf(
				" — %s",
				video.Title,
			)
		}

		fmt.Println()

		err := publishVideo(video.YouTubeID)

		if err != nil {
			video.Attempts++

			failedCount++
			changed = true

			fmt.Printf(
				"   ❌ FAILED: %s\n",
				video.YouTubeID,
			)

			fmt.Printf(
				"      Attempt: %d/%d\n",
				video.Attempts,
				maxAttempts,
			)

			fmt.Printf(
				"      Error: %v\n",
				err,
			)

			continue
		}

		// ====================================================
		// Successful Publish
		// ====================================================

		video.Published = true

		publishedCount++
		changed = true

		fmt.Printf(
			"   ✅ PUBLISHED: https://youtu.be/%s\n",
			video.YouTubeID,
		)

		if strings.TrimSpace(video.Title) != "" {
			fmt.Printf(
				"      %s\n",
				video.Title,
			)
		}
	}

	// ========================================================
	// Save
	// ========================================================

	if changed {
		if err := savePending(pending); err != nil {
			fmt.Printf(
				"❌ SAVE FAILED: %v\n",
				err,
			)
			return
		}

		fmt.Println("💾 Pending state saved")
	}

	// ========================================================
	// Final Report
	// ========================================================

	fmt.Println()
	fmt.Println("════════════════════════════════════")
	fmt.Println("🏁 PUBLISHER COMPLETE")

	fmt.Printf(
		"✅ Published: %d\n",
		publishedCount,
	)

	fmt.Printf(
		"❌ Failed: %d\n",
		failedCount,
	)

	fmt.Printf(
		"⏳ Waiting: %d\n",
		waitingCount,
	)

	fmt.Printf(
		"⏭️ Skipped: %d\n",
		skippedCount,
	)

	fmt.Printf(
		"📋 Total: %d\n",
		len(pending),
	)

	fmt.Printf(
		"🕐 Finished: %s\n",
		time.Now().UTC().Format(time.RFC3339),
	)

	fmt.Println("════════════════════════════════════")
}

// ============================================================
// Publish Video
// ============================================================

func publishVideo(videoID string) error {
	videoID = strings.TrimSpace(videoID)

	if videoID == "" {
		return errors.New("empty YouTube video ID")
	}

	if err := youtube.SetPublic(videoID); err != nil {
		return fmt.Errorf(
			"set video public: %w",
			err,
		)
	}

	return nil
}

// ============================================================
// Load Pending
// ============================================================

func loadPending() ([]PendingVideo, error) {
	fileMu.Lock()
	defer fileMu.Unlock()

	data, err := os.ReadFile(pendingFile)

	if err != nil {
		if os.IsNotExist(err) {
			return []PendingVideo{}, nil
		}

		return nil, fmt.Errorf(
			"read %s: %w",
			pendingFile,
			err,
		)
	}

	if len(strings.TrimSpace(string(data))) == 0 {
		return []PendingVideo{}, nil
	}

	// ========================================================
	// New format
	//
	// {
	//   "videos": [...]
	// }
	// ========================================================

	var st state

	if err := json.Unmarshal(data, &st); err == nil {
		if st.Videos == nil {
			st.Videos = []PendingVideo{}
		}

		return st.Videos, nil
	}

	// ========================================================
	// Legacy format
	//
	// [
	//   {
	//     "youtube_id": "...",
	//     "title": "...",
	//     "publish_at": "..."
	//   }
	// ]
	// ========================================================

	var legacy []PendingVideo

	if err := json.Unmarshal(data, &legacy); err != nil {
		return nil, fmt.Errorf(
			"invalid pending.json: %w",
			err,
		)
	}

	if legacy == nil {
		legacy = []PendingVideo{}
	}

	return legacy, nil
}

// ============================================================
// Save Pending
// ============================================================

func savePending(videos []PendingVideo) error {
	fileMu.Lock()
	defer fileMu.Unlock()

	dir := filepath.Dir(pendingFile)

	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf(
			"create schedule directory: %w",
			err,
		)
	}

	// ========================================================
	// Normalize state
	// ========================================================

	for i := range videos {
		videos[i].YouTubeID = strings.TrimSpace(
			videos[i].YouTubeID,
		)

		if !videos[i].PublishAt.IsZero() {
			videos[i].PublishAt =
				videos[i].PublishAt.UTC()
		}

		if videos[i].Attempts < 0 {
			videos[i].Attempts = 0
		}
	}

	// ========================================================
	// Encode JSON
	// ========================================================

	payload := state{
		Videos: videos,
	}

	data, err := json.MarshalIndent(
		payload,
		"",
		"  ",
	)

	if err != nil {
		return fmt.Errorf(
			"encode pending state: %w",
			err,
		)
	}

	data = append(data, '\n')

	// ========================================================
	// Atomic Write
	// ========================================================

	tempFile := pendingFile + ".tmp"

	if err := os.WriteFile(
		tempFile,
		data,
		0644,
	); err != nil {
		return fmt.Errorf(
			"write temporary state: %w",
			err,
		)
	}

	if err := os.Rename(
		tempFile,
		pendingFile,
	); err != nil {

		_ = os.Remove(tempFile)

		return fmt.Errorf(
			"replace pending state: %w",
			err,
		)
	}

	return nil
}

// ============================================================
// Publisher Lock
// ============================================================

func acquireLock() (func(), error) {
	dir := filepath.Dir(lockFile)

	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf(
			"create lock directory: %w",
			err,
		)
	}

	file, err := os.OpenFile(
		lockFile,
		os.O_CREATE|os.O_EXCL|os.O_WRONLY,
		0644,
	)

	if err != nil {
		if os.IsExist(err) {
			return nil, errors.New(
				"publisher lock already exists",
			)
		}

		return nil, fmt.Errorf(
			"create publisher lock: %w",
			err,
		)
	}

	_, writeErr := fmt.Fprintf(
		file,
		"pid=%d\nstarted=%s\n",
		os.Getpid(),
		time.Now().UTC().Format(time.RFC3339),
	)

	closeErr := file.Close()

	if writeErr != nil {
		_ = os.Remove(lockFile)

		return nil, fmt.Errorf(
			"write publisher lock: %w",
			writeErr,
		)
	}

	if closeErr != nil {
		_ = os.Remove(lockFile)

		return nil, fmt.Errorf(
			"close publisher lock: %w",
			closeErr,
		)
	}

	unlock := func() {
		if err := os.Remove(lockFile); err != nil &&
			!os.IsNotExist(err) {

			fmt.Printf(
				"⚠️ LOCK CLEANUP FAILED: %v\n",
				err,
			)
		}
	}

	return unlock, nil
}
```
