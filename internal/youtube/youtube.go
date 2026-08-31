```go
package youtube

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
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
	clientIDFile     = "credentials/client_id.txt"
	clientSecretFile = "credentials/client_secret.txt"
	refreshTokenFile = "credentials/refresh_token.txt"
	tokenFile        = "credentials/token.json"

	pendingFile = "schedule/pending.json"

	youtubeUploadScope = "https://www.googleapis.com/auth/youtube.upload"
	youtubeForceScope  = "https://www.googleapis.com/auth/youtube.force-ssl"
)

var serviceMu sync.Mutex

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
// Credentials
// ============================================================

func readCredential(path string) (string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf(
			"read credential %s: %w",
			path,
			err,
		)
	}

	value := strings.TrimSpace(string(data))

	if value == "" {
		return "", fmt.Errorf(
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
	clientID, err := readCredential(clientIDFile)
	if err != nil {
		return nil, "", err
	}

	clientSecret, err := readCredential(clientSecretFile)
	if err != nil {
		return nil, "", err
	}

	refreshToken, err := readCredential(refreshTokenFile)
	if err != nil {
		return nil, "", err
	}

	cfg := &oauth2.Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		Endpoint:     google.Endpoint,
		Scopes: []string{
			youtubeUploadScope,
			youtubeForceScope,
		},
	}

	return cfg, refreshToken, nil
}

// ============================================================
// YouTube Service
// ============================================================

func newService() (*youtube.Service, error) {
	serviceMu.Lock()
	defer serviceMu.Unlock()

	cfg, refreshToken, err := getOAuthConfig()
	if err != nil {
		return nil, err
	}

	ctx := context.Background()

	refreshTokenObj := &oauth2.Token{
		RefreshToken: refreshToken,
	}

	tokenSource := cfg.TokenSource(
		ctx,
		refreshTokenObj,
	)

	token, err := tokenSource.Token()
	if err != nil {
		return nil, fmt.Errorf(
			"obtain YouTube access token: %w",
			err,
		)
	}

	if err := saveToken(tokenFile, token); err != nil {
		fmt.Printf(
			"⚠️ token cache warning: %v\n",
			err,
		)
	}

	client := cfg.Client(ctx, token)

	service, err := youtube.NewService(
		ctx,
		option.WithHTTPClient(client),
	)
	if err != nil {
		return nil, fmt.Errorf(
			"create YouTube service: %w",
			err,
		)
	}

	return service, nil
}

// ============================================================
// Upload
// ============================================================

func Upload(
	videoPath string,
	m Meta,
) (string, error) {

	videoPath = strings.TrimSpace(videoPath)

	if videoPath == "" {
		return "",
			errors.New("video path is empty")
	}

	info, err := os.Stat(videoPath)
	if err != nil {
		return "",
			fmt.Errorf(
				"video file unavailable: %w",
				err,
			)
	}

	if info.IsDir() {
		return "",
			fmt.Errorf(
				"video path is a directory: %s",
				videoPath,
			)
	}

	m.Title = strings.TrimSpace(m.Title)

	if m.Title == "" {
		return "",
			errors.New("YouTube title is empty")
	}

	srv, err := newService()
	if err != nil {
		return "", err
	}

	// ========================================================
	// Privacy
	// ========================================================

	privacy := "public"

	if m.PublishAt != nil &&
		!m.PublishAt.IsZero() {

		if m.PublishAt.UTC().After(
			time.Now().UTC(),
		) {
			privacy = "private"
		}
	}

	// ========================================================
	// Video Metadata
	// ========================================================

	video := &youtube.Video{
		Snippet: &youtube.VideoSnippet{
			Title:       m.Title,
			Description: m.Description,
			Tags:        cleanTags(m.Tags),
			CategoryId:  "27",
		},

		Status: &youtube.VideoStatus{
			PrivacyStatus: privacy,

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
		"📤 Uploading: %s\n",
		videoPath,
	)

	// ========================================================
	// Upload
	// ========================================================

	resp, err := srv.
		Videos.
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

	if resp == nil {
		return "",
			errors.New(
				"YouTube returned nil response",
			)
	}

	videoID := strings.TrimSpace(resp.Id)

	if videoID == "" {
		return "",
			errors.New(
				"YouTube returned an empty video ID",
			)
	}

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
	// Captions / Language Tracks
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
	// Scheduled Publishing
	// ========================================================

	if m.PublishAt != nil &&
		!m.PublishAt.IsZero() {

		publishAt := m.PublishAt.UTC()

		if publishAt.After(
			time.Now().UTC(),
		) {

			err := recordPending(
				PendingVideo{
					YouTubeID: videoID,
					Title:     m.Title,
					PublishAt: publishAt,
					Published: false,
					Attempts:  0,
				},
			)

			if err != nil {
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
					"2006-01-02 15:04:05",
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

	if _, err := os.Stat(path); err != nil {
		return fmt.Errorf(
			"thumbnail unavailable: %w",
			err,
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

	if len(tracks) == 0 {
		return nil
	}

	var failed int

	for lang, path := range tracks {

		lang = strings.TrimSpace(lang)
		path = strings.TrimSpace(path)

		if lang == "" || path == "" {
			continue
		}

		if _, err := os.Stat(path); err != nil {
			failed++

			fmt.Printf(
				"⚠️ Caption file missing [%s]: %v\n",
				lang,
				err,
			)

			continue
		}

		file, err := os.Open(path)
		if err != nil {
			failed++

			fmt.Printf(
				"⚠️ Caption open failed [%s]: %v\n",
				lang,
				err,
			)

			continue
		}

		caption := &youtube.Caption{
			Snippet: &youtube.CaptionSnippet{
				VideoId:  videoID,
				Language: lang,
				Name:     "Moneyverse Dub",
				IsDraft:  false,
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

		closeErr := file.Close()

		if err != nil {
			failed++

			fmt.Printf(
				"⚠️ Caption upload failed [%s]: %v\n",
				lang,
				err,
			)

			continue
		}

		if closeErr != nil {
			fmt.Printf(
				"⚠️ Caption close warning [%s]: %v\n",
				lang,
				closeErr,
			)
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

	videoID = strings.TrimSpace(videoID)

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

	fmt.Printf(
		"✅ VIDEO PUBLIC: https://youtu.be/%s\n",
		videoID,
	)

	return nil
}

// ============================================================
// Pending State
// ============================================================

func recordPending(
	video PendingVideo,
) error {

	video.YouTubeID =
		strings.TrimSpace(video.YouTubeID)

	if video.YouTubeID == "" {
		return errors.New(
			"cannot schedule empty video ID",
		)
	}

	video.Title =
		strings.TrimSpace(video.Title)

	if video.Title == "" {
		video.Title = video.YouTubeID
	}

	if video.PublishAt.IsZero() {
		return errors.New(
			"cannot schedule without publish time",
		)
	}

	video.PublishAt =
		video.PublishAt.UTC()

	if err := os.MkdirAll(
		filepath.Dir(pendingFile),
		0755,
	); err != nil {

		return fmt.Errorf(
			"create schedule directory: %w",
			err,
		)
	}

	current, err := loadPending()

	if err != nil {
		return err
	}

	// ========================================================
	// Update Existing
	// ========================================================

	for i := range current {

		if current[i].YouTubeID ==
			video.YouTubeID {

			// الاحتفاظ بعدد المحاولات الحالي
			// إذا كان السجل موجودًا.
			video.Attempts =
				current[i].Attempts

			video.Published =
				current[i].Published

			current[i] = video

			return savePending(current)
		}
	}

	// ========================================================
	// Append New
	// ========================================================

	current = append(
		current,
		video,
	)

	return savePending(current)
}

// ============================================================
// Load Pending
// ============================================================

func loadPending() ([]PendingVideo, error) {

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

	if len(
		strings.TrimSpace(
			string(data),
		),
	) == 0 {

		return []PendingVideo{}, nil
	}

	// ========================================================
	// New Format
	// ========================================================

	var state pendingState

	if err := json.Unmarshal(
		data,
		&state,
	); err == nil {

		if state.Videos == nil {
			state.Videos = []PendingVideo{}
		}

		return state.Videos, nil
	}

	// ========================================================
	// Legacy Format
	// ========================================================

	var legacy []PendingVideo

	if err := json.Unmarshal(
		data,
		&legacy,
	); err != nil {

		return nil,
			fmt.Errorf(
				"invalid pending JSON: %w",
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

	tempFile :=
		pendingFile + ".tmp"

	if err := os.WriteFile(
		tempFile,
		data,
		0644,
	); err != nil {

		return fmt.Errorf(
			"write temporary pending state: %w",
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
// Save OAuth Token
// ============================================================

func saveToken(
	path string,
	token *oauth2.Token,
) error {

	if token == nil {
		return errors.New(
			"OAuth token is nil",
		)
	}

	if err := os.MkdirAll(
		filepath.Dir(path),
		0700,
	); err != nil {

		return fmt.Errorf(
			"create token directory: %w",
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

	tempFile := path + ".tmp"

	if err := os.WriteFile(
		tempFile,
		data,
		0600,
	); err != nil {

		return fmt.Errorf(
			"write OAuth token: %w",
			err,
		)
	}

	if err := os.Rename(
		tempFile,
		path,
	); err != nil {

		_ = os.Remove(tempFile)

		return fmt.Errorf(
			"replace OAuth token: %w",
			err,
		)
	}

	return nil
}

// ============================================================
// Credential Health Check
// ============================================================

func CheckCredentials() error {
	_, _, err := getOAuthConfig()

	if err != nil {
		return err
	}

	return nil
}

// ============================================================
// Helpers
// ============================================================

func cleanTags(
	tags []string,
) []string {

	if len(tags) == 0 {
		return nil
	}

	result := make(
		[]string,
		0,
		len(tags),
	)

	seen := make(
		map[string]struct{},
	)

	for _, tag := range tags {

		tag = strings.TrimSpace(tag)

		if tag == "" {
			continue
		}

		key := strings.ToLower(tag)

		if _, exists := seen[key]; exists {
			continue
		}

		seen[key] = struct{}{}

		result = append(
			result,
			tag,
		)
	}

	if len(result) == 0 {
		return nil
	}

	return result
}
```
