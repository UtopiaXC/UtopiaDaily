package classification

import (
	"strings"
	"testing"
	"unicode/utf8"
)

func TestPrepareClassificationParts_UnlimitedAndFallback(t *testing.T) {
	longBody := strings.Repeat("段落一报道核心进展。", 300) // 3000 runes > 2000
	item := InputItem{
		Title:   "<b>重要</b> [突发新闻](https://example.com/a)",
		Summary: "摘要兜底文本",
		Content: "<p>" + longBody + "</p>",
	}

	title, body := PrepareClassificationParts(item, 0, TruncateStrategyTitleHeadTail)
	if title != "重要 突发新闻" {
		t.Fatalf("unexpected cleaned title: %q", title)
	}
	if utf8.RuneCountInString(body) != utf8.RuneCountInString(longBody) {
		t.Fatalf("expected full body when maxChars=0, got %d runes want %d", utf8.RuneCountInString(body), utf8.RuneCountInString(longBody))
	}

	title2, body2 := PrepareClassificationParts(item, -5, TruncateStrategyHead)
	if title2 != title || body2 != body {
		t.Fatalf("expected negative maxChars to disable truncation too")
	}

	fallbackItem := InputItem{
		Title:   "只有摘要",
		Summary: "这是摘要兜底内容用于分类分析",
		Content: "",
	}
	_, fbBody := PrepareClassificationParts(fallbackItem, 0, TruncateStrategyHead)
	if fbBody != fallbackItem.Summary {
		t.Fatalf("expected summary fallback when content is empty, got %q", fbBody)
	}
}

func TestPrepareClassificationParts_StrategiesRespectBudget(t *testing.T) {
	headMarker := "HEAD_MARKER_日本首相高市早苗外交表态"
	midMarker := strings.Repeat("中间补充背景细节说明内容", 40)
	tailMarker := "TAIL_MARKER_最终决定驻大阪总领事事务"

	multiParaItem := InputItem{
		Title:   "中日外交动态",
		Content: headMarker + "\n" + midMarker + "\n" + tailMarker,
	}

	t.Run("title_head_tail multi-paragraph", func(t *testing.T) {
		title, body := PrepareClassificationParts(multiParaItem, 120, TruncateStrategyTitleHeadTail)
		if title != "中日外交动态" {
			t.Fatalf("expected full title, got %q", title)
		}
		if !strings.Contains(body, "HEAD_MARKER") || !strings.Contains(body, "TAIL_MARKER") {
			t.Fatalf("expected both head and tail markers preserved, got %q", body)
		}
		totalRunes := utf8.RuneCountInString(title) + utf8.RuneCountInString(body)
		if totalRunes > 120 {
			t.Fatalf("expected total runes <= 120, got %d", totalRunes)
		}
	})

	t.Run("title_head_tail single mega-paragraph", func(t *testing.T) {
		singlePara := InputItem{
			Title:   "单段新闻",
			Content: headMarker + strings.Repeat("正文细节", 100) + tailMarker,
		}
		title, body := PrepareClassificationParts(singlePara, 110, TruncateStrategyTitleHeadTail)
		if !strings.Contains(body, "HEAD_MARKER") || !strings.Contains(body, "TAIL_MARKER") {
			t.Fatalf("expected single paragraph to keep head and tail markers, got %q", body)
		}
		if total := utf8.RuneCountInString(title) + utf8.RuneCountInString(body); total > 110 {
			t.Fatalf("expected total runes <= 110, got %d", total)
		}
	})

	t.Run("head keeps title and prefix only", func(t *testing.T) {
		title, body := PrepareClassificationParts(multiParaItem, 90, TruncateStrategyHead)
		if title != "中日外交动态" {
			t.Fatalf("expected title intact, got %q", title)
		}
		if !strings.Contains(body, "HEAD_MARKER") || strings.Contains(body, "TAIL_MARKER") {
			t.Fatalf("expected head marker without tail marker, got %q", body)
		}
		if total := utf8.RuneCountInString(title) + utf8.RuneCountInString(body); total > 90 {
			t.Fatalf("expected total runes <= 90, got %d", total)
		}
	})

	t.Run("tail prioritizes article end", func(t *testing.T) {
		title, body := PrepareClassificationParts(multiParaItem, 60, TruncateStrategyTail)
		if title != "" {
			t.Fatalf("expected title dropped when body alone fills tail budget, got %q", title)
		}
		if !strings.Contains(body, "TAIL_MARKER") || strings.Contains(body, "HEAD_MARKER") {
			t.Fatalf("expected tail marker without head marker, got %q", body)
		}
		if total := utf8.RuneCountInString(body); total > 60 {
			t.Fatalf("expected total runes <= 60, got %d", total)
		}

		// When full body fits inside budget with spare room, tail of title fills remainder.
		shortBodyItem := InputItem{Title: "前导标题尾部保留", Content: "短正文尾段"}
		shortTitle, shortBody := PrepareClassificationParts(shortBodyItem, 10, TruncateStrategyTail)
		if shortBody != "短正文尾段" || shortTitle != "题尾部保留" {
			t.Fatalf("unexpected tail partial-title result: title=%q body=%q", shortTitle, shortBody)
		}
		if txt := PrepareClassificationText(multiParaItem, 60, TruncateStrategyTail); strings.HasPrefix(txt, "\n") {
			t.Fatalf("PrepareClassificationText should not emit leading newlines when title omitted: %q", txt)
		}
	})

	t.Run("tiny maxChars smaller than title caps cleanly", func(t *testing.T) {
		longTitle := InputItem{
			Title:   "这是一个很长的标题用于测试上限截断逻辑",
			Content: "正文内容",
		}
		title, body := PrepareClassificationParts(longTitle, 6, TruncateStrategyHead)
		if utf8.RuneCountInString(title) > 6 || body != "" {
			t.Fatalf("expected truncated title <= 6 runes and empty body, got title=%q body=%q", title, body)
		}
	})
}
