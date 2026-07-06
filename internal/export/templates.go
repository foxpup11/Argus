package export

import (
	"fmt"
	"html/template"
	"os"
	"strings"
	"time"

	"argus-desktop/internal/session"

	mdtemplate "text/template"
)

// generateHTML creates an HTML report for the session.
func generateHTML(sess *session.Session, diff string, outputPath string) error {
	tmpl, err := template.New("html").Funcs(template.FuncMap{
		"formatTime":  formatTime,
		"formatToken": formatToken,
		"riskClass":   riskClass,
		"changeClass": changeClass,
		"escapeHTML":   escapeHTML,
	}).Parse(htmlTemplate)
	if err != nil {
		return fmt.Errorf("failed to parse HTML template: %w", err)
	}

	data := map[string]interface{}{
		"Session":    sess,
		"Diff":       diff,
		"Generated":  time.Now().Format("2006-01-02 15:04:05"),
		"TotalTokens": sess.TokenUsage.InputTokens + sess.TokenUsage.OutputTokens,
	}

	file, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("failed to create HTML file: %w", err)
	}
	defer file.Close()

	if err := tmpl.Execute(file, data); err != nil {
		return fmt.Errorf("failed to execute HTML template: %w", err)
	}

	return nil
}

// generateMarkdown creates a Markdown report for the session.
func generateMarkdown(sess *session.Session, diff string, outputPath string) error {
	tmpl, err := mdtemplate.New("markdown").Funcs(mdtemplate.FuncMap{
		"formatTime":  formatTime,
		"formatToken": formatToken,
		"riskBadge":   riskBadge,
		"changeBadge": changeBadge,
		"escapeMarkdown": escapeMarkdown,
	}).Parse(markdownTemplate)
	if err != nil {
		return fmt.Errorf("failed to parse Markdown template: %w", err)
	}

	data := map[string]interface{}{
		"Session":    sess,
		"Diff":       diff,
		"Generated":  time.Now().Format("2006-01-02 15:04:05"),
		"TotalTokens": sess.TokenUsage.InputTokens + sess.TokenUsage.OutputTokens,
	}

	file, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("failed to create Markdown file: %w", err)
	}
	defer file.Close()

	if err := tmpl.Execute(file, data); err != nil {
		return fmt.Errorf("failed to execute Markdown template: %w", err)
	}

	return nil
}

// Template helper functions

func formatTime(t time.Time) string {
	if t.IsZero() {
		return "N/A"
	}
	return t.Format("2006-01-02 15:04:05")
}

func formatToken(n int) string {
	if n >= 1000000 {
		return fmt.Sprintf("%.1fM", float64(n)/1000000)
	}
	if n >= 1000 {
		return fmt.Sprintf("%.1fK", float64(n)/1000)
	}
	return fmt.Sprintf("%d", n)
}

func riskClass(risk session.RiskLevel) string {
	switch risk {
	case session.RiskSafe:
		return "risk-safe"
	case session.RiskReview:
		return "risk-review"
	case session.RiskDanger:
		return "risk-danger"
	default:
		return "risk-review"
	}
}

func changeClass(changeType session.ChangeType) string {
	switch changeType {
	case session.ChangeCreated:
		return "change-created"
	case session.ChangeModified:
		return "change-modified"
	case session.ChangeDeleted:
		return "change-deleted"
	default:
		return "change-modified"
	}
}

func riskBadge(risk session.RiskLevel) string {
	return strings.ToUpper(string(risk))
}

func changeBadge(changeType session.ChangeType) string {
	return strings.ToUpper(string(changeType))
}

func escapeHTML(s string) string {
	s = strings.ReplaceAll(s, "&", "&amp;")
	s = strings.ReplaceAll(s, "<", "&lt;")
	s = strings.ReplaceAll(s, ">", "&gt;")
	s = strings.ReplaceAll(s, `"`, "&quot;")
	return s
}

func escapeMarkdown(s string) string {
	// Escape markdown special characters
	specialChars := []string{"\\", "`", "*", "_", "{", "}", "[", "]", "(", ")", "#", "+", "-", ".", "!", "|"}
	for _, char := range specialChars {
		s = strings.ReplaceAll(s, char, "\\"+char)
	}
	return s
}

