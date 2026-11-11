package scrapers

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/UtopiaXC/UtopiaDaily/internal/module"
	"golang.org/x/net/html"
)

var (
	bgURLPattern = regexp.MustCompile(`background-image:\s*url\(['"]?(https?://[^'")]+)['"]?\)`)
	htmlBrRegex  = regexp.MustCompile(`(?i)<br\s*/?>`)
)

type TelegramChannelScraper struct {
	client *http.Client
}

func NewTelegramChannelScraper() *TelegramChannelScraper {
	return &TelegramChannelScraper{
		client: &http.Client{Timeout: 20 * time.Second},
	}
}

func (t *TelegramChannelScraper) Meta() module.Metadata {
	return module.Metadata{
		ID:      "telegram_channel",
		Kind:    module.KindScraper,
		Source:  module.SourceBuiltIn,
		Version: module.DefaultVersion,
		Author:  "UtopiaXC",
	}
}

func (t *TelegramChannelScraper) ConfigSchema() []module.ConfigFieldSpec {
	return []module.ConfigFieldSpec{
		{Key: "channels", Type: "array", DefaultItems: []string{"telegram"}, SortOrder: 1},
		{Key: "lookback_days", Type: "int", DefaultValue: "3", SortOrder: 2},
		{Key: "timezone", Type: "int", DefaultValue: "0", SortOrder: 3},
		{Key: "max_pages", Type: "int", DefaultValue: "5", SortOrder: 4},
	}
}

func (t *TelegramChannelScraper) Tasks() []module.TaskSpec {
	return []module.TaskSpec{
		{Key: "fetch_news", DefaultCron: "0 */1 * * *", DefaultEnabled: false},
	}
}

func (t *TelegramChannelScraper) Locales() map[string]map[string]string {
	return map[string]map[string]string{
		"zh_CN": {
			"module.telegram_channel.name":                      "Telegram 公开频道",
			"module.telegram_channel.description":               "抓取公开 Telegram 频道网页版消息",
			"module.telegram_channel.config.channels.desc":      "频道列表",
			"module.telegram_channel.config.channels.hint":      "输入公开频道用户名或链接",
			"module.telegram_channel.config.lookback_days.desc": "回溯天数",
			"module.telegram_channel.config.lookback_days.hint": "仅抓取该天数以内的历史消息",
			"module.telegram_channel.config.timezone.desc":      "时区偏移",
			"module.telegram_channel.config.timezone.hint":      "UTC 时区小时偏移量，例如 8 或 -5",
			"module.telegram_channel.config.max_pages.desc":     "最大页数",
			"module.telegram_channel.config.max_pages.hint":     "每个频道单次最多翻页次数",
			"module.telegram_channel.task.fetch_news.name":      "抓取频道消息",
			"module.telegram_channel.task.fetch_news.desc":      "获取已配置频道的最新公开消息",
		},
		"en_US": {
			"module.telegram_channel.name":                      "Telegram Public Channel",
			"module.telegram_channel.description":               "Scrapes messages from public Telegram channel web views",
			"module.telegram_channel.config.channels.desc":      "Channels",
			"module.telegram_channel.config.channels.hint":      "Enter public channel usernames or URLs",
			"module.telegram_channel.config.lookback_days.desc": "Lookback Days",
			"module.telegram_channel.config.lookback_days.hint": "Maximum age in days for collected messages",
			"module.telegram_channel.config.timezone.desc":      "Timezone Offset",
			"module.telegram_channel.config.timezone.hint":      "UTC hour offset, e.g. 8 or -5",
			"module.telegram_channel.config.max_pages.desc":     "Max Pages",
			"module.telegram_channel.config.max_pages.hint":     "Maximum web pages traversed per channel",
			"module.telegram_channel.task.fetch_news.name":      "Fetch Messages",
			"module.telegram_channel.task.fetch_news.desc":      "Collect latest public messages from configured channels",
		},
	}
}

