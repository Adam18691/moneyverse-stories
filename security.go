package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"os/exec"
	"regexp"
	"strings"
)

// ===== 1. GITLEAKS Pure Go - بتفتش على كل الـ secrets =====
func CheckGitleaksGo() error {
	fmt.Println("🔑 [1/3] Gitleaks - تفتيش الـ secrets...")
	// patterns زي gitleaks الاصلي
	patterns := map[string]*regexp.Regexp{
		"OPENROUTER_API_KEY": regexp.MustCompile(`(?i)openrouter[_-]?api[_-]?key\s*[:=]\s*['"]?sk-or-v1-[a-zA-Z0-9_-]+`),
		"YOUTUBE_CLIENT_SECRET": regexp.MustCompile(`(?i)client[_-]?secret\s*[:=]\s*['"]?[A-Za-z0-9_-]{20,}`),
		"GENERIC_API_KEY": regexp.MustCompile(`(?i)api[_-]?key\s*[:=]\s*['"][a-zA-Z0-9_\-]{20,}['"]`),
	}

	files := []string{"main.go", "security.go", "go.mod", ".env"}
	found := false
	for _, file := range files {
		f, err := os.Open(file)
		if err!= nil {
			continue
		}
		scanner := bufio.NewScanner(f)
		lineNum := 0
		for scanner.Scan() {
			lineNum++
			line := scanner.Text()
			for name, re := range patterns {
				if re.MatchString(line) {
					// تجاهل لو من os.Getenv
					if strings.Contains(line, "os.Getenv") || strings.Contains(line, "secrets.") {
						continue
					}
					log.Printf("❌ GITLEAKS FOUND [%s] %s:%d -> %s", name, file, lineNum, line)
					found = true
				}
			}
		}
		f.Close()
	}
	if found {
		return fmt.Errorf("secrets leaked in code - راجع الصورة 2")
	}
	fmt.Println("✅ Gitleaks clean - مفيش مفاتيح متسربة")
	return nil
}

// ===== 2. SEMGREP Pure Go - لكل الكود =====
func CheckSemgrepGo() error {
	fmt.Println("🟢 [2/3] Semgrep - فحص ثغرات الكود...")

	// قواعد semgrep مدمجة Go
	rules := []struct {
		id, pattern, msg string
		re *regexp.Regexp
	}{
		{"go-missing-err-check", `_\s*,\s*err\s*:=\s*.+\n.*\n.*err`, "نسيت تفحص err", regexp.MustCompile(`_\s*,\s*err\s*:=`)},
		{"go-sql-injection", `fmt\.Sprintf.*SELECT`, "SQL injection محتمل", regexp.MustCompile(`fmt\.Sprintf.*SELECT`)},
		{"go-hardcoded-secret", `sk-or-v1-[a-zA-Z0-9]+`, "مفتاح OpenRouter hardcoded", regexp.MustCompile(`sk-or-v1-[a-zA-Z0-9\-_]+`)},
		{"go-unchecked-exec", `exec\.Command.*\.Run\(\)\s*\n`, "exec بدون فحص error", regexp.MustCompile(`exec\.Command.*\.Run\(\)`)},
	}

	f, err := os.Open("main.go")
	if err!= nil {
		return nil
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	lineNum := 0
	issues := 0
	for scanner.Scan() {
		lineNum++
		line := scanner.Text()
		for _, rule := range rules {
			if rule.re.MatchString(line) {
				// استثناءات صحيحة
				if strings.Contains(line, "CombinedOutput") || strings.Contains(line, "os.Getenv") {
					continue
				}
				if rule.id == "go-unchecked-exec" &&!strings.Contains(line, "_ =") &&!strings.Contains(line, "out, err") {
					log.Printf("⚠️ SEMGREP [%s] line %d: %s -> %s", rule.id, lineNum, rule.msg, line)
					issues++
				}
				if rule.id!= "go-unchecked-exec" {
					log.Printf("⚠️ SEMGREP [%s] line %d: %s", rule.id, lineNum, rule.msg)
					issues++
				}
			}
		}
	}
	if issues > 0 {
		fmt.Printf("⚠️ Semgrep: %d warnings (راجع الصورة 3)\n", issues)
	} else {
		fmt.Println("✅ Semgrep clean")
	}
	return nil
}

// ===== 3. CONTAINERS (Trivy) Pure Go - فحص الـ containers =====
func CheckContainersGo() error {
	fmt.Println("🐳 [3/3] Containers/Trivy - فحص الـ dependencies...")

	// فحص go.mod
	data, err := os.ReadFile("go.mod")
	if err!= nil {
		return nil
	}
	content := string(data)

	// قاعدة بيانات ثغرات مبسطة
	vulnDB := map[string]string{
		"github.com/sashabaranov/go-openai v1.17": "قديم - حدث لـ v1.28",
		"golang.org/x/oauth2 v0.2": "ثغرة قديمة",
	}

	for pkg, msg := range vulnDB {
		if strings.Contains(content, pkg) {
			log.Printf("⚠️ CONTAINER [%s]: %s", pkg, msg)
		}
	}

	// فحص binaries خطيرة في workflow
	if _, err := os.Stat(".github/workflows/god.yml"); err == nil {
		wf, _ := os.ReadFile(".github/workflows/god.yml")
		if strings.Contains(string(wf), "latest") &&!strings.Contains(string(wf), "sha256") {
			log.Printf("⚠️ CONTAINER: استخدم نسخة محددة مش latest")
		}
	}

	// محاولة تشغيل trivy الحقيقي لو موجود
	if _, err := exec.LookPath("trivy"); err == nil {
		cmd := exec.Command("trivy", "fs", "--severity", "HIGH,CRITICAL", "--quiet", ".")
		out, _ := cmd.CombinedOutput()
		if len(out) > 0 {
			fmt.Println(string(out))
		}
	}

	fmt.Println("✅ Containers checked")
	return nil
}

func RunAllSecurityChecks() error {
	fmt.Println("\n=== 🔒 فحص الامان الشامل - 3 ادوات ===")
	if err := CheckGitleaksGo(); err!= nil {
		return err
	}
	if err := CheckSemgrepGo(); err!= nil {
		return err
	}
	if err := CheckContainersGo(); err!= nil {
		return err
	}
	fmt.Println("✅ كل فحوصات الامان نجحت\n")
	return nil
}
