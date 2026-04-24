# AI Site Assessment

A web app that takes any URL and returns a structured AI-powered site audit in seconds, with a one-click PDF export.

**Live demo:** [https://whits-ai-site-audit.up.railway.app/]

---

## What it does

Paste in a URL and get a scored audit across four categories:

- **SEO** — title, meta description, heading structure, SSL
- **UX** — content clarity, navigation, readability
- **Performance** — page weight signals, image usage, load indicators
- **Conversion** — CTA presence, messaging effectiveness, trust signals

Each category gets a 0–100 score, a list of specific issues, and actionable recommendations. A **Quick Wins** section surfaces the highest-impact, lowest-effort fixes. Results can be exported as a formatted PDF ready to hand to a sales team.

---

## How it's built

**Backend — Go**

- `scraper/` — uses [chromedp](https://github.com/chromedp/chromedp) (headless Chrome) to visit the target URL and extract page title, meta description, H1/H2 tags, body text, SSL status, image count, and links
- `auditor/` — feeds the scraped data into Claude (via the Anthropic Go SDK) with a structured prompt; parses the JSON response into typed Go structs; uses chromedp again to render the audit as a PDF via Chrome's native print engine
- `main.go` — lightweight HTTP server using Go 1.22's built-in `ServeMux` with method+path routing; no framework needed; audit results held in an in-memory store keyed by UUID

**Frontend — Vanilla JS**

- Single HTML page, no framework, no build step
- Fetch-based form submission with loading state and error handling
- Score bars and section cards rendered from the JSON response
- PDF download hits `GET /api/audit/{id}/pdf` which streams bytes directly from the server

**Deployment**

- Dockerized with a multi-stage build: Go compiler in the builder stage, minimal Debian runtime with Chromium installed
- Deployed on Railway; `ANTHROPIC_API_KEY` injected as an environment variable

---

## Running locally

```bash
cp .env.example .env
# add your ANTHROPIC_API_KEY to .env

make run
# → http://localhost:8080
```

Requires Go 1.26+ and Chrome/Chromium installed locally.

---

## Project structure

```
main.go          — HTTP server and route handlers
scraper/         — headless browser scraping logic
auditor/         — Claude API call, response parsing, PDF rendering
static/          — HTML, CSS, JS (served directly by Go)
Dockerfile       — multi-stage build for deployment
```
