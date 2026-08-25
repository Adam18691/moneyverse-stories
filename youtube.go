package youtube

import (
	"time"
	"google.golang.org/api/youtube/v3"
)

type Meta struct {
	Title       string            `json:"title"`
	Description string            `json:"description"`
	Tags        []string          `json:"tags"`
	LangTracks  map[string]string `json:"-"` // lang -> .vtt file
	ThumbPath   string            `json:"-"`
	PublishAt   *time.Time        `json:"publishAt,omitempty"` // ⏰ الجديد
}

// Upload: رفع مع جدولة ذكية
func Upload(videoPath string, m Meta) (string, error) {
	client, err := getClient()
	if err != nil { return "", err }
	srv, err := youtube.NewService(context.Background(), option.WithHTTPClient(client))
	if err != nil { return "", err }

	status := &youtube.VideoStatus{
		SelfDeclaredMadeForKids: false,
	}

	// ⏰ منطق الجدولة:
	// فيوجد publishAt → private الآن + public تلقائي لاحقاً
	// لا يوجد → public فوراً
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

	call := srv.Videos.Insert([]string{"snippet", "status"}, vid)
	file, _ := os.Open(videoPath)
	resp, err := call.Media(file).Do()
	if err != nil { return "", err }
	id := resp.Id

	// ثامبنيل + captions كما سبق...
	setThumbAndCaptions(srv, id, m)

	if m.PublishAt != nil {
		fmt.Printf("⏰ SCHEDULED: https://youtu.be/%s → goes PUBLIC at %s\n",
			id, m.PublishAt.Format("15:04 MST"))
	} else {
		fmt.Printf("✅ UPLOADED PUBLIC: https://youtu.be/%s\n", id)
	}
	return id, nil
}
