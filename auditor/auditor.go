package auditor

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"html/template"
	"os"
	"strings"

	"github.com/anthropics/anthropic-sdk-go"
	"github.com/anthropics/anthropic-sdk-go/option"
	"github.com/chromedp/cdproto/page"
	"github.com/chromedp/chromedp"

	"ai-site-audit/scraper"
)

type Section struct {
	Score           int      `json:"score"`
	Issues          []string `json:"issues"`
	Recommendations []string `json:"recommendations"`
}

type AuditResult struct {
	SiteURL     string  `json:"site_url"`
	Score       int     `json:"score"`
	Summary     string  `json:"summary"`
	SEO         Section `json:"seo"`
	UX          Section `json:"ux"`
	Performance Section `json:"performance"`
	Conversion  Section `json:"conversion"`
	QuickWins   []string `json:"quick_wins"`
}

var client *anthropic.Client

func init() {
	client = anthropic.NewClient(option.WithAPIKey(os.Getenv("ANTHROPIC_API_KEY")))
}

func Audit(ctx context.Context, scraped *scraper.Result) (*AuditResult, error) {
	prompt := fmt.Sprintf(`Analyze this website and return a JSON audit. No markdown, no code fences — raw JSON only.

URL: %s
Title: %s
Meta Description: %s
H1 Tags: %v
H2 Tags: %v
Has SSL: %v
Image Count: %d
Internal/External Links: %d total
Body Text (excerpt): %s

Return exactly this structure:
{
  "score": <overall 0-100>,
  "summary": "<2-3 sentence executive summary of the site's strengths and biggest gaps>",
  "seo": {
    "score": <0-100>,
    "issues": ["<specific issue>"],
    "recommendations": ["<actionable recommendation>"]
  },
  "ux": {
    "score": <0-100>,
    "issues": ["<specific issue>"],
    "recommendations": ["<actionable recommendation>"]
  },
  "performance": {
    "score": <0-100>,
    "issues": ["<specific issue>"],
    "recommendations": ["<actionable recommendation>"]
  },
  "conversion": {
    "score": <0-100>,
    "issues": ["<specific issue>"],
    "recommendations": ["<actionable recommendation>"]
  },
  "quick_wins": ["<high-impact, low-effort action>"]
}`,
		scraped.URL,
		scraped.Title,
		scraped.Description,
		scraped.H1s,
		scraped.H2s,
		scraped.HasSSL,
		scraped.ImageCount,
		len(scraped.Links),
		scraped.BodyText,
	)

	msg, err := client.Messages.New(ctx, anthropic.MessageNewParams{
		Model:     anthropic.F(anthropic.ModelClaude3_5SonnetLatest),
		MaxTokens: anthropic.F(int64(2048)),
		System: anthropic.F([]anthropic.TextBlockParam{
			{Text: anthropic.F("You are an expert web consultant specializing in SEO, UX, performance, and conversion rate optimization. Respond only with valid JSON — no markdown, no code blocks, no prose.")},
		}),
		Messages: anthropic.F([]anthropic.MessageParam{
			anthropic.UserMessageParam(anthropic.F([]anthropic.ContentBlockParamUnion{
				anthropic.TextBlockParam{Text: anthropic.F(prompt)},
			})),
		}),
	})
	if err != nil {
		return nil, fmt.Errorf("anthropic API: %w", err)
	}

	raw := msg.Content[0].Text
	raw = strings.TrimPrefix(raw, "```json")
	raw = strings.TrimPrefix(raw, "```")
	raw = strings.TrimSuffix(raw, "```")
	raw = strings.TrimSpace(raw)

	var result AuditResult
	if err := json.Unmarshal([]byte(raw), &result); err != nil {
		return nil, fmt.Errorf("parse response: %w\nraw: %s", err, raw)
	}
	result.SiteURL = scraped.URL
	return &result, nil
}

