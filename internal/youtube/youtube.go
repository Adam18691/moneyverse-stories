```go
package youtube

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"golang.org/x/oauth2"
	"google.golang.org/api/option"
	"google.golang.org/api/youtube/v3"
)

const (
	clientIDFile     = "credentials/client_id.txt"
	clientSecretFile = "credentials/client_secret.txt"
	refreshTokenFile = "credentials/refresh_token.txt"
	tokenFile        = "credentials/token.json"

	pendingFile = "schedule/pending.json"

	youtubeUploadScope = "https://www.googleapis.com/auth/youtube.upload"

	categoryID = "27"

	requestTimeout = 60 * time.Second
)

// ============================================================
// Meta
// ============================================================

type Meta struct {
	Title       string            `json:"title"`
	Description string            `json:"description"`
	Tags        []string          `json:"tags"`
	LangTracks  map[string]string `json:"lang_tracks"`
	ThumbPath   string            `json:"thumb_path"`
	PublishAt   *time.Time        `json:"publish_at"`
}

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

type pendingState struct {
	Videos []PendingVideo `json:"videos"`
}

var pendingMu sync.Mutex

// ============================================================
// OAuth Client
// ============================================================

// getClient يستخدم refresh token فقط.
// لا يفتح متصفحًا ولا يحتاج Authorization flow جديد.
func getClient() (*oauth2.Config, *oauth2.Token, error) {
	clientID, err := readCredential(clientIDFile)
	if err != nil {
		return nil, nil, err
	}

	clientSecret, err := readCredential(clientSecretFile)
	if err != nil {
		return nil, nil, err
	}

	refreshToken, err := readCredential(refreshTokenFile)
	if err != nil {
		return nil, nil, err
	}

	if clientID == "" {
		return nil, nil, errors.New(
			"YouTube client ID is empty",
		)
	}

	if clientSecret == "" {
		return nil, nil, errors.New(
			"YouTube client secret is empty",
		)
	}

	if refreshToken == "" {
		return nil, nil, errors.New(
			"YouTube refresh token is empty",
		)
	}

	cfg := &oauth2.Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,

		Endpoint: oauth2.Endpoint{
			AuthURL:  "https://accounts.google.com/o/oauth2/auth",
			TokenURL: "https://oauth2.googleapis.com/token",
		},

		RedirectURL: "https://developers.google.com/oauthplayground",

		Scopes: []string{
			youtubeUploadScope,
		},
	}

	refresh := &oauth2.Token{
		RefreshToken: refreshToken,
	}

	ctx, cancel := context.WithTimeout(
		context.Background(),
		requestTimeout,
	)
	defer cancel()

	tokenSource := cfg.TokenSource(
		ctx,
		refresh,
	)

	token, err := tokenSource.Token()
	if err != nil {
		return nil, nil, fmt.Errorf(
			"refresh token failed: %w",
			err,
		)
	}

	if token.AccessToken == "" {
		return nil, nil, errors.New(
			"Google returned an empty access token",
		)
	}

	// حفظ التوكن اختياري.
	// فشل الحفظ لا يمنع عملية الرفع.
	if err := saveToken(
		tokenFile,
		token,
	); err != nil {
		fmt.Printf(
			"⚠️ token cache warning: %v\n",
			err,
		)
	}

	return cfg, token, nil
}

// ============================================================
// Credential Reader
// ============================================================

func readCredential(path string) (string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "",
			fmt.Errorf(
				"read %s: %w",
				path,
				err,
			)
	}

	return strings.TrimSpace(
		string(data),
	), nil
}

// ============================================================
// YouTube Service
// ============================================================

func newService() (*youtube.Service, error) {
	cfg, token, err := getClient()
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithTimeout(
		context.Background(),
		requestTimeout,
	)
	defer cancel()

	client := cfg.Client(
		ctx,
		token,
	)

	service, err := youtube.NewService(
		ctx,
		option.WithHTTPClient(client),
	)

	if err != nil {
		return nil,
			fmt.Errorf(
				"create YouTube service: %w",
				err,
			)
	}

	return service, nil
}

// ============================================================
// Upload
// ============================================================

// Upload يرفع الفيديو.
//
// إذا كان PublishAt في المستقبل:
//   - يرفع الفيديو Private.
//   - يسجل الفيديو في schedule/pending.json.
//   - publish.go سيحوّله إلى Public عند الموعد.
//
// إذا لم يوجد PublishAt:
//   - يرفع الفيديو Public مباشرة.
func Upload(
	videoPath string,
	m Meta,
) (string, error) {

	videoPath = strings.TrimSpace(
		videoPath,
	)

	if videoPath == "" {
		return "",
			errors.New(
				"video path is empty",
			)
	}

	videoInfo, err := os.Stat(videoPath)
	if err != nil {
		return "",
			fmt.Errorf(
				"video file unavailable: %w",
				err,
			)
	}

	if videoInfo.IsDir() {
		return "",
			errors.New(
				"video path points to a directory",
			)
	}

	if videoInfo.Size() == 0 {
		return "",
			errors.New(
				"video file is empty",
			)
	}

	title := strings.TrimSpace(
		m.Title,
	)

	if title == "" {
		return "",
			errors.New(
				"YouTube title is empty",
			)
	}

	srv, err := newService()
	if err != nil {
		return "", err
	}

	// ========================================================
	// Privacy
	// ========================================================

	privacyStatus := "public"

	if m.PublishAt != nil {
		publishAt := m.PublishAt.UTC()

		if publishAt.After(
			time.Now().UTC(),
		) {
			privacyStatus = "private"
		}
	}

	// ========================================================
	// YouTube Video
	// ========================================================

	video := &youtube.Video{
		Snippet: &youtube.VideoSnippet{
			Title:       title,
			Description: m.Description,
			Tags:        m.Tags,
			CategoryId:  categoryID,
			DefaultLanguage: "ar",
		},

		Status: &youtube.VideoStatus{
			PrivacyStatus:           privacyStatus,
			SelfDeclaredMadeForKids: false,
		},
	}

	// ========================================================
	// Open Video
	// ========================================================

	file, err := os.Open(videoPath)
	if err != nil {
		return "",
			fmt.Errorf(
				"open video: %w",
				err,
			)
	}

	defer file.Close()

	fmt.Printf(
		"   📤 Uploading: %s\n",
		videoPath,
	)

	// ========================================================
	// Upload
	// ========================================================

	response, err := srv.Videos.
		Insert(
			[]string{
				"snippet",
				"status",
			},
			video,
		).
		Media(file).
		Do()

	if err != nil {
		return "",
			fmt.Errorf(
				"YouTube upload failed: %w",
				err,
			)
	}

	if response == nil ||
		response.Id == "" {

		return "",
			errors.New(
				"YouTube returned no video ID",
			)
	}

	videoID := response.Id

	fmt.Printf(
		"   📺 VIDEO ID: %s\n",
		videoID,
	)

	// ========================================================
	// Thumbnail
	// ========================================================

	if m.ThumbPath != "" {
		if err := setThumbnail(
			srv,
			videoID,
			m.ThumbPath,
		); err != nil {

			// فشل الصورة لا يلغي نجاح الرفع.
			fmt.Printf(
				"   ⚠️ THUMBNAIL FAILED: %v\n",
				err,
			)
		}
	}

	// ========================================================
	// Captions
	// ========================================================

	if len(m.LangTracks) > 0 {
		uploadCaptions(
			srv,
			videoID,
			m.LangTracks,
		)
	}

	// ========================================================
	// Future Schedule
	// ========================================================

	if m.PublishAt != nil {
		publishAt := m.PublishAt.UTC()

		if publishAt.After(
			time.Now().UTC(),
		) {

			err := recordPending(
				PendingVideo{
					YouTubeID: videoID,
					Title:     title,
					PublishAt: publishAt,
					Published: false,
					Attempts:  0,
				},
			)

			if err != nil {
				return videoID,
					fmt.Errorf(
						"video uploaded but pending schedule could not be saved: %w",
						err,
					)
			}

			fmt.Printf(
				"⏰ SCHEDULED: https://youtu.be/%s — PUBLIC at %s UTC\n",
				videoID,
				publishAt.Format(
					time.RFC3339,
				),
			)

			return videoID, nil
		}
	}

	// ========================================================
	// Immediate Public
	// ========================================================

	fmt.Printf(
		"✅ UPLOADED PUBLIC: https://youtu.be/%s\n",
		videoID,
	)

	return videoID, nil
}

// ============================================================
// Set Public
// ============================================================

// SetPublic يحول فيديو YouTube من Private إلى Public.
func SetPublic(
	videoID string,
) error {

	videoID = strings.TrimSpace(
		videoID,
	)

	if videoID == "" {
		return errors.New(
			"video ID is empty",
		)
	}

	srv, err := newService()
	if err != nil {
		return err
	}

	video := &youtube.Video{
		Id: videoID,

		Status: &youtube.VideoStatus{
			PrivacyStatus:           "public",
			SelfDeclaredMadeForKids: false,
		},
	}

	_, err = srv.Videos.
		Update(
			[]string{"status"},
			video,
		).
		Do()

	if err != nil {
		return fmt.Errorf(
			"set video public failed: %w",
			err,
		)
	}

	fmt.Printf(
		"🌐 PUBLIC: https://youtu.be/%s\n",
		videoID,
	)

	return nil
}

// ============================================================
// Thumbnail
// ============================================================

func setThumbnail(
	srv *youtube.Service,
	videoID string,
	path string,
) error {

	path = strings.TrimSpace(path)

	if path == "" {
		return errors.New(
			"thumbnail path is empty",
		)
	}

	info, err := os.Stat(path)
	if err != nil {
		return fmt.Errorf(
			"thumbnail unavailable: %w",
			err,
		)
	}

	if info.IsDir() {
		return errors.New(
			"thumbnail path points to a directory",
		)
	}

	file, err := os.Open(path)
	if err != nil {
		return fmt.Errorf(
			"open thumbnail: %w",
			err,
		)
	}

	defer file.Close()

	_, err = srv.Thumbnails.
		Set(videoID).
		Media(file).
		Do()

	if err != nil {
		return fmt.Errorf(
			"upload thumbnail: %w",
			err,
		)
	}

	fmt.Printf(
		"🖼️ THUMB SET: %s\n",
		path,
	)

	return nil
}

// ============================================================
// Captions
// ============================================================

func uploadCaptions(
	srv *youtube.Service,
	videoID string,
	tracks map[string]string,
) {

	for lang, path := range tracks {
		lang = strings.TrimSpace(lang)
		path = strings.TrimSpace(path)

		if lang == "" || path == "" {
			continue
		}

		info, err := os.Stat(path)
		if err != nil ||
			info.IsDir() {

			fmt.Printf(
				"   ⚠️ CAPTION MISSING [%s]: %s\n",
				lang,
				path,
			)

			continue
		}

		file, err := os.Open(path)
		if err != nil {
			fmt.Printf(
				"   ⚠️ CAPTION OPEN FAILED [%s]: %v\n",
				lang,
				err,
			)
			continue
		}

		caption := &youtube.Caption{
			Snippet: &youtube.CaptionSnippet{
				VideoId: videoID,
				Language: lang,
				Name:     "Moneyverse",
				IsDraft:  false,
			},
		}

		_, err = srv.Captions.
			Insert(
				[]string{"snippet"},
				caption,
			).
			Media(file).
			Do()

		file.Close()

		if err != nil {
			fmt.Printf(
				"   ⚠️ CAPTION FAILED [%s]: %v\n",
				lang,
				err,
			)
			continue
		}

		fmt.Printf(
			"   📝 CAPTION SET: %s\n",
			lang,
		)
	}
}

// ============================================================
// Pending State
// ============================================================

// recordPending يحفظ الفيديو في نفس البنية التي يقرأها
// cmd/publisher/publish.go.
func recordPending(
	video PendingVideo,
) error {

	if video.YouTubeID == "" {
		return errors.New(
			"cannot record pending video without YouTube ID",
		)
	}

	if video.PublishAt.IsZero() {
		return errors.New(
			"cannot record pending video without publish time",
		)
	}

	pendingMu.Lock()
	defer pendingMu.Unlock()

	if err := os.MkdirAll(
		filepath.Dir(pendingFile),
		0755,
	); err != nil {
		return fmt.Errorf(
			"create schedule directory: %w",
			err,
		)
	}

	state := pendingState{
		Videos: []PendingVideo{},
	}

	// ========================================================
	// Load Existing State
	// ========================================================

	data, err := os.ReadFile(
		pendingFile,
	)

	if err == nil {
		if len(data) > 0 {
			if err := json.Unmarshal(
				data,
				&state,
			); err != nil {

				return fmt.Errorf(
					"decode pending.json: %w",
					err,
				)
			}
		}
	} else if !os.IsNotExist(err) {
		return fmt.Errorf(
			"read pending.json: %w",
			err,
		)
	}

	// ========================================================
	// Duplicate Protection
	// ========================================================

	for _, existing := range state.Videos {
		if existing.YouTubeID ==
			video.YouTubeID {

			return nil
		}
	}

	// ========================================================
	// Append
	// ========================================================

	state.Videos = append(
		state.Videos,
		video,
	)

	// ========================================================
	// Encode
	// ========================================================

	data, err = json.MarshalIndent(
		state,
		"",
		"  ",
	)

	if err != nil {
		return fmt.Errorf(
			"encode pending.json: %w",
			err,
		)
	}

	// ========================================================
	// Atomic Save
	// ========================================================

	tempPath := pendingFile + ".tmp"

	if err := os.WriteFile(
		tempPath,
		data,
		0644,
	); err != nil {

		return fmt.Errorf(
			"write temporary pending.json: %w",
			err,
		)
	}

	if err := os.Rename(
		tempPath,
		pendingFile,
	); err != nil {

		_ = os.Remove(
			tempPath,
		)

		return fmt.Errorf(
			"replace pending.json: %w",
			err,
		)
	}

	return nil
}

// ============================================================
// Token Cache
// ============================================================

func saveToken(
	path string,
	token *oauth2.Token,
) error {

	if token == nil {
		return errors.New(
			"cannot save nil OAuth token",
		)
	}

	if err := os.MkdirAll(
		filepath.Dir(path),
		0755,
	); err != nil {
		return fmt.Errorf(
			"create credential directory: %w",
			err,
		)
	}

	data, err := json.MarshalIndent(
		token,
		"",
		"  ",
	)

	if err != nil {
		return fmt.Errorf(
			"encode OAuth token: %w",
			err,
		)
	}

	if err := os.WriteFile(
		path,
		data,
		0600,
	); err != nil {
		return fmt.Errorf(
			"write OAuth token: %w",
			err,
		)
	}

	return nil
}

// ============================================================
// Health Check
// ============================================================

// CheckConnection يختبر اتصال YouTube API.
func CheckConnection() error {
	srv, err := newService()
	if err != nil {
		return err
	}

	response, err := srv.Channels.
		List([]string{"snippet"}).
		Mine(true).
		Do()

	if err != nil {
		return fmt.Errorf(
			"YouTube connection check failed: %w",
			err,
		)
	}

	if response == nil ||
		len(response.Items) == 0 {

		return errors.New(
			"YouTube account/channel was not returned",
		)
	}

	fmt.Printf(
		"✅ YouTube connected: %s\n",
		response.Items[0].Snippet.Title,
	)

	return nil
}

// ============================================================
// HTTP Helper
// ============================================================

// defaultHTTPClient موجود لتوفير timeout موحد عند الحاجة
// في مكونات مستقبلية تعتمد على HTTP.
func defaultHTTPClient() *http.Client {
	return &http.Client{
		Timeout: requestTimeout,
	}
}
```