// HTML Template
const htmlTemplate = `<!DOCTYPE html>
<html lang="zh-CN">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Argus Report - {{.Session.ID}}</title>
    <style>
        :root {
            --bg-primary: #ffffff;
            --bg-secondary: #f8f9fa;
            --text-primary: #1a1a1a;
            --text-secondary: #6c757d;
            --accent: #0d6efd;
            --success: #198754;
            --warning: #fd7e14;
            --danger: #dc3545;
            --border: #e9ecef;
        }

        * {
            margin: 0;
            padding: 0;
            box-sizing: border-box;
        }

        body {
            font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif;
            background: var(--bg-secondary);
            color: var(--text-primary);
            line-height: 1.6;
            padding: 40px 20px;
        }

        .container {
            max-width: 1000px;
            margin: 0 auto;
            background: var(--bg-primary);
            border-radius: 12px;
            box-shadow: 0 4px 20px rgba(0, 0, 0, 0.08);
            overflow: hidden;
        }

        .header {
            background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
            color: white;
            padding: 40px;
            text-align: center;
        }

        .header h1 {
            font-size: 28px;
            font-weight: 600;
            margin-bottom: 8px;
        }

        .header .subtitle {
            font-size: 14px;
            opacity: 0.9;
        }

        .content {
            padding: 40px;
        }

        .section {
            margin-bottom: 32px;
        }

        .section-title {
            font-size: 18px;
            font-weight: 600;
            color: var(--text-primary);
            margin-bottom: 16px;
            padding-bottom: 8px;
            border-bottom: 2px solid var(--border);
        }

        .meta-grid {
            display: grid;
            grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
            gap: 16px;
            margin-bottom: 24px;
        }

        .meta-item {
            background: var(--bg-secondary);
            padding: 16px;
            border-radius: 8px;
        }

        .meta-label {
            font-size: 12px;
            color: var(--text-secondary);
            text-transform: uppercase;
            letter-spacing: 0.05em;
            margin-bottom: 4px;
        }

        .meta-value {
            font-size: 16px;
            font-weight: 600;
            color: var(--text-primary);
        }

        .file-table {
            width: 100%;
            border-collapse: collapse;
            margin-top: 16px;
        }

        .file-table th {
            background: var(--bg-secondary);
            padding: 12px 16px;
            text-align: left;
            font-size: 12px;
            font-weight: 600;
            color: var(--text-secondary);
            text-transform: uppercase;
        }

        .file-table td {
            padding: 12px 16px;
            border-bottom: 1px solid var(--border);
            font-size: 14px;
        }

        .risk-badge, .change-badge {
            display: inline-block;
            padding: 4px 10px;
            border-radius: 4px;
            font-size: 12px;
            font-weight: 600;
        }

        .risk-safe {
            background: rgba(25, 135, 84, 0.1);
            color: var(--success);
        }

        .risk-review {
            background: rgba(253, 126, 20, 0.1);
            color: var(--warning);
        }

        .risk-danger {
            background: rgba(220, 53, 69, 0.1);
            color: var(--danger);
        }

        .change-created {
            background: rgba(13, 110, 253, 0.1);
            color: var(--accent);
        }

        .change-modified {
            background: var(--bg-secondary);
            color: var(--text-secondary);
        }

        .change-deleted {
            background: rgba(220, 53, 69, 0.1);
            color: var(--danger);
        }

        .diff-section {
            background: #1e1e1e;
            border-radius: 8px;
            padding: 20px;
            overflow-x: auto;
        }

        .diff-content {
            font-family: 'SF Mono', 'Monaco', 'Consolas', monospace;
            font-size: 13px;
            line-height: 1.6;
            color: #d4d4d4;
            white-space: pre-wrap;
            word-wrap: break-word;
        }

        .diff-add {
            color: #4ec9b0;
            background: rgba(78, 201, 176, 0.1);
        }

        .diff-remove {
            color: #f14c4c;
            background: rgba(241, 76, 76, 0.1);
        }

        .footer {
            background: var(--bg-secondary);
            padding: 20px 40px;
            text-align: center;
            font-size: 12px;
            color: var(--text-secondary);
        }

        .token-bar {
            display: flex;
            gap: 24px;
            margin-top: 16px;
        }

        .token-item {
            display: flex;
            align-items: center;
            gap: 8px;
        }

        .token-dot {
            width: 8px;
            height: 8px;
            border-radius: 50%;
        }

        .token-dot.input {
            background: var(--accent);
        }

        .token-dot.output {
            background: var(--success);
        }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>Argus Report</h1>
            <div class="subtitle">Session Analysis Report</div>
        </div>

        <div class="content">
            <!-- Session Info -->
            <div class="section">
                <h2 class="section-title">Session Information</h2>
                <div class="meta-grid">
                    <div class="meta-item">
                        <div class="meta-label">Session ID</div>
                        <div class="meta-value">{{.Session.ID}}</div>
                    </div>
                    <div class="meta-item">
                        <div class="meta-label">Model</div>
                        <div class="meta-value">{{.Session.Model}}</div>
                    </div>
                    <div class="meta-item">
                        <div class="meta-label">Started At</div>
                        <div class="meta-value">{{formatTime .Session.StartedAt}}</div>
                    </div>
                    <div class="meta-item">
                        <div class="meta-label">Branch</div>
                        <div class="meta-value">{{.Session.GitBranch}}</div>
                    </div>
                </div>

                {{if .Session.Prompt}}
                <div class="meta-item">
                    <div class="meta-label">Prompt</div>
                    <div class="meta-value" style="font-size: 14px; font-weight: normal;">{{escapeHTML .Session.Prompt}}</div>
                </div>
                {{end}}

                <div class="token-bar">
                    <div class="token-item">
                        <span class="token-dot input"></span>
                        <span>Input Tokens: {{formatToken .Session.TokenUsage.InputTokens}}</span>
                    </div>
                    <div class="token-item">
                        <span class="token-dot output"></span>
                        <span>Output Tokens: {{formatToken .Session.TokenUsage.OutputTokens}}</span>
                    </div>
                    <div class="token-item">
                        <span>Total: {{formatToken .TotalTokens}}</span>
                    </div>
                </div>
            </div>

            <!-- File Changes -->
            <div class="section">
                <h2 class="section-title">File Changes ({{len .Session.FileChanges}} files)</h2>
                {{if .Session.FileChanges}}
                <table class="file-table">
                    <thead>
                        <tr>
                            <th style="width: 100px">Risk</th>
                            <th>File Path</th>
                            <th style="width: 100px">Change</th>
                            <th style="width: 80px">Actions</th>
                        </tr>
                    </thead>
                    <tbody>
                        {{range .Session.FileChanges}}
                        <tr>
                            <td><span class="risk-badge {{riskClass .Risk}}">{{.Risk}}</span></td>
                            <td>{{.Path}}</td>
                            <td><span class="change-badge {{changeClass .ChangeType}}">{{.ChangeType}}</span></td>
                            <td>{{.ActionCount}}</td>
                        </tr>
                        {{end}}
                    </tbody>
                </table>
                {{else}}
                <p style="color: var(--text-secondary);">No file changes recorded.</p>
                {{end}}
            </div>

            <!-- Diff -->
            {{if .Diff}}
            <div class="section">
                <h2 class="section-title">Diff</h2>
                <div class="diff-section">
                    <div class="diff-content">{{.Diff}}</div>
                </div>
            </div>
            {{end}}
        </div>

        <div class="footer">
            Generated by Argus | {{.Generated}}
        </div>
    </div>
</body>
</html>`

