package mailer

import (
	"fmt"
	"html"
	"strings"
)

type DigestArticle struct {
	Title       string
	Summary     string
	SourceName  string
	Author      string
	ExternalURL string
	PublishedAt string
}

func BuildEmailHTML(siteName, title, username, bodyHTML, buttonText, buttonURL, noteText string) string {
	if siteName == "" {
		siteName = "Utopia Daily"
	}
	escSite := html.EscapeString(siteName)
	escTitle := html.EscapeString(title)
	escUser := html.EscapeString(username)

	var greetingHTML string
	if escUser != "" {
		greetingHTML = fmt.Sprintf(`<p style="margin:0 0 14px;font-size:14px;color:#334155;">Hello <strong style="color:#0f172a;">%s</strong>,</p>`, escUser)
	}

	var buttonHTML string
	var fallbackHTML string
	if buttonText != "" && buttonURL != "" {
		escURL := html.EscapeString(buttonURL)
		buttonHTML = fmt.Sprintf(`
			<div style="text-align:center;margin:26px 0 20px;">
				<a href="%s" target="_blank" rel="noopener noreferrer" style="display:inline-block;background:#2563eb;color:#ffffff;font-weight:600;font-size:14px;text-decoration:none;padding:12px 28px;border-radius:8px;">%s</a>
			</div>`, escURL, html.EscapeString(buttonText))
		fallbackHTML = fmt.Sprintf(`
			<div style="margin-top:20px;padding-top:16px;border-top:1px solid #e2e8f0;font-size:12px;color:#64748b;word-break:break-all;">
				<a href="%s" style="color:#2563eb;text-decoration:underline;">%s</a>
			</div>`, escURL, escURL)
	}

	var noteHTML string
	if noteText != "" {
		noteHTML = fmt.Sprintf(`<div style="margin-top:16px;padding:10px 14px;background:#f8fafc;border-left:3px solid #94a3b8;font-size:12px;color:#475569;line-height:1.5;">%s</div>`, html.EscapeString(noteText))
	}

	return fmt.Sprintf(`<!DOCTYPE html>
<html>
<head><meta charset="UTF-8"><meta name="viewport" content="width=device-width,initial-scale=1.0"><title>%s</title></head>
<body style="margin:0;padding:0;background:#f1f5f9;font-family:-apple-system,BlinkMacSystemFont,'Segoe UI',Roboto,Arial,sans-serif;color:#1e293b;">
	<table role="presentation" width="100%%" cellspacing="0" cellpadding="0" style="padding:32px 16px;">
		<tr><td align="center">
			<table role="presentation" width="100%%" cellspacing="0" cellpadding="0" style="max-width:600px;background:#ffffff;border-radius:12px;border:1px solid #e2e8f0;overflow:hidden;">
				<tr>
					<td style="padding:20px 28px;background:#111827;color:#ffffff;font-size:17px;font-weight:700;">%s</td>
				</tr>
				<tr>
					<td style="padding:28px;">
						<h1 style="margin:0 0 16px;font-size:18px;font-weight:600;color:#0f172a;">%s</h1>
						%s
						<div style="font-size:14px;color:#334155;line-height:1.65;">%s</div>
						%s
						%s
						%s
					</td>
				</tr>
				<tr>
					<td style="padding:16px 28px;background:#f8fafc;border-top:1px solid #e2e8f0;text-align:center;font-size:12px;color:#94a3b8;">
						%s
					</td>
				</tr>
			</table>
		</td></tr>
	</table>
</body>
</html>`, escTitle, escSite, escTitle, greetingHTML, bodyHTML, buttonHTML, fallbackHTML, noteHTML, escSite)
}

func RenderActivationEmail(siteName, username, link string) (subject, htmlBody, textBody string) {
	if siteName == "" {
		siteName = "Utopia Daily"
	}
	subject = fmt.Sprintf("[%s] Account Activation", siteName)
	bodyHTML := `<p style="margin:0 0 12px;">Your account has been created. Click the button below within 24 hours to activate it.</p>`
	note := "If you did not create this account, ignore this email."
	htmlBody = BuildEmailHTML(siteName, "Account Activation", username, bodyHTML, "Activate Account", link, note)
	textBody = fmt.Sprintf("Hello %s,\r\n\r\nActivate your account within 24 hours:\r\n%s\r\n\r\n%s", username, link, siteName)
	return
}

func RenderPasswordResetEmail(siteName, username, link string) (subject, htmlBody, textBody string) {
	if siteName == "" {
		siteName = "Utopia Daily"
	}
	subject = fmt.Sprintf("[%s] Password Reset", siteName)
	bodyHTML := `<p style="margin:0 0 12px;">A password reset was requested for your account. Click the button below within 15 minutes to set a new password.</p>`
	note := "If you did not request a password reset, your password remains unchanged."
	htmlBody = BuildEmailHTML(siteName, "Password Reset", username, bodyHTML, "Reset Password", link, note)
	textBody = fmt.Sprintf("Hello %s,\r\n\r\nReset your password within 15 minutes:\r\n%s\r\n\r\n%s", username, link, siteName)
	return
}