func (t *TelegramChannelScraper) Scrape(ctx context.Context, taskKey string, cfg module.ConfigAccessor) ([]module.ScrapedItem, error) {
	channels := cfg.GetStringSlice("channels")
	if len(channels) == 0 {
		return nil, nil
	}
	lookback := cfg.GetInt("lookback_days", 3)
	if lookback < 1 {
		lookback = 3
	}
	maxPages := cfg.GetInt("max_pages", 5)
	if maxPages < 1 {
		maxPages = 5
	}
	tzOffset := cfg.GetInt("timezone", 0)
	cutoff := time.Now().UTC().AddDate(0, 0, -lookback)

	var all []module.ScrapedItem
	for _, rawCh := range channels {
		handle := extractTelegramHandle(rawCh)
		if handle == "" {
			continue
		}
		items, err := t.scrapeChannel(ctx, handle, cutoff, maxPages, tzOffset)
		if err != nil {
			continue
		}
		all = append(all, items...)
	}
	return all, nil
}

func (t *TelegramChannelScraper) TestConfig(ctx context.Context, cfg module.ConfigAccessor) module.TestResult {
	channels := cfg.GetStringSlice("channels")
	target := "telegram"
	if len(channels) > 0 {
		if h := extractTelegramHandle(channels[0]); h != "" {
			target = h
		}
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "https://t.me/s/"+target, nil)
	if err != nil {
		return module.TestResult{Success: false, MessageKey: "module.test.failed", Detail: err.Error()}
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (compatible; UtopiaDaily/"+module.DefaultVersion+")")
	resp, err := t.client.Do(req)
	if err != nil {
		return module.TestResult{Success: false, MessageKey: "module.test.failed", Detail: err.Error()}
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusOK {
		return module.TestResult{Success: false, MessageKey: "module.test.failed", Detail: fmt.Sprintf("HTTP %d", resp.StatusCode)}
	}
	return module.TestResult{Success: true, MessageKey: "module.test.passed", Detail: target}
}

func (t *TelegramChannelScraper) scrapeChannel(ctx context.Context, handle string, cutoff time.Time, maxPages, tzOffset int) ([]module.ScrapedItem, error) {
	var collected []module.ScrapedItem
	beforeID := 0
	channelTitle := handle
	loc := time.FixedZone("TG", tzOffset*3600)

	for page := 0; page < maxPages; page++ {
		pageURL := "https://t.me/s/" + handle
		if beforeID > 0 {
			pageURL = fmt.Sprintf("%s?before=%d", pageURL, beforeID)
		}

		req, err := http.NewRequestWithContext(ctx, http.MethodGet, pageURL, nil)
		if err != nil {
			return collected, err
		}
		req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 Chrome/120.0.0.0 Safari/537.36")

		resp, err := t.client.Do(req)
		if err != nil {
			return collected, err
		}
		doc, parseErr := html.Parse(resp.Body)
		_ = resp.Body.Close()
		if parseErr != nil {
			return collected, parseErr
		}

		if titleNode := findFirstByClass(doc, "tgme_channel_info_header_title"); titleNode != nil {
			if text := strings.TrimSpace(extractPlainText(titleNode)); text != "" {
				channelTitle = text
			}
		}

		msgNodes := findAllByClass(doc, "tgme_widget_message")
		if len(msgNodes) == 0 {
			break
		}

		lowestID := beforeID
		hasRecentOnPage := false
		for _, node := range msgNodes {
			postAttr := getAttr(node, "data-post")
			if postAttr == "" {
				continue
			}
			parts := strings.Split(postAttr, "/")
			postID, _ := strconv.Atoi(parts[len(parts)-1])
			if postID > 0 && (lowestID == 0 || postID < lowestID) {
				lowestID = postID
			}

			var pubTime time.Time
			if timeNode := findFirstTag(node, "time"); timeNode != nil {
				dt := getAttr(timeNode, "datetime")
				if parsed, err := time.Parse(time.RFC3339, dt); err == nil {
					pubTime = parsed.UTC()
				}
			}
			if pubTime.IsZero() {
				pubTime = time.Now().UTC()
			}
			if pubTime.Before(cutoff) {
				continue
			}
			hasRecentOnPage = true

			var mdText string
			if textNode := findFirstMessageText(node); textNode != nil {
				mdText = strings.TrimSpace(renderMarkdown(textNode))
			} else if previewNode := findFirstByClass(node, "tgme_widget_message_link_preview"); previewNode != nil {
				mdText = strings.TrimSpace(extractPlainText(previewNode))
			}

			var media []module.ScrapedMedia
			for _, photoNode := range findAllByClass(node, "tgme_widget_message_photo_wrap") {
				style := getAttr(photoNode, "style")
				if matches := bgURLPattern.FindStringSubmatch(style); len(matches) > 1 {
					media = append(media, module.ScrapedMedia{Type: "image", URL: matches[1]})
				}
			}

			if mdText == "" && len(media) == 0 {
				continue
			}

			var quotation *module.ScrapedQuotation
			if replyNode := findFirstByClass(node, "tgme_widget_message_reply"); replyNode != nil {
				qAuthor := ""
				if aNode := findFirstByClass(replyNode, "tgme_widget_message_author_name"); aNode != nil {
					qAuthor = strings.TrimSpace(extractPlainText(aNode))
				}
				qText := ""
				if tNode := findFirstByClass(replyNode, "tgme_widget_message_metatext"); tNode != nil {
					qText = strings.TrimSpace(extractPlainText(tNode))
				} else if tNode := findFirstByClass(replyNode, "tgme_widget_message_text"); tNode != nil {
					qText = strings.TrimSpace(extractPlainText(tNode))
				}
				qHref := getAttr(replyNode, "href")
				if qHref != "" && !strings.HasPrefix(qHref, "http") {
					qHref = "https://t.me" + qHref
				}
				quotation = &module.ScrapedQuotation{Author: qAuthor, Content: qText, URL: qHref}
			}

			postURL := fmt.Sprintf("https://t.me/%s", postAttr)
			sum := sha256.Sum256([]byte("telegram_channel|" + postURL))
			fp := hex.EncodeToString(sum[:])

			title := buildTelegramTitle(mdText, channelTitle, pubTime.In(loc))
			summary := buildSummary(mdText)

			collected = append(collected, module.ScrapedItem{
				Title:       title,
				Summary:     summary,
				Content:     mdText,
				ContentType: "markdown",
				SourceName:  channelTitle,
				Author:      channelTitle,
				ExternalURL: postURL,
				PublishedAt: pubTime,
				Fingerprint: fp,
				Media:       media,
				Quotation:   quotation,
			})
		}

		if !hasRecentOnPage || lowestID <= 1 || lowestID == beforeID {
			break
		}
		beforeID = lowestID
	}

	return collected, nil
}

func extractTelegramHandle(raw string) string {
	s := strings.TrimSpace(raw)
	if s == "" {
		return ""
	}
	s = strings.TrimPrefix(s, "@")
	s = strings.TrimPrefix(s, "https://t.me/s/")
	s = strings.TrimPrefix(s, "http://t.me/s/")
	s = strings.TrimPrefix(s, "https://t.me/")
	s = strings.TrimPrefix(s, "http://t.me/")
	parts := strings.Split(strings.Trim(s, "/"), "/")
	if len(parts) > 0 {
		return parts[0]
	}
	return ""
}

func buildTelegramTitle(content, channelName string, localTime time.Time) string {
	plain := stripMarkdownForPlainText(content)
	for _, line := range strings.Split(plain, "\n") {
		cleaned := strings.TrimSpace(line)
		cleaned = strings.TrimLeft(cleaned, "#*>- ")
		if cleaned != "" {
			runes := []rune(cleaned)
			if len(runes) > 90 {
				return string(runes[:90]) + "..."
			}
			return cleaned
		}
	}
	return fmt.Sprintf("%s %s", channelName, localTime.Format("2006-01-02 15:04"))
}

func buildSummary(content string) string {
	plain := stripMarkdownForPlainText(content)
	flat := strings.Join(strings.Fields(plain), " ")
	runes := []rune(flat)
	if len(runes) > 200 {
		return string(runes[:200]) + "..."
	}
	return flat
}

func stripMarkdownForPlainText(md string) string {
	if strings.TrimSpace(md) == "" {
		return ""
	}
	s := strings.ReplaceAll(md, "\r\n", "\n")
	s = strings.ReplaceAll(s, "\r", "\n")

	// Strip inline links and images repeatedly so nested formatting like [**label**](url) cleanly resolves.
	for iter := 0; iter < 4; iter++ {
		next := replaceMarkdownLinks(s)
		if next == s {
			break
		}
		s = next
	}

	lines := strings.Split(s, "\n")
	cleanedLines := make([]string, 0, len(lines))
	inFence := false
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "```") || strings.HasPrefix(trimmed, "~~~") {
			inFence = !inFence
			continue
		}
		if inFence {
			cleanedLines = append(cleanedLines, line)
			continue
		}
		cleanedLines = append(cleanedLines, stripInlineAndLinePrefixMarkdown(line))
	}
	return strings.Join(cleanedLines, "\n")
}

