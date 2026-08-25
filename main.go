package main

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"image"
	"image/color"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"strings"
	"sync"
	"time"

	"github.com/disintegration/imaging"
	"github.com/kkdai/youtube/v2"
	"github.com/sashabaranov/go-openai"
	"golang.org/x/oauth2"
	"google.golang.org/api/option"
	ytapi "google.golang.org/api/youtube/v3"
)

var ModelMap = map[string]string{
	"title": "qwen/qwen-2.5-72b-instruct:free", "desc": "deepseek/deepseek-chat:free",
	"tags": "meta-llama/llama-3.3-70b-instruct:free", "script": "moonshotai/kimi-k2:free",
}
var (
	keys []string
	keyIdx int
	mu sync.Mutex
)

func getKey() (string, error) {
	mu.Lock()
	defer mu.Unlock()
	if len(keys) == 0 {
		keys = []string{os.Getenv("OPENROUTER_API_KEY"), os.Getenv("OPENROUTER_API_KEY_2")}
	}
	if os.Getenv("OPENROUTER_API_KEY") == "" {
		return "", errors.New("OPENROUTER_API_KEY missing")
	}
	k := keys[keyIdx%len(keys)]
	keyIdx++
	if k == "" {
		k = os.Getenv("OPENROUTER_API_KEY")
	}
	return k, nil
}

func Generate(task, prompt string) (string, error) {
	key, err := getKey()
	if err!= nil {
		return "", err
	}
	cfg := openai.DefaultConfig(key)
	cfg.BaseURL = "https://openrouter.ai/api/v1"
	client := openai.NewClientWithConfig(cfg)
	resp, err := client.CreateChatCompletion(context.Background(), openai.ChatCompletionRequest{
		Model: ModelMap[task],
		Messages: []openai.ChatCompletionMessage{
			{Role: openai.ChatMessageRoleSystem, Content: "انت طيبات الدكتور ضياء - هادئ"},
			{Role: openai.ChatMessageRoleUser, Content: prompt},
		},
		MaxTokens: 400,
	})
	if err!= nil {
		return "", fmt.Errorf("AI %s: %w", task, err)
	}
	if len(resp.Choices) == 0 {
		return "", errors.New("empty AI")
	}
	return strings.TrimSpace(resp.Choices[0].Message.Content), nil
}

func gatewayHandler(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err!= nil {
		http.Error(w, err.Error(), 400)
		return
	}
	s := string(body)
	model := "z-ai/glm-4.5:free"
	for k, v := range ModelMap {
		if strings.Contains(s, k) {
			model = v
			break
		}
	}
	newBody := strings.ReplaceAll(s, `"model":"claude-3-5-sonnet"`, fmt.Sprintf(`"model":"%s"`, model))
	key, _ := getKey()
	req, _ := http.NewRequest("POST", "https://openrouter.ai/api/v1/chat/completions", bytes.NewBuffer([]byte(newBody)))
	req.Header.Set("Authorization", "Bearer "+key)
	req.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err!= nil {
		http.Error(w, err.Error(), 502)
		return
	}
	defer resp.Body.Close()
	w.Header().Set("Content-Type", "application/json")
	io.Copy(w, resp.Body)
}

func DownloadYTGo(ytURL, outFile string) error {
	fmt.Printf("⬇️ تحميل يوتيوب: %s\n", ytURL)
	client := youtube.Client{}
	video, err := client.GetVideo(ytURL)
	if err == nil {
		formats := video.Formats.WithAudioChannels()
		if len(formats) > 0 {
			stream, _, err := client.GetStream(video, &formats[0])
			if err == nil {
				defer stream.Close()
				f, err := os.Create(outFile)
				if err == nil {
					_, err = io.Copy(f, stream)
					f.Close()
					if err == nil {
						exec.Command("ffmpeg", "-y", "-i", outFile, "-ss", "00:00:05", "-t", "10", "-vn", "diaa_sample.wav").Run()
						return nil
					}
				}
			}
		}
	}
	cmd := exec.Command("yt-dlp", "-x", "--audio-format", "wav", "-o", "downloaded.%(ext)s", ytURL)
	if out, err := cmd.CombinedOutput(); err!= nil {
		return fmt.Errorf("yt-dlp %v %s", err, string(out))
	}
	exec.Command("ffmpeg", "-y", "-i", "downloaded.wav", "-ss", "00:00:05", "-t", "10", "diaa_sample.wav").Run()
	exec.Command("sh", "-c", "rm -f downloaded.*").Run()
	return nil
}