func RenderTestEmail(siteName string) (subject, htmlBody, textBody string) {
	if siteName == "" {
		siteName = "Utopia Daily"
	}
	subject = fmt.Sprintf("[%s] SMTP Test", siteName)
	bodyHTML := `<p style="margin:0;">SMTP delivery is working properly.</p>`
	htmlBody = BuildEmailHTML(siteName, "SMTP Configuration Test", "", bodyHTML, "", "", "")
	textBody = fmt.Sprintf("%s SMTP delivery test succeeded.", siteName)
	return
}

func RenderNewsDigestEmail(siteName, digestTitle string, articles []DigestArticle) (subject, htmlBody, textBody string) {
	if siteName == "" {
		siteName = "Utopia Daily"
	}
	if digestTitle == "" {
		digestTitle = "Daily Digest"
	}
	subject = fmt.Sprintf("[%s] %s", siteName, digestTitle)

	var sb strings.Builder
	var tb strings.Builder
	tb.WriteString(digestTitle + "\r\n\r\n")

	for _, item := range articles {
		cleanTitle := stripMarkdownPlain(item.Title)
		cleanSummary := stripMarkdownPlain(item.Summary)
		escTitle := html.EscapeString(cleanTitle)
		escSummary := html.EscapeString(cleanSummary)
		escSource := html.EscapeString(item.SourceName)
		escDate := html.EscapeString(item.PublishedAt)

		titleBlock := escTitle
		if item.ExternalURL != "" {
			titleBlock = fmt.Sprintf(`<a href="%s" style="color:#1d4ed8;text-decoration:none;">%s</a>`, html.EscapeString(item.ExternalURL), escTitle)
		}

		sb.WriteString(`<div style="padding:14px 0;border-bottom:1px solid #e2e8f0;">`)
		sb.WriteString(fmt.Sprintf(`<div style="font-size:15px;font-weight:600;margin-bottom:4px;">%s</div>`, titleBlock))
		if escSource != "" || escDate != "" {
			sb.WriteString(fmt.Sprintf(`<div style="font-size:12px;color:#64748b;margin-bottom:6px;">%s`, escSource))
			if escSource != "" && escDate != "" {
				sb.WriteString(" · ")
			}
			sb.WriteString(fmt.Sprintf(`%s</div>`, escDate))
		}
		if escSummary != "" {
			sb.WriteString(fmt.Sprintf(`<div style="font-size:13px;color:#334155;line-height:1.5;">%s</div>`, escSummary))
		}
		sb.WriteString(`</div>`)

		tb.WriteString(fmt.Sprintf("- %s [%s]\r\n", cleanTitle, item.SourceName))
		if item.ExternalURL != "" {
			tb.WriteString(fmt.Sprintf("  %s\r\n", item.ExternalURL))
		}
		if cleanSummary != "" {
			tb.WriteString(fmt.Sprintf("  %s\r\n", cleanSummary))
		}
		tb.WriteString("\r\n")
	}

	htmlBody = BuildEmailHTML(siteName, digestTitle, "", sb.String(), "", "", "")
	textBody = tb.String()
	return
}

func stripMarkdownPlain(md string) string {
	s := strings.TrimSpace(md)
	if s == "" {
		return ""
	}
	for iter := 0; iter < 4; iter++ {
		next := stripMarkdownLinks(s)
		if next == s {
			break
		}
		s = next
	}
	s = strings.ReplaceAll(s, "```", "")
	s = strings.ReplaceAll(s, "**", "")
	s = strings.ReplaceAll(s, "~~", "")
	s = strings.ReplaceAll(s, "`", "")
	return strings.TrimSpace(s)
}

func stripMarkdownLinks(s string) string {
	var out strings.Builder
	out.Grow(len(s))
	i := 0
	for i < len(s) {
		isImg := false
		bracketStart := -1
		if i+1 < len(s) && s[i] == '!' && s[i+1] == '[' {
			isImg = true
			bracketStart = i + 1
		} else if s[i] == '[' {
			bracketStart = i
		}
		if bracketStart >= 0 {
			depth := 1
			j := bracketStart + 1
			closed := false
			for j < len(s) {
				if s[j] == '\\' && j+1 < len(s) {
					j += 2
					continue
				}
				if s[j] == '\n' {
					break
				}
				if s[j] == '[' {
					depth++
				} else if s[j] == ']' {
					depth--
					if depth == 0 {
						closed = true
						break
					}
				}
				j++
			}
			if closed && j+1 < len(s) && s[j+1] == '(' {
				pDepth := 1
				k := j + 2
				parenClosed := false
				for k < len(s) {
					if s[k] == '\\' && k+1 < len(s) {
						k += 2
						continue
					}
					if s[k] == '\n' {
						break
					}
					if s[k] == '(' {
						pDepth++
					} else if s[k] == ')' {
						pDepth--
						if pDepth == 0 {
							parenClosed = true
							break
						}
					}
					k++
				}
				if parenClosed {
					label := s[bracketStart+1 : j]
					if !isImg {
						out.WriteString(label)
					} else if alt := strings.TrimSpace(label); alt != "" {
						out.WriteString(alt)
					}
					i = k + 1
					continue
				}
			}
		}
		out.WriteByte(s[i])
		i++
	}
	return out.String()
}