func stripInlineAndLinePrefixMarkdown(line string) string {
	s := strings.TrimLeft(line, " \t")
	for strings.HasPrefix(s, ">") {
		s = strings.TrimLeft(s[1:], " \t")
	}
	// Strip ATX heading markers (# Title) only when followed by whitespace.
	hashCount := 0
	for hashCount < len(s) && hashCount < 6 && s[hashCount] == '#' {
		hashCount++
	}
	if hashCount > 0 && hashCount < len(s) && (s[hashCount] == ' ' || s[hashCount] == '\t') {
		s = strings.TrimLeft(s[hashCount+1:], " \t")
	}
	// Strip unordered list prefixes (- item, * item, + item) when followed by space.
	if len(s) >= 2 && (s[0] == '-' || s[0] == '*' || s[0] == '+') && (s[1] == ' ' || s[1] == '\t') {
		s = strings.TrimLeft(s[2:], " \t")
	}

	// Strip paired symmetric inline markers.
	s = stripPairedMarker(s, "`", false)
	s = stripPairedMarker(s, "***", false)
	s = stripPairedMarker(s, "**", false)
	s = stripPairedMarker(s, "~~", false)
	s = stripPairedMarker(s, "||", false)
	s = stripPairedMarker(s, "___", true)
	s = stripPairedMarker(s, "__", true)
	s = stripPairedSingleAsterisk(s)
	return s
}

