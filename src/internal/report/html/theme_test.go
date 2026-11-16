package html

import (
	"os"
	"regexp"
	"strings"
	"testing"
)

func TestNoRawColorsOutsideTheme(t *testing.T) {
	colorLiteral := regexp.MustCompile(`(?i)(?:^|[^&\w])#[0-9a-f]{3,8}\b|\brgba?\s*\(|\bhsla?\s*\(`)
	files := map[string]string{
		"assets/style.css":     styleCSS,
		"assets/app.js":        appJS,
		"templates/components": mustReadTemplate(t, "templates/components.gohtml"),
		"templates/report":     mustReadTemplate(t, "templates/report.gohtml"),
		"templates/index":      mustReadTemplate(t, "templates/index.gohtml"),
	}
	for name, body := range files {
		if match := colorLiteral.FindString(body); match != "" {
			t.Errorf("%s contains a raw colour %q; add a token to theme.css and docs/theme.md", name, match)
		}
	}
}

func TestThemeTokensDocumented(t *testing.T) {
	doc, err := os.ReadFile("../../../../docs/theme.md")
	if err != nil {
		t.Fatalf("read theme guideline: %v", err)
	}
	text := string(doc)
	tokenRe := regexp.MustCompile(`(--[a-z0-9-]+)\s*:`)
	seen := map[string]bool{}
	for _, match := range tokenRe.FindAllStringSubmatch(themeCSS, -1) {
		token := match[1]
		if strings.HasPrefix(token, "--nc-") || seen[token] {
			continue
		}
		seen[token] = true
		if !strings.Contains(text, token) {
			t.Errorf("token %s is missing from docs/theme.md", token)
		}
	}
	if len(seen) < 50 {
		t.Fatalf("theme token scan found only %d tokens", len(seen))
	}
}

func TestRequiredThemeTokens(t *testing.T) {
	for _, token := range []string{
		"--bg-0", "--panel", "--text", "--muted", "--dim",
		"--accent", "--accent-strong", "--accent-ink",
		"--team-ct-fill", "--team-ct-text", "--team-ct-ink",
		"--team-t-fill", "--team-t-text", "--team-t-ink",
		"--danger", "--danger-text", "--warning", "--warning-text",
		"--success", "--success-text", "--info", "--info-text",
		"--round-a", "--round-b", "--focus",
	} {
		if !strings.Contains(themeCSS, token+":") {
			t.Errorf("required theme token %s is missing", token)
		}
	}
}

func TestEmbeddedAssetBudget(t *testing.T) {
	total := 0
	for _, uri := range []string{
		string(reportAssets.BrandMark),
		string(reportAssets.Logo),
		string(reportAssets.Favicon),
		string(reportAssets.AppleTouch),
	} {
		total += len(uri)
	}
	if total > 100*1024 {
		t.Fatalf("embedded brand assets use %d bytes, budget is 100 KiB", total)
	}
}

func mustReadTemplate(t *testing.T, name string) string {
	t.Helper()
	b, err := templatesFS.ReadFile(name)
	if err != nil {
		t.Fatalf("read %s: %v", name, err)
	}
	return string(b)
}