var pdfTemplate = template.Must(template.New("pdf").Parse(`<!DOCTYPE html>
<html lang="en">
<head>
<meta charset="UTF-8">
<style>
  * { box-sizing: border-box; margin: 0; padding: 0; }
  body { font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", sans-serif; color: #1a1a2e; padding: 40px; font-size: 14px; line-height: 1.6; }
  header { border-bottom: 3px solid #2563eb; padding-bottom: 16px; margin-bottom: 24px; }
  header h1 { font-size: 22px; color: #2563eb; }
  header p { color: #555; font-size: 13px; margin-top: 4px; }
  .overall { display: flex; align-items: center; gap: 24px; background: #f0f4ff; border-radius: 8px; padding: 20px; margin-bottom: 24px; }
  .score-circle { font-size: 48px; font-weight: 700; color: #2563eb; line-height: 1; }
  .summary { color: #333; }
  .section { border: 1px solid #e2e8f0; border-radius: 8px; padding: 16px; margin-bottom: 16px; }
  .section-header { display: flex; justify-content: space-between; align-items: center; margin-bottom: 12px; }
  .section-title { font-size: 16px; font-weight: 600; }
  .section-score { font-size: 20px; font-weight: 700; color: #2563eb; }
  .label { font-weight: 600; font-size: 12px; text-transform: uppercase; letter-spacing: 0.05em; color: #666; margin: 10px 0 6px; }
  ul { padding-left: 18px; }
  li { margin-bottom: 4px; }
  .quick-wins { background: #f0fdf4; border-left: 4px solid #16a34a; padding: 16px; border-radius: 4px; margin-top: 8px; }
  .quick-wins h2 { color: #16a34a; font-size: 16px; margin-bottom: 8px; }
</style>
</head>
<body>
  <header>
    <h1>Site Assessment Report</h1>
    <p>{{.SiteURL}}</p>
  </header>

  <div class="overall">
    <div class="score-circle">{{.Score}}</div>
    <div class="summary">
      <strong>Overall Score</strong>
      <p>{{.Summary}}</p>
    </div>
  </div>

  {{range .Sections}}
  <div class="section">
    <div class="section-header">
      <span class="section-title">{{.Name}}</span>
      <span class="section-score">{{.Score}}/100</span>
    </div>
    {{if .Issues}}
    <div class="label">Issues</div>
    <ul>{{range .Issues}}<li>{{.}}</li>{{end}}</ul>
    {{end}}
    {{if .Recommendations}}
    <div class="label">Recommendations</div>
    <ul>{{range .Recommendations}}<li>{{.}}</li>{{end}}</ul>
    {{end}}
  </div>
  {{end}}

  <div class="quick-wins">
    <h2>Quick Wins</h2>
    <ul>{{range .QuickWins}}<li>{{.}}</li>{{end}}</ul>
  </div>
</body>
</html>`))

type sectionData struct {
	Name            string
	Score           int
	Issues          []string
	Recommendations []string
}

func RenderPDF(ctx context.Context, result *AuditResult) ([]byte, error) {
	sections := []sectionData{
		{"SEO", result.SEO.Score, result.SEO.Issues, result.SEO.Recommendations},
		{"UX", result.UX.Score, result.UX.Issues, result.UX.Recommendations},
		{"Performance", result.Performance.Score, result.Performance.Issues, result.Performance.Recommendations},
		{"Conversion", result.Conversion.Score, result.Conversion.Issues, result.Conversion.Recommendations},
	}

	var buf bytes.Buffer
	if err := pdfTemplate.Execute(&buf, map[string]any{
		"SiteURL":   result.SiteURL,
		"Score":     result.Score,
		"Summary":   result.Summary,
		"Sections":  sections,
		"QuickWins": result.QuickWins,
	}); err != nil {
		return nil, fmt.Errorf("template: %w", err)
	}
	htmlContent := buf.String()

	allocCtx, cancel := chromedp.NewExecAllocator(context.Background(),
		append(chromedp.DefaultExecAllocatorOptions[:],
			chromedp.Flag("headless", true),
			chromedp.Flag("disable-gpu", true),
			chromedp.Flag("no-sandbox", true),
		)...,
	)
	defer cancel()

	chromCtx, cancel := chromedp.NewContext(allocCtx)
	defer cancel()

	var pdfBuf []byte
	if err := chromedp.Run(chromCtx,
		chromedp.Navigate("about:blank"),
		chromedp.ActionFunc(func(ctx context.Context) error {
			frameTree, err := page.GetFrameTree().Do(ctx)
			if err != nil {
				return err
			}
			return page.SetDocumentContent(frameTree.Frame.ID, htmlContent).Do(ctx)
		}),
		chromedp.ActionFunc(func(ctx context.Context) error {
			var err error
			pdfBuf, _, err = page.PrintToPDF().WithPrintBackground(true).Do(ctx)
			return err
		}),
	); err != nil {
		return nil, fmt.Errorf("chromedp PDF: %w", err)
	}

	return pdfBuf, nil
}
