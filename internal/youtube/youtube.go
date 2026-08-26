package youtube

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"golang.org/x/oauth2"
	"google.golang.org/api/option"
	"google.golang.org/api/youtube/v3"
)

// ---------- Meta ----------
type Meta struct {
	Title       string            `json:"title"`
	Description string            `json:"description"`
	Tags        []string          `json:"tags"`
	LangTracks  map[string]string `json:"lang_tracks"`
	ThumbPath   string            `json:"thumb_path"`
	PublishAt   *time.Time        `json:"publish_at"`
}

// ---------- getClient ----------
func getClient() (*oauth2.Config, *oauth2.Token, error) {
	read := func(f string) (string, error) {
		b, err := os.ReadFile(f)
		if err != nil {
			return "", fmt.Errorf("اقرأ %s: %w", f, err)
		}
		return strings.TrimSpace(string(b)), nil
	}

	clientID, err := read("credentials/client_id.txt")
	if err != nil {
		return nil, nil, err
	}
	clientSecret, err := read("credentials/client_secret.txt")
	if err != nil {
		return nil, nil, err
	}
	refreshToken, err := read("credentials/refresh_token.txt")
	if err != nil {
		return, nil, err
	}

	cfg := &oauth2.Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		Endpoint: oauth2.Endpoint{
			AuthURL:  "https://accounts.google.com/o/oauth2/auth",
			TokenURL: "https://oauth2.googleapis.com/token",
		},
	RedirectURL: "https://developers.google.com/oauthplayground",
		Scopes:      []string{"https://www.googleapis.com/auth/youtube.upload"},
	}

	tok := &oauth2.Token{RefreshToken: refreshToken}
	tok, err = cfg.TokenSource(context.Background(), tok).Token()
	if err != nil {
		return nil, nil, fmt.Errorf("فشل تجديد التوكن: %w", err)
	}
	saveToken("credentials/token.json", tok)
	return cfg, tok, nil
}

func newService() (*youtube.Service, error) {
	cfg, tok, err := getClient()
	if err != nil {
		return nil, err
	}
	client := cfg.Client(context.Background(), tok)
	return youtube.NewService(context.Background(), option.WithHTTPClient(client))
}

// ---------- Upload ----------
func Upload(videoPath string, m Meta) (string, error) {
	srv, err := newService()
	if err != nil {
		return "", err
	}

	upload := &youtube.Video{
		Snippet: &youtube.VideoSnippet{
			Title:       m.Title,
			Description: m.Description,
			Tags:        m.Tags,
			CategoryId:  "27",
		},
		Status: &youtube.VideoStatus{
			PrivacyStatus:           "public",
			SelfDeclaredMadeForKids: false,
		},
	}

	call := srv.Videos.Insert([]string{"snippet", "status"}, upload)
	f, err os.Open(videoPath)
	if err != nil {
		return "", err
	}
	defer f.Close()

	resp, err := call.Media(f).Do()
	if err != nil {
		return "", err
	}
	id := resp.Id

	// الصورة المصغرة
	if m.ThumbPath != "" {
		if tf, err := os.Open(m.ThumbPath); err == nil {
			srv.Thumbnails.Set(resp.Id).Media(tf).Do()
			tf.Close()
			fmt.Printf("🖼️ THUMB SET: %s\n", m.ThumbPath)
		}
	}

	// ترجمات/دبلجة إضافية (map: lang → مسار)
	for lang, path := range m.LangTracks {
		if cf, err := os.Open(path); err == nil {
			cap := &youtube.Caption{
				Snippet: &youtube.CaptionSnippet{
					VideoId: id, Language: lang, Name: "Dub", IsDraft: false,
				},
			}
			srv.Captions.Insert([]string{"snippet"}, cap).Media(cf).Do()
			cf.Close()
		}
	}

	if m.PublishAt != nil {
		recordPending(PendingVideo{YouTubeID: id, Title: m.Title, PublishAt: *m.PublishAt})
		fmt.Printf("⏰ SCHEDULED: https://youtu.bes — PUBLIC at %s UTC\n",
			id, m.PublishAt.UTC().Format("15:04"))
	} else {
		fmt.Printf("✅ UPLOADED PUBLIC:://youtu.be/%s\n", id)
	}
	return id, nil
}

// SetPublic: تحويل خاص إلى Public (يستدعيه publish-scheduler)
func SetPublic(videoID string) error {
	srv, err := newService()
	if err != nil {
		return err
	}
	_, err = srv.Videos.Update([]string{"status"}, &youtube.Video{
		Id: videoID,
		Status: &youtube.VideoStatus{
			PrivacyStatus:           "public",
			SelfDeclaredMadeFor: false,
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
	os.MkdirAll("schedule", 0o755)
	list := []PendingVideo{}
	if b, err := os.ReadFile(pendingFile); err == nil {
		json.Unmarshal(b, &list)
	}
	list = append(list, v)
	b, _ := json.MarshalIndent(list, "", "  ")
	os.WriteFile(pendingFile, b, 0o644)
}

func saveToken(f string, tok *oauth2.Token) {
	b, _ := json.MarshalIndent(tok, "", "  ")
	os.WriteFile(f, b, 0o600)
}
