```go
// publish.go — ينشر الفيديوهات المجدولة التي حان وقتها
package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"
)

const (
	pendingFile = "schedule/pending.json"

	tokenURL = "https://oauth2.googleapis.com/token"

	youtubeVideosURL = "https://www.googleapis.com/youtube/v3/videos?part=status"

	maxAttempts = 3

	httpTimeout = 30 * time.Second
)

// PendingVideo: فيديو مجدول في طابور النشر.
type PendingVideo struct {
	YouTubeID string    `json:"youtube_id"`
	Title     string    `json:"title"`
	PublishAt time.Time `json:"publish_at"`
	Published bool      `json:"published"`
	Attempts  int       `json:"attempts"`
}

// state: بنية ملف الطابور.
type state struct {
	Videos []PendingVideo `json:"videos"`
}

// HTTP client موحد بمهلة زمنية لمنع تعليق الـ publisher.
var httpClient = &http.Client{
	Timeout: httpTimeout,
}

func main() {
	fmt.Println(
		"🗓️ PUBLISHER CHECK —",
		time.Now().UTC().Format("2006-01-02 15:04:05 UTC"),
	)

	pending, err := loadPending()

	if err != nil {
		fmt.Printf(
			"❌ تعذر قراءة طابور النشر: %v\n",
			err,
		)
		os.Exit(1)
	}

	if len(pending) == 0 {
		fmt.Println("📭 لا يوجد فيديوهات مجدولة")
		return
	}

	now := time.Now().UTC()

	changed := false
	processed := 0
	published := 0
	failed := 0
	waiting := 0

	for i := range pending {
		video := &pending[i]

		// =========================
		// Already Published
		// =========================

		if video.Published {
			continue
		}

		// =========================
		// Validate YouTube ID
		// =========================

		video.YouTubeID = strings.TrimSpace(
			video.YouTubeID,
		)

		if video.YouTubeID == "" {
			fmt.Println(
				"   ⚠️ SKIP — فيديو بدون YouTube ID",
			)
			continue
		}

		// =========================
		// Validate Publish Time
		// =========================

		if video.PublishAt.IsZero() {
			fmt.Printf(
				"   ⚠️ SKIP %s — وقت النشر غير محدد\n",
				video.YouTubeID,
			)
			continue
		}

		publishAt := video.PublishAt.UTC()

		// =========================
		// Not Due Yet
		// =========================

		if now.Before(publishAt) {
			waiting++

			fmt.Printf(
				"   ⏳ WAIT %s — %s\n",
				video.YouTubeID,
				publishAt.Format(
					"2006-01-02 15:04 UTC",
				),
			)

			continue
		}

		// =========================
		// Maximum Attempts
		// =========================

		if video.Attempts >= maxAttempts {
			fmt.Printf(
				"   🚫 BLOCKED %s — %d/%d attempts used\n",
				video.YouTubeID,
				video.Attempts,
				maxAttempts,
			)

			// لا نضع Published=true.
			// يظل الفيديو موجودًا للمراجعة اليدوية.
			continue
		}

		processed++

		fmt.Printf(
			"   🚀 PUBLISHING %s — %s\n",
			video.YouTubeID,
			video.Title,
		)

		// =========================
		// Publish
		// =========================

		if err := setPublic(video.YouTubeID); err != nil {
			video.Attempts++

			failed++
			changed = true

			fmt.Printf(
				"   ❌ FAILED %s — attempt %d/%d — %v\n",
				video.YouTubeID,
				video.Attempts,
				maxAttempts,
				err,
			)

			continue
		}

		// =========================
		// Confirm Published
		// =========================

		video.Published = true

		published++
		changed = true

		fmt.Printf(
			"   ✅ PUBLISHED: https://youtu.be/%s — %s\n",
			video.YouTubeID,
			video.Title,
		)
	}

	// =========================
	// Save State
	// =========================

	if changed {
		if err := savePending(pending); err != nil {
			fmt.Printf(
				"❌ فشل حفظ حالة طابور النشر: %v\n",
				err,
			)

			os.Exit(1)
		}

		fmt.Println(
			"💾 Scheduler state saved",
		)
	}

	// =========================
	// Report
	// =========================

	fmt.Println(
		"\n📊 PUBLISHER REPORT",
	)

	fmt.Printf(
		"   📦 Processed : %d\n",
		processed,
	)

	fmt.Printf(
		"   ✅ Published : %d\n",
		published,
	)

	fmt.Printf(
		"   ❌ Failed    : %d\n",
		failed,
	)

	fmt.Printf(
		"   ⏳ Waiting   : %d\n",
		waiting,
	)

	fmt.Println(
		"🏁 Publisher check complete",
	)
}

// accessToken يستبدل refresh_token بـ access_token
// باستخدام Google OAuth 2.0.
func accessToken() (string, error) {
	clientID := strings.TrimSpace(
		os.Getenv("YT_CLIENT_ID"),
	)

	clientSecret := strings.TrimSpace(
		os.Getenv("YT_CLIENT_SECRET"),
	)

	refreshToken := strings.TrimSpace(
		os.Getenv("YT_REFRESH_TOKEN"),
	)

	if clientID == "" {
		return "",
			errors.New(
				"YT_CLIENT_ID غير موجود",
			)
	}

	if clientSecret == "" {
		return "",
			errors.New(
				"YT_CLIENT_SECRET غير موجود",
			)
	}

	if refreshToken == "" {
		return "",
			errors.New(
				"YT_REFRESH_TOKEN غير موجود",
			)
	}

	// Google OAuth Token Endpoint
	// يتطلب application/x-www-form-urlencoded.
	form := url.Values{}

	form.Set(
		"client_id",
		clientID,
	)

	form.Set(
		"client_secret",
		clientSecret,
	)

	form.Set(
		"refresh_token",
		refreshToken,
	)

	form.Set(
		"grant_type",
		"refresh_token",
	)

	req, err := http.NewRequest(
		http.MethodPost,
		tokenURL,
		strings.NewReader(
			form.Encode(),
		),
	)

	if err != nil {
		return "",
			fmt.Errorf(
				"create OAuth request: %w",
				err,
			)
	}

	req.Header.Set(
		"Content-Type",
		"application/x-www-form-urlencoded",
	)

	resp, err := httpClient.Do(req)

	if err != nil {
		return "",
			fmt.Errorf(
				"OAuth request failed: %w",
				err,
			)
	}

	defer resp.Body.Close()

	responseBody, err := io.ReadAll(
		io.LimitReader(
			resp.Body,
			1024*1024,
		),
	)

	if err != nil {
		return "",
			fmt.Errorf(
				"read OAuth response: %w",
				err,
			)
	}

	// يجب التحقق من HTTP status قبل اعتبار الاستجابة ناجحة.
	if resp.StatusCode < 200 ||
		resp.StatusCode >= 300 {

		detail := strings.TrimSpace(
			string(responseBody),
		)

		if detail == "" {
			detail = "empty response"
		}

		return "",
			fmt.Errorf(
				"OAuth token exchange failed: HTTP %d: %s",
				resp.StatusCode,
				detail,
			)
	}

	var out struct {
		AccessToken string `json:"access_token"`
		TokenType   string `json:"token_type"`
		ExpiresIn   int    `json:"expires_in"`
	}

	if err := json.Unmarshal(
		responseBody,
		&out,
	); err != nil {
		return "",
			fmt.Errorf(
				"decode OAuth response: %w",
				err,
			)
	}

	if strings.TrimSpace(
		out.AccessToken,
	) == "" {
		return "",
			errors.New(
				"Google لم يرجع access_token",
			)
	}

	return out.AccessToken, nil
}

// setPublic يجعل فيديو YouTube عامًا.
func setPublic(videoID string) error {
	videoID = strings.TrimSpace(
		videoID,
	)

	if videoID == "" {
		return errors.New(
			"YouTube video ID فارغ",
		)
	}

	token, err := accessToken()

	if err != nil {
		return err
	}

	payload := map[string]interface{}{
		"id": videoID,

		"status": map[string]interface{}{
			"privacyStatus": "public",
		},
	}

	body, err := json.Marshal(
		payload,
	)

	if err != nil {
		return fmt.Errorf(
			"encode YouTube request: %w",
			err,
		)
	}

	req, err := http.NewRequest(
		http.MethodPut,
		youtubeVideosURL,
		bytes.NewReader(body),
	)

	if err != nil {
		return fmt.Errorf(
			"create YouTube request: %w",
			err,
		)
	}

	req.Header.Set(
		"Authorization",
		"Bearer "+token,
	)

	req.Header.Set(
		"Content-Type",
		"application/json",
	)

	resp, err := httpClient.Do(req)

	if err != nil {
		return fmt.Errorf(
			"YouTube API request failed: %w",
			err,
		)
	}

	defer resp.Body.Close()

	responseBody, _ := io.ReadAll(
		io.LimitReader(
			resp.Body,
			2*1024*1024,
		),
	)

	if resp.StatusCode < 200 ||
		resp.StatusCode >= 300 {

		detail := strings.TrimSpace(
			string(responseBody),
		)

		if detail == "" {
			detail = "empty response"
		}

		return fmt.Errorf(
			"YouTube API HTTP %d: %s",
			resp.StatusCode,
			detail,
		)
	}

	return nil
}

// loadPending يقرأ schedule/pending.json.
func loadPending() ([]PendingVideo, error) {
	data, err := os.ReadFile(
		pendingFile,
	)

	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}

		return nil,
			fmt.Errorf(
				"read %s: %w",
				pendingFile,
				err,
			)
	}

	if len(
		bytes.TrimSpace(data),
	) == 0 {
		return nil, nil
	}

	var st state

	if err := json.Unmarshal(
		data,
		&st,
	); err != nil {
		return nil,
			fmt.Errorf(
				"invalid JSON in %s: %w",
				pendingFile,
				err,
			)
	}

	return st.Videos, nil
}

// savePending يحفظ حالة الطابور بطريقة ذرية.
// نكتب أولًا إلى ملف مؤقت ثم نستبدل الملف الأصلي.
func savePending(
	videos []PendingVideo,
) error {
	dir := filepath.Dir(
		pendingFile,
	)

	if err := os.MkdirAll(
		dir,
		0755,
	); err != nil {
		return fmt.Errorf(
			"create schedule directory: %w",
			err,
		)
	}

	data, err := json.MarshalIndent(
		state{
			Videos: videos,
		},
		"",
		"  ",
	)

	if err != nil {
		return fmt.Errorf(
			"encode scheduler state: %w",
			err,
		)
	}

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
			"replace scheduler state: %w",
			err,
		)
	}

	return nil
}
```
