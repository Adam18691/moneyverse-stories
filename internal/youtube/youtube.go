```go
package youtube

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/option"
	"google.golang.org/api/youtube/v3"
)

// ============================================================
// Configuration
// ============================================================

const (
	credentialsDir = "credentials"

	clientIDFile     = "credentials/client_id.txt"
	clientSecretFile = "credentials/client_secret.txt"
	refreshTokenFile = "credentials/refresh_token.txt"
	tokenFile        = "credentials/token.json"

	pendingFile = "schedule/pending.json"

	youtubeUploadScope =
		"https://www.googleapis.com/auth/youtube.upload"

	youtubeForceSSL =
		"https://www.googleapis.com/auth/youtube.force-ssl"
)

var (
	serviceMu sync.Mutex
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

// ============================================================
// File Helpers
// ============================================================

func readCredential(path string) (string, error) {
	data, err := os.ReadFile(path)

	if err != nil {
		return "",
			fmt.Errorf(
				"read credential %s: %w",
				path,
				err,
			)
	}

	value := strings.TrimSpace(
		string(data),
	)

	if value == "" {
		return "",
			fmt.Errorf(
				"credential file is empty: %s",
				path,
			)
	}

	return value, nil
}

// ============================================================
// OAuth Configuration
// ============================================================

func getOAuthConfig() (*oauth2.Config, string, error) {
	clientID, err := readCredential(
		clientIDFile,
	)

	if err != nil {
		return nil, "", err
	}

	clientSecret, err := readCredential(
		clientSecretFile,
	)

	if err != nil {
		return nil, "", err
	}

	refreshToken, err := readCredential(
		refreshTokenFile,
	)

	if err != nil {
		return nil, "", err
	}

	cfg := &oauth2.Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,

		Endpoint: google.Endpoint,

		Scopes: []string{
			youtubeUploadScope,
			youtubeForceSSL,
		},
	}

	return cfg, refreshToken, nil
}

// ============================================================
// OAuth Client
// ============================================================

func getClient() (*httpClientWrapper, error) {
	cfg, refreshToken, err :=
		getOAuthConfig()

	if err != nil {
		return nil, err
	}

	token := &oauth2.Token{
		RefreshToken: refreshToken,
	}

	ctx := context.Background()

	tokenSource := cfg.TokenSource(
		ctx,
		token,
	)

	accessToken, err :=
		tokenSource.Token()

	if err != nil {
		return nil,
			fmt.Errorf(
				"refresh YouTube OAuth token: %w",
				err,
			)
	}

	// حفظ آخر Access Token للاستخدام التشخيصي.
	if err := saveToken(
		tokenFile,
		accessToken,
	); err != nil {
		fmt.Printf(
			"⚠️ token cache warning: %v\n",
			err,
		)
	}

	return &httpClientWrapper{
		Client: cfg.Client(
			ctx,
			accessToken,
		),
	}, nil
}

// ============================================================
// HTTP Wrapper
// ============================================================

type httpClientWrapper struct {
	Client interface {
		Do(*http.Request) (*http.Response, error)
	}
}

// ============================================================
// YouTube Service
// ============================================================

func newService() (*youtube.Service, error) {
	serviceMu.Lock()
	defer serviceMu.Unlock()

	cfg, refreshToken, err :=
		getOAuthConfig()

	if err != nil {
		return nil, err
	}

	ctx := context.Background()

	token := &oauth2.Token{
		RefreshToken: refreshToken,
	}

	tokenSource := cfg.TokenSource(
		ctx,
		token,
	)

	client := cfg.Client(
		ctx,
		&oauth2.Token{
			RefreshToken: refreshToken,
		},
	)

	// إجبار OAuth على تحديث Access Token
	// إذا كان ذلك مطلوبًا.
	if _, err := tokenSource.Token(); err != nil {
		return nil,
			fmt.Errorf(
				"obtain YouTube access token: %w",
				err,
			)
	}

	return youtube.NewService(
		ctx,
		option.WithHTTPClient(client),
	)
}

// ============================================================
// Upload
// ============================================================

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

	if _, err := os.Stat(
		videoPath,
	); err != nil {
		return "",
			fmt.Errorf(
				"video file unavailable: %w",
				err,
			)
	}

	if strings.TrimSpace(
		m.Title,
	) == "" {
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
	// Upload object
	// ========================================================

	privacy := "public"

	if m.PublishAt != nil &&
		!m.PublishAt.IsZero() {

		privacy = "private"
	}

	upload := &youtube.Video{
		Snippet: &youtube.VideoSnippet{
			Title: m.Title,

			Description: m.Description,

			Tags: m.Tags,

			CategoryId: "27",
		},

		Status: &youtube.VideoStatus{
			PrivacyStatus: privacy,

			SelfDeclaredMadeForKids: false,
		},
	}

	// ========================================================
	// Open video
	// ========================================================

	file, err := os.Open(
		videoPath,
	)

	if err != nil {
		return "",
			fmt.Errorf(
				"open video: %w",
				err,
			)
	}

	defer file.Close()

	// ========================================================
	// Upload
	// ========================================================

	fmt.Printf(
		"📤 Uploading: %s\n",
		videoPath,
	)

	resp, err := srv.
		Videos.
		Insert(
			[]string{
				"snippet",
				"status",
			},
			upload,
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

	if resp == nil ||
		strings.TrimSpace(
			resp.Id,
		) == "" {

		return "",
			errors.New(
				"YouTube returned an empty video ID",
			)
	}

	videoID := resp.Id

	fmt.Printf(
		"📺 VIDEO ID: %s\n",
		videoID,
	)

	// ========================================================
	// Thumbnail
	// ========================================================

	if strings.TrimSpace(
		m.ThumbPath,
	) != "" {

		if err := setThumbnail(
			srv,
			videoID,
			m.ThumbPath,
		); err != nil {

			fmt.Printf(
				"⚠️ Thumbnail failed: %v\n",
				err,
			)
		}
	}

	// ========================================================
	// Captions
	// ========================================================

	if len(m.LangTracks) > 0 {

		if err := uploadCaptions(
			srv,
			videoID,
			m.LangTracks,
		); err != nil {

			fmt.Printf(
				"⚠️ Captions warning: %v\n",
				err,
			)
		}
	}

	// ========================================================
	// Scheduled publishing
	// ========================================================

	if m.PublishAt != nil &&
		!m.PublishAt.IsZero() {

		publishAt := m.PublishAt.UTC()

		if publishAt.After(
			time.Now().UTC(),
		) {

			if err := recordPending(
				PendingVideo{
					YouTubeID: videoID,
					Title: m.Title,
					PublishAt: publishAt,
					Published: false,
					Attempts:  0,
				},
			); err != nil {

				return videoID,
					fmt.Errorf(
						"record schedule: %w",
						err,
					)
			}

			fmt.Printf(
				"⏰ SCHEDULED: https://youtu.be/%s — PUBLIC at %s UTC\n",
				videoID,
				publishAt.Format(
					"2006-01-02 15:04",
				),
			)

			return videoID, nil
		}
	}

	fmt.Printf(
		"✅ UPLOADED PUBLIC: https://youtu.be/%s\n",
		videoID,
	)

	return videoID, nil
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

	file, err := os.Open(path)

	if err != nil {
		return fmt.Errorf(
			"open thumbnail: %w",
			err,
		)
	}

	defer file.Close()

	_, err = srv.
		Thumbnails.
		Set(videoID).
		Media(file).
		Do()

	if err != nil {
		return fmt.Errorf(
			"set thumbnail: %w",
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
) error {

	var failed int

	for lang, path := range tracks {

		lang = strings.TrimSpace(lang)
		path = strings.TrimSpace(path)

		if lang == "" ||
			path == "" {

			continue
		}

		file, err := os.Open(path)

		if err != nil {
			failed++

			fmt.Printf(
				"⚠️ Caption %s: %v\n",
				lang,
				err,
			)

			continue
		}

		caption := &youtube.Caption{
			Snippet: &youtube.CaptionSnippet{
				VideoId: videoID,

				Language: lang,

				Name: "Moneyverse Dub",

				IsDraft: false,
			},
		}

		_, err = srv.
			Captions.
			Insert(
				[]string{
					"snippet",
				},
				caption,
			).
			Media(file).
			Do()

		file.Close()

		if err != nil {
			failed++

			fmt.Printf(
				"⚠️ Caption upload failed [%s]: %v\n",
				lang,
				err,
			)

			continue
		}

		fmt.Printf(
			"🌍 CAPTION UPLOADED: %s\n",
			lang,
		)
	}

	if failed > 0 {
		return fmt.Errorf(
			"%d caption(s) failed",
			failed,
		)
	}

	return nil
}

// ============================================================
// Set Public
// ============================================================

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

	_, err = srv.
		Videos.
		Update(
			[]string{
				"status",
			},
			&youtube.Video{
				Id: videoID,

				Status: &youtube.VideoStatus{
					PrivacyStatus: "public",

					SelfDeclaredMadeForKids: false,
				},
			},
		).
		Do()

	if err != nil {
		return fmt.Errorf(
			"set video public: %w",
			err,
		)
	}

	return nil
}

// ============================================================
// Pending State
// ============================================================

func recordPending(
	video PendingVideo,
) error {

	if video.YouTubeID == "" {
		return errors.New(
			"cannot schedule empty video ID",
		)
	}

	if video.PublishAt.IsZero() {
		return errors.New(
			"cannot schedule without publish time",
		)
	}

	if err := os.MkdirAll(
		filepath.Dir(pendingFile),
		0755,
	); err != nil {

		return fmt.Errorf(
			"create schedule directory: %w",
			err,
		)
	}

	var current pendingState

	data, err := os.ReadFile(
		pendingFile,
	)

	if err == nil &&
		len(
			strings.TrimSpace(
				string(data),
			),
		) > 0 {

		// الصيغة الجديدة.
		if err := json.Unmarshal(
			data,
			&current,
		); err != nil {

			// دعم الصيغة القديمة [].
			var legacy []PendingVideo

			if legacyErr :=
				json.Unmarshal(
					data,
					&legacy,
				); legacyErr != nil {

				return fmt.Errorf(
					"decode pending state: %w",
					err,
				)
			}

			current.Videos = legacy
		}
	}

	if current.Videos == nil {
		current.Videos = []PendingVideo{}
	}

	// منع إضافة نفس الفيديو مرتين.
	for i := range current.Videos {

		if current.Videos[i].YouTubeID ==
			video.YouTubeID {

			current.Videos[i] = video

			return savePending(
				current.Videos,
			)
		}
	}

	current.Videos = append(
		current.Videos,
		video,
	)

	return savePending(
		current.Videos,
	)
}

// ============================================================
// Save Pending
// ============================================================

func savePending(
	videos []PendingVideo,
) error {

	if err := os.MkdirAll(
		filepath.Dir(pendingFile),
		0755,
	); err != nil {

		return fmt.Errorf(
			"create schedule directory: %w",
			err,
		)
	}

	data, err := json.MarshalIndent(
		pendingState{
			Videos: videos,
		},
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

	temp := pendingFile + ".tmp"

	if err := os.WriteFile(
		temp,
		data,
		0644,
	); err != nil {

		return fmt.Errorf(
			"write pending temp file: %w",
			err,
		)
	}

	if err := os.Rename(
		temp,
		pendingFile,
	); err != nil {

		_ = os.Remove(temp)

		return fmt.Errorf(
			"replace pending file: %w",
			err,
		)
	}

	return nil
}

// ============================================================
// Save OAuth Token
// ============================================================

func saveToken(
	path string,
	token *oauth2.Token,
) error {

	if token == nil {
		return errors.New(
			"token is nil",
		)
	}

	if err := os.MkdirAll(
		filepath.Dir(path),
		0700,
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
			"encode token: %w",
			err,
		)
	}

	temp := path + ".tmp"

	if err := os.WriteFile(
		temp,
		data,
		0600,
	); err != nil {
		return fmt.Errorf(
			"write token: %w",
			err,
		)
	}

	if err := os.Rename(
		temp,
		path,
	); err != nil {

		_ = os.Remove(temp)

		return fmt.Errorf(
			"replace token: %w",
			err,
		)
	}

	return nil
}

// ============================================================
// Debug / Health Check
// ============================================================

func CheckCredentials() error {
	_, _, err := getOAuthConfig()

	if err != nil {
		return err
	}

	return nil
}

// ============================================================
// Utility
// ============================================================

func fileExists(path string) bool {
	if path == "" {
		return false
	}

	_, err := os.Stat(path)

	return err == nil
}

// منع compiler من اعتبار io غير مستخدم
// إذا تم توسيع الرفع لاحقًا لاستخدام io مباشرة.
var _ io.Reader
```
