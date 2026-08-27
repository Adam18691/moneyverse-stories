package ai

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"
)

type Provider struct {
	Name  string
	URL   string
	Model string
	Key   string
}

// providersList: يُبنى وقت التشغيل من الأسرار — fallback تلقائي
func providersList() []Provider {
	return []Provider{
		{"groq", "https://api.groq.com/openai/v1/chat/completions",
			"llama-3.3-70b-versatile", os.Getenv("GROQ_KEY")},
		{"gemini", "https://generativelanguage.googleapis.com/v1beta/openai/chat/completions",
			"gemini-1.5-flash", os.Getenv("GEMINI_KEY")},
		{"openrouter", "https://openrouter.ai/api/v1/chat/completions",
			"qwen/qwen-2.5-72b-instruct:free", os.Getenv("OPENROUTER_API_KEY")},
	}
}

// Chat: استدعاء موحّد مع fallback بين المصادر المجانية
func Chat(systemPrompt, userPrompt string) (string, error) {
	for _, p := range providersList() {
		if p.Key == "" {
			continue
		}
		resp, err := call(p, systemPrompt, userPrompt)
		if err != nil {
			fmt.Printf("   ⚠️ AI[%s] فشل → التالي\n", p.Name)
			continue
		}
		fmt.Printf("   🤖 AI via %s (%s)\n", p.Name, p.Model)
		return resp, nil
	}
	return "", fmt.Errorf("كل مزودي AI فشلوا")
}

func call(p Provider, systemMsg, userMsg string) (string, error) {
	body, _ := json.Marshal(map[string]interface{}{
		"model": p.Model,
		"messages": []map[string]string{
			{"role": "system", "content": systemMsg},
			{"role": "user", "content": userMsg},
		},
		"temperature": 0.7,
		"max_tokens":  2000,
	})
	req, _ := http.NewRequest("POST", p.URL, bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+p.Key)

	client := &http.Client{Timeout: 60 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var out struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return "", err
	}
	if len(out.Choices) == 0 {
		return "", fmt.Errorf("رد فارغ")
	}
	return out.Choices[0].Message.Content, nil
}

// ExtractJSON: تنظيف أي رد نموذج → JSON نقي (عام لكل الحزم)
func ExtractJSON(raw string) string {
	s := strings.ReplaceAll(raw, "```json", "")
	s = strings.ReplaceAll(s, "```", "")
	start := strings.Index(s, "{")
	end := strings.LastIndex(s, "}")
	if start >= 0 && end > start {
		return s[start : end+1]
	}
	return s
}