// Markdown Template
const markdownTemplate = `# Argus Report

**Session Analysis Report**

---

## Session Information

| Property | Value |
|----------|-------|
| **Session ID** | ` + "`" + `{{.Session.ID}}` + "`" + ` |
| **Model** | {{.Session.Model}} |
| **Started At** | {{formatTime .Session.StartedAt}} |
| **Branch** | {{.Session.GitBranch}} |

{{if .Session.Prompt}}
### Prompt

> {{.Session.Prompt}}
{{end}}

### Token Usage

- **Input Tokens:** {{formatToken .Session.TokenUsage.InputTokens}}
- **Output Tokens:** {{formatToken .Session.TokenUsage.OutputTokens}}
- **Total:** {{formatToken .TotalTokens}}

---

## File Changes ({{len .Session.FileChanges}} files)

{{if .Session.FileChanges}}
| Risk | File Path | Change | Actions |
|------|-----------|--------|---------|
{{range .Session.FileChanges}}| {{riskBadge .Risk}} | ` + "`" + `{{.Path}}` + "`" + ` | {{changeBadge .ChangeType}} | {{.ActionCount}} |
{{end}}
{{else}}
*No file changes recorded.*
{{end}}

---

## Diff

{{if .Diff}}
` + "```" + `
{{.Diff}}
` + "```" + `
{{else}}
*No diff data available.*
{{end}}

---

*Generated by Argus | {{.Generated}}*
`
