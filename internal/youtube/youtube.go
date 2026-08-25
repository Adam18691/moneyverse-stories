package youtube

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/option"
	"google.golang.org/api/youtube/v3"
)

type Meta struct {
	Title       string            `json:"title"`
	Description string            `json:"description"`
	Tags        []string          `json:"tags"`
	LangTracks  map[string]string `json:"-"`
	ThumbPath   string            `json:"-"`
	PublishAt   *time.Time        `json:"publishAt,omitempty"`
}

func getClient() (*http.Client, error) {
	b, err := os.ReadFile("credentials/client_secret.json")
	if err != nil {
		return nil, err
	}
	cfg, err := google.ConfigFromJSON(b,
		youtube.YoutubeUploadScope, youtube.YoutubepartnerScope,
		youtube.YoutubeScope)
	if err != nil {
		return nil, err
	}
	const tokFile = "token.json"
	tok, err := tokenFromFile(tokFile)
	if err != nil {
		tok = getTokenFromWeb(cfg)
		saveToken(tokFile, tok)
	}
	return cfg.Client(context.Background(), tok), nil
}

// Upload: رفع Public فوري أو Private مجدول بـ publishAt
func Upload(videoPath string, m Meta) (string, error) {
	client, err := getClient()
	if err != nil {
		return "", err
	}
	srv, err := youtube.NewService(context.Background(), option.WithHTTPClient(client))
	if err != nil {
		return "", err
	}

	status := &youtube.VideoStatus{SelfDeclaredMadeForKids: false}
	if m.PublishAt != nil {
		status.PrivacyStatus = "private"
		status.PublishAt = m.PublishAt.UTC().Format(time.RFC3339)
	} else {
		status.PrivacyStatus = "public"
	}

	vid := &youtube.Video{
		Snippet: &youtube.VideoSnippet{
			Title: m.Title, Description: m.Description,
			Tags: m.Tags, CategoryId: "27",
			DefaultLanguage: "ar", DefaultAudioLanguage: "ar",
		},
		Status: status,
	}

	file, err := os.Open(videoPath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	resp, err := srv.Videos.Insert([]string{"snippet", "status"}, vid).Media(file).Do()
	if err != nil {
		return "", err
	}
	id := resp.Id

	// الثامبنيل
	if th, err := os.Open(m.ThumbPath); err == nil {
		srv.Thumbnails.Set(id).Media(th).Do()
		th.Close()
	}

	// captions لكل لغة
	for lang, path := range m.LangTracks {
		f, err := os.Open(path)
		if err != nil {
			continue
		}
		cap := &youtube.Caption{
			Snippet: &youtube.CaptionSnippet{
				VideoId: id, Language: lang, Name: "AutoDub", IsDraft: false,
			},
		}
		srv.Captions.Insert([]string{"snippet"}, cap).Media(f).Do()
		f.Close()
	}

	if m.PublishAt != nil {
		recordPending(PendingVideo{YouTubeID: id, Title: m.Title, PublishAt: *m.PublishAt})
		fmt.Printf("⏰ SCHEDULED: https://youtu.be/%s → PUBLIC at %s UTC\n",
			id, m.PublishAt.UTC().Format("15:04"))
	} else {
		fmt.Printf("✅ UPLOADED PUBLIC: https://youtu.be/%s\n", id)
	}
	return id, nil
}

// SetPublic: تحويل فيديو إلى Public (يستخدمه publish.go)
func SetPublic(videoID string) error {
	client, err := getClient()
	if err != nil {
		return err
	}
	srv, err := youtube.NewService(context.Background(), option.WithHTTPClient(client))
	if err != nil {
		return err
	}
	_, err = srv.Videos.Update([]string{"status"}, &youtube.Video{
		Id: videoID,
		Status: &youtube.VideoStatus{
			PrivacyStatus:           "public",
			SelfDeclaredMadeForKids: false,
		},
	}).Do()
	return err
}

// ---------- pending state ----------
type PendingVideo struct {
	YouTubeID string    `json:"youtube_id"`
	Title     string    `json:"title"`
	PublishAt time.Time `json:"publish_at"`
	Published bool      `json:"published"`
}

const pendingFile = "schedule/pending.json"

func recordPending(v PendingVideo) {
	os.MkdirAll("schedule", 0755)
	list := []PendingVideo{}
	if b, err := os.ReadFile(pendingFile); err == nil {
		json.Unmarshal(b, &list)
	}
	list = append(list, v)
	b, _ := json.MarshalIndent(list, "", "  ")
	os.WriteFile(pendingFile, b, 0644)
}

// ---------- oauth helpers ----------
func tokenFromFile(f string) (*oauth2.Token, error) {
	b, err := os.ReadFile(f)
	if err != nil {
		return nil, err
	}
	tok := &oauth2.Token{}
	return tok, json.Unmarshal(b, tok)
}

func saveToken(f string, tok *oauth2.Token) {
	b, _ := json.MarshalIndent(tok, "", "  ")
	os.WriteFile(f, b, 0600)
}

func getTokenFromWeb(cfg *oauth2.Config) *oauth2.Token {
	fmt.Println("🔗 افتح هذا الرابط وامنح الإذن:")
	fmt.Println(cfg.AuthCodeURL("state-token"))
	fmt.Print("\nالصق الكود هنا: ")
	var code string
	fmt.Scan(&code)
	tok, err := cfg.Exchange(context.Background(), code)
	if err != nil {
		panic(err)
	}
	return tok
}