func replaceMarkdownLinks(s string) string {
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
			bracketClosed := false
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
						bracketClosed = true
						break
					}
				}
				j++
			}
			if bracketClosed && j+1 < len(s) && s[j+1] == '(' {
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
					} else if trimmedAlt := strings.TrimSpace(label); trimmedAlt != "" {
						out.WriteString(trimmedAlt)
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

func isASCIIWordByte(b byte) bool {
	return (b >= 'a' && b <= 'z') || (b >= 'A' && b <= 'Z') || (b >= '0' && b <= '9')
}

func stripPairedMarker(s, marker string, wordBoundary bool) string {
	mLen := len(marker)
	if mLen == 0 || !strings.Contains(s, marker) {
		return s
	}
	var out strings.Builder
	out.Grow(len(s))
	i := 0
	for i < len(s) {
		if i+mLen <= len(s) && s[i:i+mLen] == marker {
			if wordBoundary && i > 0 && isASCIIWordByte(s[i-1]) {
				out.WriteByte(s[i])
				i++
				continue
			}
			searchFrom := i + mLen
			relEnd := strings.Index(s[searchFrom:], marker)
			if relEnd > 0 {
				endIdx := searchFrom + relEnd
				afterEnd := endIdx + mLen
				if !strings.Contains(s[searchFrom:endIdx], "\n") && (!wordBoundary || afterEnd >= len(s) || !isASCIIWordByte(s[afterEnd])) {
					out.WriteString(s[searchFrom:endIdx])
					i = afterEnd
					continue
				}
			}
		}
		out.WriteByte(s[i])
		i++
	}
	return out.String()
}

func stripPairedSingleAsterisk(s string) string {
	if !strings.Contains(s, "*") {
		return s
	}
	var out strings.Builder
	out.Grow(len(s))
	i := 0
	for i < len(s) {
		if s[i] == '*' && (i+1 >= len(s) || s[i+1] != '*') && (i == 0 || s[i-1] != '*') {
			if i+1 < len(s) && s[i+1] != ' ' && s[i+1] != '\t' {
				relEnd := strings.IndexByte(s[i+1:], '*')
				if relEnd > 0 {
					endIdx := i + 1 + relEnd
					if s[endIdx-1] != ' ' && s[endIdx-1] != '\t' && !strings.Contains(s[i+1:endIdx], "\n") {
						out.WriteString(s[i+1 : endIdx])
						i = endIdx + 1
						continue
					}
				}
			}
		}
		out.WriteByte(s[i])
		i++
	}
	return out.String()
}

func getAttr(n *html.Node, key string) string {
	for _, a := range n.Attr {
		if a.Key == key {
			return a.Val
		}
	}
	return ""
}

func hasClass(n *html.Node, className string) bool {
	if n.Type != html.ElementNode {
		return false
	}
	for _, cls := range strings.Fields(getAttr(n, "class")) {
		if cls == className {
			return true
		}
	}
	return false
}

func findFirstByClass(root *html.Node, className string) *html.Node {
	if hasClass(root, className) {
		return root
	}
	for c := root.FirstChild; c != nil; c = c.NextSibling {
		if found := findFirstByClass(c, className); found != nil {
			return found
		}
	}
	return nil
}

func findFirstMessageText(root *html.Node) *html.Node {
	if hasClass(root, "tgme_widget_message_reply") {
		return nil
	}
	if hasClass(root, "tgme_widget_message_text") {
		return root
	}
	for c := root.FirstChild; c != nil; c = c.NextSibling {
		if found := findFirstMessageText(c); found != nil {
			return found
		}
	}
	return nil
}

func findAllByClass(root *html.Node, className string) []*html.Node {
	var res []*html.Node
	var walk func(*html.Node)
	walk = func(n *html.Node) {
		if hasClass(n, className) {
			res = append(res, n)
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			walk(c)
		}
	}
	walk(root)
	return res
}

func findFirstTag(root *html.Node, tag string) *html.Node {
	if root.Type == html.ElementNode && root.Data == tag {
		return root
	}
	for c := root.FirstChild; c != nil; c = c.NextSibling {
		if found := findFirstTag(c, tag); found != nil {
			return found
		}
	}
	return nil
}

func extractPlainText(root *html.Node) string {
	var b strings.Builder
	var walk func(*html.Node)
	walk = func(n *html.Node) {
		if n.Type == html.TextNode {
			b.WriteString(n.Data)
		} else if n.Type == html.ElementNode && n.Data == "br" {
			b.WriteByte('\n')
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			walk(c)
		}
	}
	walk(root)
	return b.String()
}

func renderMarkdown(root *html.Node) string {
	return strings.TrimSpace(renderMarkdownNode(root))
}

func extractChildrenMarkdown(n *html.Node) string {
	var b strings.Builder
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		b.WriteString(renderMarkdownNode(c))
	}
	return b.String()
}

