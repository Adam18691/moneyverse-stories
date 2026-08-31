```go
//go:build vlypse
// +build vlypse

// Vlypse IDE - One Click Boilerplate for Go
// Project: moneyverse-stories
// Language: Go 1.24

package main

import (
	"encoding/json"
	"fmt"
	"os"
)

const (
	projectName = "moneyverse-stories"
	goVersion   = "1.24"
	moduleName  = "github.com/Adam18691/moneyverse-stories"
)

// VlypseConfig هي إعدادات مشروع Vlypse.
type VlypseConfig struct {
	Project   string   `json:"project"`
	Framework string   `json:"framework"`
	Language  string   `json:"language"`
	GoVersion string   `json:"go_version"`
	Module    string   `json:"module"`
	Entry     string   `json:"entry"`
	Plugins   []string `json:"plugins"`
}

// init يتم تشغيله فقط عند البناء باستخدام:
// -tags vlypse
func init() {
	config := VlypseConfig{
		Project:   projectName,
		Framework: "Go + FFmpeg + Melt + YouTube API",
		Language:  "Go",
		GoVersion: goVersion,
		Module:    moduleName,
		Entry:     "main.go",

		Plugins: []string{
			"ffmpeg",
			"melt",
			"youtube-api",
			"analytics",
			"432hz-music",
			"cinematic-angles",
		},
	}

	if os.Getenv("VLYPSE_MODE") != "generate" {
		return
	}

	fmt.Println(
		"⚡ Vlypse IDE - Generating project configuration...",
	)

	fmt.Printf(
		"Project: %s\n",
		config.Project,
	)

	fmt.Printf(
		"Framework: %s\n",
		config.Framework,
	)

	fmt.Printf(
		"Language: %s %s\n",
		config.Language,
		config.GoVersion,
	)

	fmt.Printf(
		"Module: %s\n",
		config.Module,
	)

	fmt.Printf(
		"Plugins: %v\n",
		config.Plugins,
	)

	if err := generateGoBoilerplate(config); err != nil {
		fmt.Printf(
			"❌ Vlypse generation failed: %v\n",
			err,
		)

		os.Exit(1)
	}

	fmt.Println(
		"✅ Vlypse project configuration generated successfully",
	)

	os.Exit(0)
}

// generateGoBoilerplate ينشئ ملفات إعداد Vlypse
// دون استبدال go.mod الموجود في المشروع.
func generateGoBoilerplate(
	config VlypseConfig,
) error {

	// =========================
	// vlypse.json
	// =========================

	data, err := json.MarshalIndent(
		config,
		"",
		"  ",
	)

	if err != nil {
		return fmt.Errorf(
			"encode Vlypse config: %w",
			err,
		)
	}

	if err := writeFile(
		"vlypse.json",
		append(data, '\n'),
	); err != nil {
		return err
	}

	// =========================
	// Makefile
	// =========================

	makefile := `# Moneyverse Stories
# Vlypse / Go 1.24

.PHONY: build test fmt vet run vlypse clean

build:
	go build -trimpath -o moneyverse-stories .

run:
	go run .

test:
	go test ./...

fmt:
	gofmt -w $$(find . -type f -name '*.go' -not -path './vendor/*')

vet:
	go vet ./...

vlypse:
	VLYPSE_MODE=generate go run -tags vlypse boilerplate.go

clean:
	rm -f moneyverse-stories
`

	if err := writeFile(
		"Makefile",
		[]byte(makefile),
	); err != nil {
		return err
	}

	// =========================
	// .vlypse/config.json
	// =========================

	vlypseDir := ".vlypse"

	if err := os.MkdirAll(
		vlypseDir,
		0755,
	); err != nil {
		return fmt.Errorf(
			"create .vlypse directory: %w",
			err,
		)
	}

	vlypseConfig := `{
  "project": "moneyverse-stories",
  "language": "Go",
  "go_version": "1.24",
  "module": "github.com/Adam18691/moneyverse-stories",
  "entry": "main.go",
  "build_cmd": "go build -trimpath -o moneyverse-stories .",
  "run_cmd": "go run .",
  "test_cmd": "go test ./...",
  "format_cmd": "gofmt -w .",
  "vet_cmd": "go vet ./...",
  "one_click": true
}
`

	if err := writeFile(
		".vlypse/config.json",
		[]byte(vlypseConfig),
	); err != nil {
		return err
	}

	// =========================
	// Generation Summary
	// =========================

	fmt.Println(
		"📦 Generated:",
	)

	fmt.Println(
		"   ✅ vlypse.json",
	)

	fmt.Println(
		"   ✅ Makefile",
	)

	fmt.Println(
		"   ✅ .vlypse/config.json",
	)

	fmt.Println(
		"ℹ️ Existing go.mod was preserved",
	)

	return nil
}

// writeFile يكتب الملف مع معالجة الأخطاء.
func writeFile(
	path string,
	data []byte,
) error {

	if err := os.WriteFile(
		path,
		data,
		0644,
	); err != nil {
		return fmt.Errorf(
			"write %s: %w",
			path,
			err,
		)
	}

	return nil
}
```
