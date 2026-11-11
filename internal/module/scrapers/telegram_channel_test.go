package scrapers

import (
	"strings"
	"testing"
	"time"

	"golang.org/x/net/html"
)

func TestStripMarkdownForPlainTextAndTelegramTitleSummary(t *testing.T) {
	md := "**[重要通知] 关于（[中国社会科学院大学](https://www.ucass.edu.cn/info/1073/5906.htm)）讲座安排**\n" +
		"> 详情请查看 [wiki (镜像)](https://en.wikipedia.org/wiki/University_(disambiguation)) 以及 ![封面](https://cdn.example.com/a_(1).png) `code` 和 ~~删除线~~ *斜体*。"

	title := buildTelegramTitle(md, "test_channel", time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC))
	wantTitle := "[重要通知] 关于（中国社会科学院大学）讲座安排"
	if title != wantTitle {
		t.Fatalf("buildTelegramTitle = %q, want %q", title, wantTitle)
	}

	summary := buildSummary(md)
	if strings.Contains(summary, "https://") || strings.Contains(summary, "**") || strings.Contains(summary, "~~") || strings.Contains(summary, "`") {
		t.Fatalf("summary still contains raw markdown/url: %q", summary)
	}
	if !strings.Contains(summary, "[重要通知]") {
		t.Fatalf("plain bracketed tag without url should be preserved, got %q", summary)
	}
	if !strings.Contains(summary, "（中国社会科学院大学）") {
		t.Fatalf("expected clean link label in Chinese parens, got %q", summary)
	}
	if !strings.Contains(summary, "wiki (镜像)") {
		t.Fatalf("expected balanced-paren link stripped cleanly, got %q", summary)
	}
	if !strings.Contains(summary, "封面") {
		t.Fatalf("expected image alt text preserved cleanly, got %q", summary)
	}
}

func TestRenderMarkdownPreservesLinksAndOuterWhitespace(t *testing.T) {
	rawHTML := `<div class="tgme_widget_message_text"><b>重要 </b>通知（<a href="https://www.ucass.edu.cn/info/1073/5906.htm">中国社会科学院大学</a>）</div>`
	doc, err := html.Parse(strings.NewReader(rawHTML))
	if err != nil {
		t.Fatalf("html.Parse failed: %v", err)
	}
	textNode := findFirstMessageText(doc)
	if textNode == nil {
		t.Fatal("message text node not found")
	}
	md := renderMarkdown(textNode)
	want := "**重要** 通知（[中国社会科学院大学](https://www.ucass.edu.cn/info/1073/5906.htm)）"
	if md != want {
		t.Fatalf("renderMarkdown = %q, want %q", md, want)
	}
}