func renderMarkdownNode(n *html.Node) string {
	var b strings.Builder
	var walk func(*html.Node)
	walk = func(cur *html.Node) {
		switch cur.Type {
		case html.TextNode:
			b.WriteString(cur.Data)
		case html.ElementNode:
			switch cur.Data {
			case "br":
				b.WriteByte('\n')
				return
			case "b", "strong":
				b.WriteString(wrapInlineMarkdown(extractChildrenMarkdown(cur), "**"))
				return
			case "i", "em":
				b.WriteString(wrapInlineMarkdown(extractChildrenMarkdown(cur), "*"))
				return
			case "s", "del", "strike":
				b.WriteString(wrapInlineMarkdown(extractChildrenMarkdown(cur), "~~"))
				return
			case "tg-spoiler":
				b.WriteString(wrapInlineMarkdown(extractChildrenMarkdown(cur), "||"))
				return
			case "span":
				if hasClass(cur, "tg-spoiler") {
					b.WriteString(wrapInlineMarkdown(extractChildrenMarkdown(cur), "||"))
					return
				}
			case "code":
				if inner := extractPlainText(cur); inner != "" {
					b.WriteString("`" + inner + "`")
				}
				return
			case "pre":
				if inner := extractPlainText(cur); inner != "" {
					b.WriteString("\n```\n" + inner + "\n```\n")
				}
				return
			case "a":
				href := strings.TrimSpace(getAttr(cur, "href"))
				label := extractPlainText(cur)
				cleanLabel := strings.TrimSpace(strings.ReplaceAll(strings.ReplaceAll(label, "\r", " "), "\n", " "))
				if href != "" && cleanLabel != "" {
					b.WriteString(fmt.Sprintf("[%s](%s)", cleanLabel, href))
				} else {
					b.WriteString(label)
				}
				return
			}
		}
		for c := cur.FirstChild; c != nil; c = c.NextSibling {
			walk(c)
		}
	}
	walk(n)
	return htmlBrRegex.ReplaceAllString(b.String(), "\n")
}

func wrapInlineMarkdown(raw, marker string) string {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return raw
	}
	leadLen := len(raw) - len(strings.TrimLeft(raw, " \t\n\r"))
	trimmedEnd := len(strings.TrimRight(raw, " \t\n\r"))
	return raw[:leadLen] + marker + raw[leadLen:trimmedEnd] + marker + raw[trimmedEnd:]
}
