package classification

import (
	"html"
	"strings"
)

func NormalizeTruncateStrategy(raw string) string {
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case TruncateStrategyHead:
		return TruncateStrategyHead
	case TruncateStrategyTail:
		return TruncateStrategyTail
	default:
		return TruncateStrategyTitleHeadTail
	}
}

func CleanTitleForClassification(raw string) string {
	if strings.TrimSpace(raw) == "" {
		return ""
	}
	cleaned := llmScriptRegex.ReplaceAllString(raw, " ")
	cleaned = llmStyleRegex.ReplaceAllString(cleaned, " ")
	cleaned = llmCommentRegex.ReplaceAllString(cleaned, " ")
	cleaned = llmTagRegex.ReplaceAllString(cleaned, " ")
	cleaned = llmImgMdRegex.ReplaceAllString(cleaned, " ")
	cleaned = llmLinkMdRegex.ReplaceAllString(cleaned, "$1")
	cleaned = html.UnescapeString(cleaned)
	return strings.TrimSpace(strings.Join(strings.Fields(cleaned), " "))
}

// PrepareClassificationParts extracts cleaned title and body text, applying maxChars
// and strategy across both LLM and ONNX classification engines.
// maxChars <= 0 disables truncation and returns full cleaned content.
func PrepareClassificationParts(item InputItem, maxChars int, strategy string) (string, string) {
	cleanTitle := CleanTitleForClassification(item.Title)

	rawBody := strings.TrimSpace(item.Content)
	if rawBody == "" {
		rawBody = strings.TrimSpace(item.Summary)
	}
	cleanBody := CleanContentForLLM(rawBody)
	if cleanBody == "" && rawBody != "" {
		cleanBody = CleanTitleForClassification(rawBody)
	}

	if cleanTitle == "" && cleanBody == "" {
		return "", ""
	}
	if maxChars <= 0 {
		return cleanTitle, cleanBody
	}

	titleRunes := []rune(cleanTitle)
	bodyRunes := []rune(cleanBody)
	if len(titleRunes)+len(bodyRunes) <= maxChars {
		return cleanTitle, cleanBody
	}

	switch NormalizeTruncateStrategy(strategy) {
	case TruncateStrategyHead:
		if len(titleRunes) >= maxChars {
			return strings.TrimSpace(string(titleRunes[:maxChars])), ""
		}
		budget := maxChars - len(titleRunes)
		return cleanTitle, takeHeadRunes(cleanBody, budget)

	case TruncateStrategyTail:
		if len(bodyRunes) >= maxChars {
			return "", takeTailRunes(cleanBody, maxChars)
		}
		titleBudget := maxChars - len(bodyRunes)
		return takeTailRunes(cleanTitle, titleBudget), cleanBody

	default: // TruncateStrategyTitleHeadTail
		if len(titleRunes) >= maxChars {
			return strings.TrimSpace(string(titleRunes[:maxChars])), ""
		}
		budget := maxChars - len(titleRunes)
		return cleanTitle, sampleHeadTailBody(cleanBody, budget)
	}
}

func PrepareClassificationText(item InputItem, maxChars int, strategy string) string {
	title, body := PrepareClassificationParts(item, maxChars, strategy)
	switch {
	case title != "" && body != "":
		return title + "\n\n" + body
	case title != "":
		return title
	default:
		return body
	}
}

func takeHeadRunes(text string, limit int) string {
	if limit <= 0 {
		return ""
	}
	runes := []rune(strings.TrimSpace(text))
	if len(runes) <= limit {
		return string(runes)
	}
	return strings.TrimSpace(string(runes[:limit]))
}

func takeTailRunes(text string, limit int) string {
	if limit <= 0 {
		return ""
	}
	runes := []rune(strings.TrimSpace(text))
	if len(runes) <= limit {
		return string(runes)
	}
	return strings.TrimSpace(string(runes[len(runes)-limit:]))
}

func sampleHeadTailBody(fullText string, limit int) string {
	if limit <= 0 {
		return ""
	}
	runes := []rune(strings.TrimSpace(fullText))
	if len(runes) <= limit {
		return string(runes)
	}
	const sep = "\n...\n"
	sepRunes := len([]rune(sep))
	if limit <= sepRunes+8 {
		return strings.TrimSpace(string(runes[:limit]))
	}

	avail := limit - sepRunes
	headLimit := (avail * 55) / 100
	if headLimit < 1 {
		headLimit = avail / 2
	}
	tailLimit := avail - headLimit
	if tailLimit < 1 {
		return strings.TrimSpace(string(runes[:limit]))
	}

	var paragraphs []string
	for _, line := range strings.Split(strings.TrimSpace(fullText), "\n") {
		trimmed := strings.TrimSpace(line)
		if trimmed != "" {
			paragraphs = append(paragraphs, trimmed)
		}
	}
	if len(paragraphs) <= 1 {
		headSlice := strings.TrimSpace(string(runes[:headLimit]))
		tailSlice := strings.TrimSpace(string(runes[len(runes)-tailLimit:]))
		return headSlice + sep + tailSlice
	}

	var headParts []string
	curHeadLen := 0
	headConsumed := 0
	for idx, p := range paragraphs {
		pRunes := []rune(p)
		extra := len(pRunes)
		if len(headParts) > 0 {
			extra++ // newline between paragraphs
		}
		if curHeadLen+extra > headLimit {
			if len(headParts) == 0 {
				headParts = append(headParts, strings.TrimSpace(string(pRunes[:headLimit])))
				curHeadLen = len([]rune(headParts[0]))
				headConsumed = idx + 1
			}
			break
		}
		headParts = append(headParts, p)
		curHeadLen += extra
		headConsumed = idx + 1
	}

	var tailParts []string
	curTailLen := 0
	for idx := len(paragraphs) - 1; idx >= headConsumed; idx-- {
		p := paragraphs[idx]
		pRunes := []rune(p)
		extra := len(pRunes)
		if len(tailParts) > 0 {
			extra++ // newline between paragraphs
		}
		if curTailLen+extra > tailLimit {
			if len(tailParts) == 0 {
				tailParts = append(tailParts, strings.TrimSpace(string(pRunes[len(pRunes)-tailLimit:])))
			}
			break
		}
		tailParts = append([]string{p}, tailParts...)
		curTailLen += extra
	}

	if len(headParts) == 0 || len(tailParts) == 0 {
		headSlice := strings.TrimSpace(string(runes[:headLimit]))
		tailSlice := strings.TrimSpace(string(runes[len(runes)-tailLimit:]))
		return headSlice + sep + tailSlice
	}
	return strings.Join(headParts, "\n") + sep + strings.Join(tailParts, "\n")
}
