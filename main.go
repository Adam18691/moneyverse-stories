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

// ===== Models - مفيش مفاتيح هاردكود =====
var ModelMap = map[string]string{
	"title": "qwen/qwen-2.5-72b-instruct:free",
	"desc": "deepseek/deepseek-chat:free",
	"tags": "meta-llama/llama-3.3-70b-instruct:free",
	"script": "moonshotai/kimi-k2:free",
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
		// gitleaks:allow
		keys = []string{os.Getenv("OPENROUTER_API_KEY"), os.Getenv("OPENROUTER_API_KEY_2")}
	}
	k1 := os.Getenv("OPENROUTER_API_KEY")
	if k1 == "" {
		return "", errors.New("OPENROUTER_API_KEY missing in secrets")
	}
	k := keys[keyIdx%len(keys)]
	keyIdx++
	if k == "" {
		k = k1
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

	model := ModelMap[task]
	if model == "" {
		model = "z-ai/glm-4.5:free"
	}
	log.Printf("Model %s -> %s", task, model)

	resp, err := client.CreateChatCompletion(context.Background(), openai.ChatCompletionRequest{
		Model: model,
		Messages: []openai.ChatCompletionMessage{
			{Role: openai.ChatMessageRoleSystem, Content: "انت طيبات الدكتور ضياء العوضي - هادئ واقعي بسيط مطمئن"},
			{Role: openai.ChatMessageRoleUser, Content: prompt},
		},
		MaxTokens: 400,
	})
	if err!= nil {
		return "", fmt.Errorf("AI %s: %w", task, err)
	}
	if len(resp.Choices) == 0 {
		return "", errors.New("empty AI response")
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
	key, err := getKey()
	if err!= nil {
		http.Error(w, err.Error(), 500)
		return
	}
	req, err := http.NewRequest("POST", "https://openrouter.ai/api/v1/chat/completions", bytes.NewBuffer([]byte(newBody)))
	if err!= nil {
		http.Error(w, err.Error(), 500)
		return
	}
	req.Header.Set("Authorization", "Bearer "+key)
	req.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err!= nil {
		http.Error(w, err.Error(), 502)
		return
	}
	defer func() {
		if err := resp.Body.Close(); err!= nil {
			log.Printf("close body: %v", err)
		}
	}()
	w.Header().Set("Content-Type", "application/json")
	if _, err := io