func MakeThumbnail(title string) error {
	p := url.QueryEscape("Calm Dr Diaa " + title)
	u := fmt.Sprintf("https://image.pollinations.ai/prompt/%s?width=1280&height=720&model=flux&nologo=true", p)
	resp, err := http.Get(u)
	if err!= nil {
		return err
	}
	defer resp.Body.Close()
	f, err := os.Create("thumb.jpg")
	if err!= nil {
		return err
	}
	if _, err := io.Copy(f, resp.Body); err!= nil {
		f.Close()
		return err
	}
	f.Close()
	src, err := imaging.Open("thumb.jpg")
	if err!= nil {
		return err
	}
	dst := imaging.AdjustContrast(src, 5)
	overlay := imaging.New(1280, 720, color.NRGBA{245, 248, 255, 25})
	dst = imaging.Overlay(dst, overlay, image.Pt(0, 0), 1.0)
	return imaging.Save(dst, "thumb_pro.jpg")
}

func MakeVoice(text string) error {
	os.WriteFile("script.txt", []byte(text), 0644)
	if _, err := os.Stat("diaa_sample.wav"); err == nil {
		cmd := exec.Command("python3", "-m", "TTS", "--text", text, "--model_name", "tts_models/multilingual/multi-dataset/xtts_v2", "--speaker_wav", "diaa_sample.wav", "--language_idx", "ar", "--out_path", "voice.mp3")
		if err := cmd.Run(); err == nil {
			return nil
		}
	}
	cmd := exec.Command("edge-tts", "--voice", "ar-EG-ShakirNeural", "--rate", "-18%", "--text", text, "--write-media", "voice.mp3")
	out, err := cmd.CombinedOutput()
	if err!= nil {
		return fmt.Errorf("edge-tts %w %s", err, string(out))
	}
	return nil
}

func MakeVideo() error {
	cmd := exec.Command("ffmpeg", "-y", "-loop", "1", "-i", "thumb_pro.jpg", "-i", "voice.mp3", "-filter_complex", "[0:v]scale=1280:720,zoompan=z='if(lte(zoom,1.08),zoom+0.0004,1.08)':d=1:x='iw/2-(iw/zoom/2)':y='ih/2-(ih/zoom/2)',fps=24[v]", "-map", "[v]", "-map", "1:a", "-c:v", "libx264", "-preset", "slow", "-crf", "21", "-shortest", "final.mp4")
	out, err := cmd.CombinedOutput()
	if err!= nil {
		return fmt.Errorf("ffmpeg %w %s", err, string(out))
	}
	return nil
}

func UploadYT(file, title, desc string) error {
	ctx := context.Background()
	conf := &oauth2.Config{ClientID: os.Getenv("YOUTUBE_CLIENT_ID"), ClientSecret: os.Getenv("YOUTUBE_CLIENT_SECRET"), Endpoint: oauth2.Endpoint{AuthURL: "https://accounts.google.com/o/oauth2/auth", TokenURL: "https://oauth2.googleapis.com/token"}, Scopes: []string{ytapi.YoutubeUploadScope}}
	tok := &oauth2.Token{RefreshToken: os.Getenv("YOUTUBE_REFRESH_TOKEN")}
	client := conf.Client(ctx, tok)
	srv, err := ytapi.NewService(ctx, option.WithHTTPClient(client))
	if err!= nil {
		return err
	}
	f, err := os.Open(file)
	if err!= nil {
		return err
	}
	defer f.Close()
	vid := &ytapi.Video{Snippet: &ytapi.VideoSnippet{Title: title, Description: desc, CategoryId: "27"}, Status: &ytapi.VideoStatus{PrivacyStatus: "public"}}
	_, err = srv.Videos.Insert([]string{"snippet,status"}, vid).Media(f).Do()
	return err
}

func main() {
	// ===== فحص الامان اولا - التلات ادوات من الصور =====
	if err := RunAllSecurityChecks(); err!= nil {
		log.Fatalf("🔒 Security failed: %v", err)
	}

	go func() {
		http.HandleFunc("/", gatewayHandler)
		http.HandleFunc("/v1/messages", gatewayHandler)
		port := os.Getenv("PORT")
		if port == "" {
			port = "8787"
		}
		log.Printf("🔥 Gateway on :%s", port)
		http.ListenAndServe(":"+port, nil)
	}()
	time.Sleep(1 * time.Second)

	if ytLink := os.Getenv("SOURCE_YT_URL"); ytLink!= "" {
		_ = DownloadYTGo(ytLink, "source.m4a")
	}

	title, _ := Generate("title", "عنوان هادئ عن القولون")
	script, _ := Generate("script", "سكريبت هادئ عن: "+title)
	desc, _ := Generate("desc", "وصف هادئ عن: "+title)
	tags, _ := Generate("tags", "هاشتاج #ضياء_العوضي")

	log.Println("📌", title)
	if err := MakeVoice(script); err!= nil {
		log.Fatalf("voice: %v", err)
	}
	if err := MakeThumbnail(title); err!= nil {
		log.Fatalf("thumb: %v", err)
	}
	if err := MakeVideo(); err!= nil {
		log.Fatalf("video: %v", err)
	}
	if err := UploadYT("final.mp4", title, desc+"\n\n"+tags); err!= nil {
		log.Fatalf("upload: %v", err)
	}
	log.Println("✅ Done secure + realistic")
}
