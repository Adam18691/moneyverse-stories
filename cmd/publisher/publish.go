```go
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

	maxAttempts = 3

	lockFile = "schedule/publisher.lock"
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
	fmt.Println(
		"🗓️ MONEYVERSE PUBLISHER",
	)

	fmt.Println(
		"════════════════════════════════════",
	)

	fmt.Printf(
		"⏰ CHECK: %s\n",
		time.Now().
			UTC().
			Format(time.RFC3339),
	)

	fmt.Printf(
		"📁 STATE: %s\n",
		pendingFile,
	)

	fmt.Printf(
		"🔁 MAX ATTEMPTS: %d\n",
		maxAttempts,
	)

	fmt.Println(
		"════════════════════════════════════",
	)

	// ========================================================
	// Prevent overlapping publisher processes
	// ========================================================

	unlock, err := acquireLock()
	if err != nil {
		fmt.Printf(
			"🛑 Publisher already running: %v\n",
			err,
		)
		return
	}

	defer unlock()

	// ========================================================
	// Load
	// ========================================================

	pending, err := loadPending()
	if err != nil {
		fmt.Printf(
			"❌ LOAD FAILED: %v\n",
			err,
		)
		return
	}

	if len(pending) == 0 {
		fmt.Println(
			"📭 لا يوجد فيديوهات مجدولة",
		)
		return
	}

	fmt.Printf(
		"📋 PENDING VIDEOS: %d\n",
		len(pending),
	)

	// ========================================================
	// Process
	// ========================================================

	now := time.Now().UTC()

	changed := false
	publishedCount := 0
	failedCount := 0
	waitingCount := 0
	skippedCount := 0

	for i := range pending {
		video := &pending[i]

		// ----------------------------------------------------
		// Invalid entry
		// ----------------------------------------------------

		if strings.TrimSpace(
			video.YouTubeID,
		) == "" {

			fmt.Printf(
				"⚠️ ENTRY %d — missing YouTube ID — skip\n",
				i+1,
			)

			skippedCount++
			continue
		}

		// ----------------------------------------------------
		// Already published
		// ----------------------------------------------------

		if video.Published {
			skippedCount++

			fmt.Printf(
				"   ✓ %s — already published\n",
				video.YouTubeID,
			)

			continue
		}

		// ----------------------------------------------------
		// Invalid publish time
		// ----------------------------------------------------

		if video.PublishAt.IsZero() {
			fmt.Printf(
				"   ⚠️ %s — invalid publish time — skip\n",
				video.YouTubeID,
			)

			skippedCount++
			continue
		}

		// ----------------------------------------------------
		// Normalize time to UTC
		// ----------------------------------------------------

		publishAt := video.PublishAt.UTC()

		// ----------------------------------------------------
		// Not due yet
		// ----------------------------------------------------

		if now.Before(publishAt) {
			waitingCount++

			fmt.Printf(
				"   ⏳ %s — waiting until %s UTC\n",
				video.YouTubeID,
				publishAt.Format(
					time.RFC3339,
				),
			)

			continue
		}

		// ----------------------------------------------------
		// Attempts limit
		// ----------------------------------------------------

		if video.Attempts >= maxAttempts {
			fmt.Printf(
				"   🛑 %s — %d attempts reached — permanently skipped\n",
				video.YouTubeID,
				maxAttempts,
			)

			// لا نضع Published=true هنا؛
			// لأن الفيديو لم يُنشر بنجاح.
			//
			// إبقاؤه false يسمح بالتعرف عليه كفيديو
			// فشل نهائيًا بدل اعتباره منشورًا.
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

		if video.Title != "" {
			fmt.Printf(
				" — %s",
				video.Title,
			)
		}

		fmt.Println()

		err := publishVideo(
			video.YouTubeID,
		)

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

		// ----------------------------------------------------
		// Success
		// ----------------------------------------------------

		video.Published = true
		changed = true
		publishedCount++

		fmt.Printf(
			"   ✅ PUBLISHED: https://youtu.be/%s\n",
			video.YouTubeID,
		)

		if video.Title != "" {
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
		if err := savePending(
			pending,
		); err != nil {

			fmt.Printf(
				"❌ SAVE FAILED: %v\n",
				err,
			)

			return
		}

		fmt.Println(
			"💾 Pending state saved",
		)
	}

	// ========================================================
	// Final Report
	// ========================================================

	fmt.Println(
		"\n════════════════════════════════════",
	)

	fmt.Println(
		"🏁 PUBLISHER COMPLETE",
	)

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
		"🕐 Finished: %s UTC\n",
		time.Now().
			UTC().
			Format(time.RFC3339),
	)

	fmt.Println(
		"════════════════════════════════════",
	)
}

// ============================================================
// Publish Video
// ============================================================

func publishVideo(
	videoID string,
) error {

	videoID = strings.TrimSpace(
		videoID,
	)

	if videoID == "" {
		return errors.New(
			"empty YouTube video ID",
		)
	}

	if err := youtube.SetPublic(
		videoID,
	); err != nil {

		return fmt.Errorf(
			"set public: %w",
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

	data, err := os.ReadFile(
		pendingFile,
	)

	if err != nil {
		if os.IsNotExist(err) {
			return []PendingVideo{}, nil
		}

		return nil,
			fmt.Errorf(
				"read %s: %w",
				pendingFile,
				err,
			)
	}

	if len(strings.TrimSpace(
		string(data),
	)) == 0 {

		return []PendingVideo{}, nil
	}

	// ========================================================
	// New format
	// ========================================================

	var st state

	if err := json.Unmarshal(
		data,
		&st,
	); err == nil {

		if st.Videos == nil {
			st.Videos = []PendingVideo{}
		}

		return st.Videos, nil
	}

	// ========================================================
	// Legacy format
	//
	// يدعم أيضًا pending.json القديم:
	//
	// [
	//   {
	//     "youtube_id": "...",
	//     ...
	//   }
	// ]
	// ========================================================

	var legacy []PendingVideo

	if err := json.Unmarshal(
		data,
		&legacy,
	); err != nil {

		return nil,
			fmt.Errorf(
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

func savePending(
	videos []PendingVideo,
) error {

	fileMu.Lock()
	defer fileMu.Unlock()

	if err := os.MkdirAll(
		filepath.Dir(pendingFile),
		0755,
	); err != nil {

		return fmt.Errorf(
			"create schedule directory: %w",
			err,
		)
	}

	// ========================================================
	// Normalize
	// ========================================================

	for i := range videos {
		videos[i].YouTubeID =
			strings.TrimSpace(
				videos[i].YouTubeID,
			)

		if !videos[i].PublishAt.IsZero() {
			videos[i].PublishAt =
				videos[i].PublishAt.UTC()
		}
	}

	// ========================================================
	// Encode
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

	data = append(
		data,
		'\n',
	)

	// ========================================================
	// Atomic write
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

		_ = os.Remove(
			tempFile,
		)

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
	if err := os.MkdirAll(
		filepath.Dir(lockFile),
		0755,
	); err != nil {

		return nil,
			fmt.Errorf(
				"create lock directory: %w",
				err,
			)
	}

	// O_EXCL يجعل إنشاء الملف ذريًا:
	// إذا كان Publisher آخر يعمل بالفعل فسيفشل.
	file, err := os.OpenFile(
		lockFile,
		os.O_CREATE|os.O_EXCL|os.O_WRONLY,
		0644,
	)

	if err != nil {
		if os.IsExist(err) {
			return nil,
				errors.New(
					"lock file already exists",
				)
		}

		return nil,
			fmt.Errorf(
				"create publisher lock: %w",
				err,
			)
	}

	_, _ = fmt.Fprintf(
		file,
		"pid=%d\nstarted=%s\n",
		os.Getpid(),
		time.Now().
			UTC().
			Format(time.RFC3339),
	)

	_ = file.Close()

	unlock := func() {
		_ = os.Remove(
			lockFile,
		)
	}

	return unlock, nil
}
```
