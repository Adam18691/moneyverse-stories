// publish.go — ينشر الفيديوهات المجدولة التي حان وقتها
package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"
)

const pendingFile = "schedule/pending.json"

// PendingVideo: فيديو مجدول في الطابور
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

func main() {
	fmt.Println("🗓️ PUBLISHER CHECK —", time.Now().UTC().Format("15:04 UTC"))

	pending := loadPending()
	if len(pending) == 0 {
		fmt.Println("📭 لا يوجد فيديوهات مجدولة")
		return
	}

	now := time.Now().UTC()
	changed := false

	for i := range pending {
		v := &pending[i]
		if v.Published || now.Before(v.PublishAt) {
			continue
		}
		if v.Attempts >= 3 {
			fmt.Printf("   ⚠️ SKIP %s — 3 محاولات فاشلة\n", v.YouTubeID)
			v.Published = true
			changed = true
			continue
		}

		if err := setPublic(v.YouTubeID); err != nil {
			v.Attempts++
			fmt.Printf("   ❌ FAILED %s — %v\n", v.YouTubeID, err)
			changed = true
			continue
		}

		v.Published = true
		changed = true
		fmt.Printf("   ✅ PUBLISHED: https://youtu.be/%s — %s\n", v.YouTubeID, v.Title)
	}

	if changed {
		savePending(pending)
	}
	fmt.Println("🏁 Publisher check complete")
}

// accessToken: تبديل refresh_token بـ access_token
func accessToken() (string, error) {
	clientID := os.Getenv("YT_CLIENT_ID")
	secret := os.Getenv("YT_CLIENT_SECRET")
	refresh := os.Getenv("YT_REFRESH_TOKEN")
	if clientID == "" || secret == "" || refresh == "" {
		return "", fmt.Errorf("أسرار YouTube ناقصة")
	}

	body, _ := json.Marshal(map[string]string{
		"client_id":     clientID,
		"client_secret": secret,
		"refresh_token": refresh,
		"grant_type":    "refresh_token",
	})
	resp, err := http.Post("https://oauth2.googleapis.com/token",
		"application/json", bytes.NewBuffer(body))
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var out struct {
		AccessToken string `json:"access_token"`
	}
	if json.NewDecoder(resp.Body).Decode(&out) != nil || out.AccessToken == "" {
		return "", fmt.Errorf("token exchange failed")
	}
	return out.AccessToken, nil
}

// setPublic: جعل الفيديو عاماً
func setPublic(videoID string) error {
	tok, err := accessToken()
	if err != nil {
		return err
	}

	body, _ := json.Marshal(map[string]interface{}{
		"id": videoID,
		"status": map[string]interface{}{
			"privacyStatus": "public",
		},
	})
	req, err := http.NewRequest("PUT",
		"https://www.googleapis.com/youtube/v3/videos?part=status",
		bytes.NewBuffer(body))
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+tok)
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return fmt.Errorf("youtube api status %d", resp.StatusCode)
	}
	return nil
}

func loadPending() []PendingVideo {
	b, err := os.ReadFile(pendingFile)
	if err != nil {
		return nil
	}
	var st state
	if json.Unmarshal(b, &st) != nil {
		return nil
	}
	return st.Videos
}

func savePending(videos []PendingVideo) {
	_ = os.MkdirAll("schedule", 0755)
	b, _ := json.MarshalIndent(state{Videos: videos}, "", "  ")
	_ = os.WriteFile(pendingFile, b, 0644)
}

