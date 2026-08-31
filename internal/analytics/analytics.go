```go
package analytics

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"
)

const apiBase = "https://www.googleapis.com/youtube/v3"

type Video struct {
	ID    string `json:"videoId"`
	Title string `json:"title"`

	Stats struct {
		Views int `json:"viewCount"`
		Likes int `json:"likeCount"`
	} `json:"statistics"`
}

// Report generates a performance report for the latest uploaded videos.
func Report() {
	videos := fetchUploads()

	if len(videos) == 0 {
		fmt.Println("📊 analytics: no YouTube data available — skipping")
		return
	}

	best := videos[0]
	worst := videos[0]

	var b strings.Builder

	b.WriteString("\n═════ 📊 تقرير الأداء — آخر 20 فيديو ═════\n")

	for _, v := range videos {
		if v.Stats.Views > best.Stats.Views {
			best = v
		}

		if v.Stats.Views < worst.Stats.Views {
			worst = v
		}

		b.WriteString(
			fmt.Sprintf(
				"  ▫️ [%5d مشاهدة 👍%d] %s\n",
				v.Stats.Views,
				v.Stats.Likes,
				cut(v.Title),
			),
		)
	}

	b.WriteString(
		fmt.Sprintf(
			"\n🥇 الأفضل: %s (%d مشاهدة)\n",
			best.Title,
			best.Stats.Views,
		),
	)

	b.WriteString(
		fmt.Sprintf(
			"📉 الأضعف: %s (%d مشاهدة)\n",
			worst.Title,
			worst.Stats.Views,
		),
	)

	b.WriteString(
		"💡 القرار: استخدم القوالب الناجحة لتحسين المحتوى القادم\n",
	)

	fmt.Print(b.String())

	if err := os.MkdirAll("data", 0755); err != nil {
		fmt.Printf("⚠️ analytics: cannot create data directory: %v\n", err)
		return
	}

	report := map[string]interface{}{
		"best":        best.Title,
		"best_views":  best.Stats.Views,
		"worst":       worst.Title,
		"worst_views": worst.Stats.Views,
		"updated":     time.Now().UTC().Format("2006-01-02"),
	}

	data, err := json.MarshalIndent(
		report,
		"",
		"  ",
	)

	if err != nil {
		fmt.Printf("⚠️ analytics: cannot encode report: %v\n", err)
		return
	}

	if err := os.WriteFile(
		"data/analytics.json",
		data,
		0644,
	); err != nil {
		fmt.Printf(
			"⚠️ analytics: cannot save report: %v\n",
			err,
		)
	}
}

// fetchUploads retrieves the latest videos from the authenticated YouTube channel.
func fetchUploads() []Video {
	access, ok := getAccessToken()

	if !ok {
		return nil
	}

	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	get := func(
		u string,
		out interface{},
	) bool {
		req, err := http.NewRequest(
			http.MethodGet,
			u,
			nil,
		)

		if err != nil {
			return false
		}

		req.Header.Set(
			"Authorization",
			"Bearer "+access,
		)

		resp, err := client.Do(req)

		if err != nil {
			return false
		}

		defer resp.Body.Close()

		if resp.StatusCode < 200 || resp.StatusCode >= 300 {
			return false
		}

		return json.NewDecoder(
			resp.Body,
		).Decode(out) == nil
	}

	// 1) Get the authenticated channel upload playlist.
	var ch struct {
		Items []struct {
			ContentDetails struct {
				RelatedPlaylists struct {
					Uploads string `json:"uploads"`
				} `json:"relatedPlaylists"`
			} `json:"contentDetails"`
		} `json:"items"`
	}

	channelURL := apiBase +
		"/channels?part=contentDetails&mine=true"

	if !get(channelURL, &ch) ||
		len(ch.Items) == 0 {
		return nil
	}

	uploads := ch.Items[0].
		ContentDetails.
		RelatedPlaylists.
		Uploads

	if uploads == "" {
		return nil
	}

	// 2) Get the latest 20 videos.
	var pl struct {
		Items []struct {
			ContentDetails struct {
				VideoID string `json:"videoId"`
			} `json:"contentDetails"`
		} `json:"items"`
	}

	playlistURL := fmt.Sprintf(
		"%s/playlistItems?part=contentDetails&maxResults=20&playlistId=%s",
		apiBase,
		url.QueryEscape(uploads),
	)

	if !get(playlistURL, &pl) {
		return nil
	}

	if len(pl.Items) == 0 {
		return nil
	}

	// 3) Retrieve statistics.
	out := make(
		[]Video,
		0,
		len(pl.Items),
	)

	for _, item := range pl.Items {
		videoID := item.ContentDetails.VideoID

		if videoID == "" {
			continue
		}

		var vid struct {
			Items []struct {
				Snippet struct {
					Title string `json:"title"`
				} `json:"snippet"`

				Statistics struct {
					Views int `json:"viewCount"`
					Likes int `json:"likeCount"`
				} `json:"statistics"`
			} `json:"items"`
		}

		videoURL := fmt.Sprintf(
			"%s/videos?part=snippet,statistics&id=%s",
			apiBase,
			url.QueryEscape(videoID),
		)

		if !get(videoURL, &vid) ||
			len(vid.Items) == 0 {
			continue
		}

		v := Video{
			ID:    videoID,
			Title: vid.Items[0].Snippet.Title,
		}

		v.Stats.Views = vid.Items[0].Statistics.Views
		v.Stats.Likes = vid.Items[0].Statistics.Likes

		out = append(
			out,
			v,
		)
	}

	return out
}

// getAccessToken exchanges the YouTube refresh token for an access token.
func getAccessToken() (string, bool) {
	refreshToken := readSecret(
		"secrets/yt_refresh_token",
	)

	if refreshToken == "" {
		return "", false
	}

	clientID := readSecret(
		"secrets/yt_client_id",
	)

	clientSecret := readSecret(
		"secrets/yt_client_secret",
	)

	if clientID == "" ||
		clientSecret == "" {
		return "", false
	}

	resp, err := http.PostForm(
		"https://oauth2.googleapis.com/token",
		url.Values{
			"client_id": {
				clientID,
			},
			"client_secret": {
				clientSecret,
			},
			"refresh_token": {
				refreshToken,
			},
			"grant_type": {
				"refresh_token",
			},
		},
	)

	if err != nil {
		return "", false
	}

	defer resp.Body.Close()

	if resp.StatusCode < 200 ||
		resp.StatusCode >= 300 {
		return "", false
	}

	var token struct {
		AccessToken string `json:"access_token"`
	}

	if err := json.NewDecoder(
		resp.Body,
	).Decode(&token); err != nil {
		return "", false
	}

	if token.AccessToken == "" {
		return "", false
	}

	return token.AccessToken, true
}

func readSecret(path string) string {
	data, err := os.ReadFile(path)

	if err != nil {
		return ""
	}

	return strings.TrimSpace(
		string(data),
	)
}

func cut(s string) string {
	runes := []rune(s)

	const maxLength = 55

	if len(runes) > maxLength {
		return string(runes[:maxLength]) + "..."
	}

	return s
}
```
