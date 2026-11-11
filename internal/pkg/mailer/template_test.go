package mailer

import (
	"strings"
	"testing"
)

func TestRenderNewsDigestEmailStripsMarkdownLinks(t *testing.T) {
	_, htmlBody, textBody := RenderNewsDigestEmail("Utopia Daily", "Daily Digest", []DigestArticle{
		{
			Title:       "**（[中国社会科学院大学](https://www.ucass.edu.cn/info/1073/5906.htm)）公告**",
			Summary:     "查看 [详情](https://example.com/a_(1).html) 与 [快讯] 说明",
			SourceName:  "TG Channel",
			ExternalURL: "https://t.me/test/1",
			PublishedAt: "2026-01-01 12:00",
		},
	})

	if strings.Contains(htmlBody, "ucass.edu.cn") || strings.Contains(htmlBody, "**") {
		t.Fatalf("html digest leaked markdown link syntax: %s", htmlBody)
	}
	if !strings.Contains(htmlBody, "（中国社会科学院大学）公告") {
		t.Fatalf("expected cleaned title in html digest: %s", htmlBody)
	}
	if !strings.Contains(textBody, "查看 详情 与 [快讯] 说明") {
		t.Fatalf("expected cleaned summary in text digest: %s", textBody)
	}
}
