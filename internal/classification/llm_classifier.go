package classification

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"html"
	"io"
	"math"
	"net/http"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"
)

var (
	llmScriptRegex  = regexp.MustCompile(`(?is)<script.*?>.*?</script>`)
	llmStyleRegex   = regexp.MustCompile(`(?is)<style.*?>.*?</style>`)
	llmCommentRegex = regexp.MustCompile(`(?is)<!--.*?-->`)
	llmBlockRegex   = regexp.MustCompile(`(?i)<(p|div|br|h1|h2|h3|h4|li|article|section)[^>]*>`)
	llmTagRegex     = regexp.MustCompile(`<[^>]+>`)
	llmThinkRegex   = regexp.MustCompile(`(?is)<think>.*?</think>`)
	llmImgMdRegex   = regexp.MustCompile(`!\[.*?\]\(.*?\)`)
	llmLinkMdRegex  = regexp.MustCompile(`\[(.*?)\]\(.*?\)`)
)

type LLMClassifier struct {
	client *http.Client
}

func NewLLMClassifier() *LLMClassifier {
	return &LLMClassifier{
		client: &http.Client{Timeout: 120 * time.Second},
	}
}

func (l *LLMClassifier) Provider() string {
	return "llm"
}

func PresetLLMModels() []string {
	return []string{
		"qwen2.5:0.5b",
		"qwen2.5:3b",
		"qwen3:0.6b",
		"qwen3:4b",
		"qwen3.5:0.8b",
		"qwen3.5:2b",
		"qwen3.5:4b",
		"gemma3:270m",
		"smollm2:360m-instruct-q4_K_S",
	}
}

func CleanContentForLLM(raw string) string {
	if strings.TrimSpace(raw) == "" {
		return ""
	}
	cleaned := llmScriptRegex.ReplaceAllString(raw, " ")
	cleaned = llmStyleRegex.ReplaceAllString(cleaned, " ")
	cleaned = llmCommentRegex.ReplaceAllString(cleaned, " ")
	cleaned = llmBlockRegex.ReplaceAllString(cleaned, "\n")
	cleaned = llmTagRegex.ReplaceAllString(cleaned, " ")
	cleaned = llmImgMdRegex.ReplaceAllString(cleaned, " ")
	cleaned = llmLinkMdRegex.ReplaceAllString(cleaned, "$1")
	cleaned = html.UnescapeString(cleaned)

	lines := strings.Split(cleaned, "\n")
	valid := make([]string, 0, len(lines))
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if len([]rune(trimmed)) >= 8 {
			valid = append(valid, trimmed)
		}
	}
	if len(valid) == 0 {
		flat := strings.Join(strings.Fields(cleaned), " ")
		return strings.TrimSpace(flat)
	}
	return strings.Join(valid, "\n")
}

func (l *LLMClassifier) ClassifyItem(ctx context.Context, item InputItem, opts EngineOptions) ([]TagAssignment, error) {
	protocol := strings.ToLower(strings.TrimSpace(opts.LLMProtocol))
	if protocol == "" {
		protocol = "ollama"
	}
	baseURL := strings.TrimRight(strings.TrimSpace(opts.LLMURL), "/")
	if baseURL == "" {
		return nil, errors.New("LLM endpoint URL is required")
	}
	modelName := strings.TrimSpace(opts.LLMModel)
	if modelName == "" {
		modelName = strings.TrimSpace(opts.ModelName)
	}
	if modelName == "" {
		return nil, errors.New("LLM model name is required")
	}
	if len(opts.Profile.Groups) == 0 {
		return nil, errors.New("no classification groups or tags configured")
	}

	headerTitle, sampled := PrepareClassificationParts(item, opts.MaxInputChars, opts.TruncateStrategy)
	if headerTitle == "" && sampled == "" {
		return nil, errors.New("article content is empty")
	}

	systemPrompt, userPrompt := buildLLMPrompts(headerTitle, item.SourceName, sampled, opts.Profile, opts.LLMPrompt)
	schema := buildOllamaTaxonomySchema(opts.Profile)

	var contentStr string
	var err error
	switch protocol {
	case "ollama":
		contentStr, err = l.callOllamaChat(ctx, baseURL, modelName, systemPrompt, userPrompt, schema)
	case "openai":
		contentStr, err = l.callOpenAIChat(ctx, baseURL, opts.LLMAPIKey, modelName, systemPrompt, userPrompt)
	default:
		return nil, fmt.Errorf("unsupported LLM protocol: %s", protocol)
	}
	if err != nil {
		return nil, err
	}

	return parseLLMClassificationOutput(contentStr, opts.Profile, opts.Confidence)
}

func buildOllamaTaxonomySchema(profile ProfileSpec) map[string]any {
	groupProps := make(map[string]any)
	requiredGroups := make([]string, 0, len(profile.Groups))

	for _, g := range profile.Groups {
		tagNames := make([]string, 0, len(g.Tags))
		seen := make(map[string]bool, len(g.Tags))
		for _, t := range g.Tags {
			name := strings.TrimSpace(t.Name)
			if name == "" || seen[name] {
				continue
			}
			seen[name] = true
			tagNames = append(tagNames, name)
		}
		if len(tagNames) == 0 {
			continue
		}

		maxItems := len(tagNames)
		switch strings.ToLower(strings.TrimSpace(g.SelectionMode)) {
		case "single":
			maxItems = 1
		case "limit":
			if g.MaxResults > 0 && g.MaxResults < maxItems {
				maxItems = g.MaxResults
			} else if g.MaxResults <= 0 {
				maxItems = 1
			}
		}

		groupProps[g.Name] = map[string]any{
			"type": "object",
			"properties": map[string]any{
				"reason": map[string]any{
					"type": "string",
				},
				"tags": map[string]any{
					"type":     "array",
					"maxItems": maxItems,
					"items": map[string]any{
						"type": "object",
						"properties": map[string]any{
							"tag": map[string]any{
								"type": "string",
								"enum": tagNames,
							},
							"confidence": map[string]any{
								"type": "number",
							},
						},
						"required":             []string{"tag", "confidence"},
						"additionalProperties": false,
					},
				},
			},
			"required":             []string{"reason", "tags"},
			"additionalProperties": false,
		}
		requiredGroups = append(requiredGroups, g.Name)
	}

	if len(groupProps) == 0 {
		return nil
	}

	return map[string]any{
		"type": "object",
		"properties": map[string]any{
			"groups": map[string]any{
				"type":                 "object",
				"properties":           groupProps,
				"required":             requiredGroups,
				"additionalProperties": false,
			},
		},
		"required":             []string{"groups"},
		"additionalProperties": false,
	}
}

func buildLLMPrompts(title, sourceName, sampledBody string, profile ProfileSpec, customPrompt string) (string, string) {
	type exampleTag struct {
		Tag        string  `json:"tag"`
		Confidence float64 `json:"confidence"`
	}
	type exampleGroup struct {
		Reason string       `json:"reason"`
		Tags   []exampleTag `json:"tags"`
	}
	encodeNoEscape := func(v any) string {
		var buf bytes.Buffer
		enc := json.NewEncoder(&buf)
		enc.SetEscapeHTML(false)
		_ = enc.Encode(v)
		return strings.TrimSpace(buf.String())
	}

	var exampleSB strings.Builder
	exampleSB.WriteString(`{"groups":{`)
	addedExample := 0
	for _, g := range profile.Groups {
		if len(g.Tags) == 0 {
			continue
		}
		if addedExample > 0 {
			exampleSB.WriteByte(',')
		}
		exampleSB.WriteString(encodeNoEscape(g.Name))
		exampleSB.WriteByte(':')
		exampleSB.WriteString(encodeNoEscape(exampleGroup{
			Reason: "<brief 1-phrase factual analysis naming core topic and all involved countries/locations>",
			Tags: []exampleTag{
				{
					Tag:        fmt.Sprintf("<exact tag name from group %s>", g.Name),
					Confidence: 0.9,
				},
			},
		}))
		addedExample++
	}
	if addedExample == 0 {
		exampleSB.WriteString(`"<exact group name>":{"reason":"<brief 1-phrase factual analysis>","tags":[{"tag":"<exact tag name>","confidence":0.9}]}`)
	}
	exampleSB.WriteString("}}")

	var sb strings.Builder
	sb.WriteString("You are an automated multilingual news classification engine.\n")
	sb.WriteString("Classify the article independently across EVERY configured group below based on overall meaning, subject matter, and ALL involved countries/regions.\n\n")
	sb.WriteString("Return ONLY a valid JSON object matching this exact structure:\n")
	sb.WriteString(exampleSB.String() + "\n\n")
	sb.WriteString("Rules:\n")
	sb.WriteString("- Evaluate EVERY group independently; for each group, first state the brief factual basis in `reason` (e.g., editorial format, core topics, or exact countries/locations involved under 20 words), then select matching `tag`(s) in `tags`.\n")
	sb.WriteString("- Every `tag` MUST be an exact Tag name from that group's list (never output group names, raw keywords, or unlisted country/city names).\n")
	sb.WriteString("- If a tag covers a broader category or region whose name or `optional_keywords` refer to specific countries, cities, or subtopics, map those countries/subtopics to that parent Tag name.\n")
	sb.WriteString("- Any `optional_keywords` listed under a tag are ONLY auxiliary reference examples, NOT required match conditions; assign the best semantic tag whenever the article fits even if zero listed keywords appear verbatim.\n")
	sb.WriteString("- Include all distinct clearly relevant tags up to each group's selection limit (for example, when multiple major regions or domains are directly involved, include up to the limit). Only emit high-confidence matches rather than 0.0 filler entries; use [] only when zero tags apply.\n")
	sb.WriteString("- Prefer higher-priority and more specific tags over broader umbrella tags in the same group when selection slots are limited (for example, do not select a broad region or umbrella tag when higher-priority specific country or subtopic tags already match).\n")
	sb.WriteString("- Honor each group's selection rule strictly (single = 1 best tag, limit = up to max_results tags, multi = all clearly applicable tags).\n")
	sb.WriteString("- Set confidence between 0.0 and 1.0 (0.85-0.98 for strong clear matches, 0.65-0.84 for moderate semantic fits, below 0.50 only for weak candidates).\n")
	if cp := strings.TrimSpace(customPrompt); cp != "" {
		sb.WriteString("\nAdditional User Instructions:\n")
		sb.WriteString(cp + "\n")
	}
	sb.WriteString("\nConfigured Taxonomy:\n")
	for _, g := range profile.Groups {
		modeDesc := "multi: pick all matching tags"
		switch strings.ToLower(strings.TrimSpace(g.SelectionMode)) {
		case "single":
			modeDesc = "single: pick 1 best tag"
		case "limit":
			maxR := g.MaxResults
			if maxR <= 0 {
				maxR = 1
			}
			modeDesc = fmt.Sprintf("limit: pick 1 to %d best tags", maxR)
		case "multi":
			modeDesc = "multi: pick all matching tags"
		}
		sb.WriteString(fmt.Sprintf("- Group %q [%s]", g.Name, modeDesc))
		if strings.TrimSpace(g.Description) != "" {
			sb.WriteString(": " + strings.TrimSpace(g.Description))
		}
		sb.WriteString("\n  Tags:\n")
		for _, t := range g.Tags {
			sb.WriteString(fmt.Sprintf("  * %q", t.Name))
			if t.Priority != 0 {
				sb.WriteString(fmt.Sprintf(" [priority: %d]", t.Priority))
			}
			if strings.TrimSpace(t.Description) != "" {
				sb.WriteString(" — " + strings.TrimSpace(t.Description))
			}
			if len(t.Keywords) > 0 {
				sb.WriteString(" [optional_keywords: " + strings.Join(t.Keywords, ", ") + "]")
			}
			sb.WriteString("\n")
		}
	}

	var userSB strings.Builder
	if strings.TrimSpace(sourceName) != "" {
		userSB.WriteString("Source: " + strings.TrimSpace(sourceName) + "\n")
	}
	if strings.TrimSpace(title) != "" {
		userSB.WriteString("Title: " + strings.TrimSpace(title) + "\n")
	}
	if strings.TrimSpace(sampledBody) != "" {
		userSB.WriteString("\nArticle Content:\n")
		userSB.WriteString(sampledBody)
	}
	return sb.String(), userSB.String()
}

func estimateOllamaNumCtx(systemPrompt, userPrompt string) int {
	totalRunes := utf8.RuneCountInString(systemPrompt) + utf8.RuneCountInString(userPrompt)
	switch {
	case totalRunes > 9000:
		return 16384
	case totalRunes > 4500:
		return 8192
	default:
		return 4096
	}
}

func (l *LLMClassifier) callOllamaChat(ctx context.Context, baseURL, modelName, systemPrompt, userPrompt string, schema map[string]any) (string, error) {
	endpoint := strings.TrimRight(baseURL, "/")
	if !strings.HasSuffix(endpoint, "/api/chat") {
		endpoint += "/api/chat"
	}

	type ollamaAttempt struct {
		useSchema    bool
		includeThink bool
	}

	numCtx := estimateOllamaNumCtx(systemPrompt, userPrompt)
	buildAttempt := func(att ollamaAttempt) ([]byte, error) {
		var formatVal any = "json"
		if att.useSchema && schema != nil {
			formatVal = schema
		}
		payload := map[string]any{
			"model": modelName,
			"messages": []map[string]string{
				{"role": "system", "content": systemPrompt},
				{"role": "user", "content": userPrompt},
			},
			"format": formatVal,
			"stream": false,
			"options": map[string]any{
				"temperature": 0.0,
				"num_ctx":     numCtx,
				"num_predict": 1024,
			},
		}
		if att.includeThink {
			payload["think"] = false
		}
		return json.Marshal(payload)
	}

	attempts := []ollamaAttempt{
		{useSchema: false, includeThink: true},
		{useSchema: false, includeThink: false},
	}
	if schema != nil {
		attempts = []ollamaAttempt{
			{useSchema: true, includeThink: true},
			{useSchema: true, includeThink: false},
			{useSchema: false, includeThink: true},
			{useSchema: false, includeThink: false},
		}
	}

	var respBytes []byte
	var status int
	var err error
	for i, att := range attempts {
		respBytes, status, err = l.postJSONPayload(ctx, endpoint, "", func() ([]byte, error) {
			return buildAttempt(att)
		})
		if err == nil {
			break
		}
		if status != http.StatusBadRequest || i == len(attempts)-1 {
			break
		}
		lowerBody := strings.ToLower(string(respBytes))
		if strings.Contains(lowerBody, "not found") || strings.Contains(lowerBody, "does not exist") {
			break
		}
	}
	if err != nil {
		return "", err
	}

	var parsed struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
		Error string `json:"error"`
	}
	if err := json.Unmarshal(respBytes, &parsed); err != nil {
		return "", fmt.Errorf("invalid ollama response: %w", err)
	}
	if parsed.Error != "" {
		return "", fmt.Errorf("ollama error: %s", parsed.Error)
	}
	return parsed.Message.Content, nil
}

func (l *LLMClassifier) callOpenAIChat(ctx context.Context, baseURL, apiKey, modelName, systemPrompt, userPrompt string) (string, error) {
	endpoint := strings.TrimRight(baseURL, "/")
	switch {
	case strings.HasSuffix(endpoint, "/chat/completions"):
	case strings.HasSuffix(endpoint, "/v1"):
		endpoint += "/chat/completions"
	default:
		endpoint += "/v1/chat/completions"
	}

	buildBody := func(useJSONMode bool) ([]byte, error) {
		payload := map[string]any{
			"model": modelName,
			"messages": []map[string]string{
				{"role": "system", "content": systemPrompt},
				{"role": "user", "content": userPrompt},
			},
			"temperature": 0.0,
		}
		if useJSONMode {
			payload["response_format"] = map[string]string{"type": "json_object"}
		}
		return json.Marshal(payload)
	}

	respBytes, status, err := l.postJSON(ctx, endpoint, apiKey, buildBody, true)
	if err != nil && status == http.StatusBadRequest {
		respBytes, _, err = l.postJSON(ctx, endpoint, apiKey, buildBody, false)
	}
	if err != nil {
		return "", err
	}

	var parsed struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
		Error *struct {
			Message string `json:"message"`
		} `json:"error"`
	}
	if err := json.Unmarshal(respBytes, &parsed); err != nil {
		return "", fmt.Errorf("invalid openai response: %w", err)
	}
	if parsed.Error != nil && parsed.Error.Message != "" {
		return "", fmt.Errorf("openai error: %s", parsed.Error.Message)
	}
	if len(parsed.Choices) == 0 {
		return "", errors.New("empty choices from LLM")
	}
	return parsed.Choices[0].Message.Content, nil
}

func (l *LLMClassifier) postJSON(ctx context.Context, endpoint, apiKey string, build func(bool) ([]byte, error), flag bool) ([]byte, int, error) {
	return l.postJSONPayload(ctx, endpoint, apiKey, func() ([]byte, error) {
		return build(flag)
	})
}

func (l *LLMClassifier) postJSONPayload(ctx context.Context, endpoint, apiKey string, build func() ([]byte, error)) ([]byte, int, error) {
	body, err := build()
	if err != nil {
		return nil, 0, err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(body))
	if err != nil {
		return nil, 0, err
	}
	req.Header.Set("Content-Type", "application/json")
	if strings.TrimSpace(apiKey) != "" {
		req.Header.Set("Authorization", "Bearer "+strings.TrimSpace(apiKey))
	}
	resp, err := l.client.Do(req)
	if err != nil {
		return nil, 0, err
	}
	defer func() { _ = resp.Body.Close() }()
	raw, _ := io.ReadAll(io.LimitReader(resp.Body, 4<<20))
	if resp.StatusCode >= 300 {
		return raw, resp.StatusCode, fmt.Errorf("HTTP %d: %s", resp.StatusCode, strings.TrimSpace(string(raw)))
	}
	return raw, resp.StatusCode, nil
}

type flexFloat struct {
	Value float64
	Set   bool
}

func (f *flexFloat) UnmarshalJSON(data []byte) error {
	s := strings.TrimSpace(string(data))
	if s == "" || s == "null" {
		f.Set = false
		return nil
	}
	if len(s) >= 2 && s[0] == '"' && s[len(s)-1] == '"' {
		var unquoted string
		if err := json.Unmarshal(data, &unquoted); err != nil {
			return err
		}
		s = strings.TrimSpace(unquoted)
	}
	s = strings.TrimSpace(strings.TrimSuffix(s, "%"))
	if s == "" {
		f.Set = false
		return nil
	}
	v, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return err
	}
	f.Value = v
	f.Set = true
	return nil
}

type llmRawTag struct {
	Group      string    `json:"group"`
	GroupName  string    `json:"group_name"`
	Tag        string    `json:"tag"`
	TagName    string    `json:"tag_name"`
	Confidence flexFloat `json:"confidence"`
	Reason     string    `json:"reason"`
}

func extractJSONPayload(rawContent string) string {
	cleaned := llmThinkRegex.ReplaceAllString(rawContent, "")
	cleaned = strings.TrimSpace(cleaned)
	cleaned = strings.TrimPrefix(cleaned, "```json")
	cleaned = strings.TrimPrefix(cleaned, "```")
	cleaned = strings.TrimSuffix(cleaned, "```")
	cleaned = strings.TrimSpace(cleaned)

	arrStart := strings.IndexByte(cleaned, '[')
	objStart := strings.IndexByte(cleaned, '{')
	useArray := false
	if arrStart >= 0 && (objStart < 0 || arrStart < objStart) {
		after := strings.TrimLeft(cleaned[arrStart+1:], " \t\r\n")
		if len(after) > 0 && (after[0] == '{' || after[0] == '"' || after[0] == ']' || after[0] == '[') {
			useArray = true
		}
	}
	if useArray {
		arrEnd := strings.LastIndexByte(cleaned, ']')
		if arrEnd > arrStart {
			return cleaned[arrStart : arrEnd+1]
		}
	}
	if objStart >= 0 {
		objEnd := strings.LastIndexByte(cleaned, '}')
		if objEnd > objStart {
			return cleaned[objStart : objEnd+1]
		}
	}
	return cleaned
}

func parseFlexibleTagArray(raw json.RawMessage, defaultGroup, defaultReason string) ([]llmRawTag, error) {
	trimmed := bytes.TrimSpace(raw)
	if len(trimmed) == 0 || bytes.Equal(trimmed, []byte("null")) {
		return nil, nil
	}

	var out []llmRawTag
	switch trimmed[0] {
	case '[':
		var elems []json.RawMessage
		if err := json.Unmarshal(trimmed, &elems); err != nil {
			return nil, err
		}
		for _, el := range elems {
			elTrim := bytes.TrimSpace(el)
			if len(elTrim) == 0 || bytes.Equal(elTrim, []byte("null")) {
				continue
			}
			switch elTrim[0] {
			case '"':
				var s string
				if err := json.Unmarshal(elTrim, &s); err != nil {
					return nil, err
				}
				if strings.TrimSpace(s) != "" {
					out = append(out, llmRawTag{
						Group:  defaultGroup,
						Tag:    s,
						Reason: defaultReason,
					})
				}
			case '{':
				var item llmRawTag
				if err := json.Unmarshal(elTrim, &item); err != nil {
					return nil, err
				}
				if strings.TrimSpace(item.Group) == "" && strings.TrimSpace(item.GroupName) == "" {
					item.Group = defaultGroup
				}
				if strings.TrimSpace(item.Reason) == "" {
					item.Reason = defaultReason
				}
				out = append(out, item)
			}
		}
	case '{':
		var item llmRawTag
		if err := json.Unmarshal(trimmed, &item); err != nil {
			return nil, err
		}
		if strings.TrimSpace(item.Group) == "" && strings.TrimSpace(item.GroupName) == "" {
			item.Group = defaultGroup
		}
		if strings.TrimSpace(item.Reason) == "" {
			item.Reason = defaultReason
		}
		out = append(out, item)
	case '"':
		var s string
		if err := json.Unmarshal(trimmed, &s); err != nil {
			return nil, err
		}
		if strings.TrimSpace(s) != "" {
			out = append(out, llmRawTag{
				Group:  defaultGroup,
				Tag:    s,
				Reason: defaultReason,
			})
		}
	}
	return out, nil
}

func appendGroupRawTags(out *[]llmRawTag, groupName string, raw json.RawMessage) error {
	trimmed := bytes.TrimSpace(raw)
	if len(trimmed) == 0 || bytes.Equal(trimmed, []byte("null")) {
		return nil
	}

	switch trimmed[0] {
	case '{':
		var wrapper struct {
			Group      string          `json:"group"`
			GroupName  string          `json:"group_name"`
			Tag        string          `json:"tag"`
			TagName    string          `json:"tag_name"`
			Confidence flexFloat       `json:"confidence"`
			Reason     string          `json:"reason"`
			Tags       json.RawMessage `json:"tags"`
		}
		if err := json.Unmarshal(trimmed, &wrapper); err != nil {
			return err
		}
		effGroup := strings.TrimSpace(wrapper.Group)
		if effGroup == "" {
			effGroup = strings.TrimSpace(wrapper.GroupName)
		}
		if effGroup == "" {
			effGroup = groupName
		}
		if len(bytes.TrimSpace(wrapper.Tags)) > 0 {
			items, err := parseFlexibleTagArray(wrapper.Tags, effGroup, strings.TrimSpace(wrapper.Reason))
			if err != nil {
				return err
			}
			*out = append(*out, items...)
			return nil
		}
		if strings.TrimSpace(wrapper.Tag) != "" || strings.TrimSpace(wrapper.TagName) != "" {
			*out = append(*out, llmRawTag{
				Group:      effGroup,
				Tag:        wrapper.Tag,
				TagName:    wrapper.TagName,
				Confidence: wrapper.Confidence,
				Reason:     wrapper.Reason,
			})
		}
		return nil
	case '[', '"':
		items, err := parseFlexibleTagArray(trimmed, groupName, "")
		if err != nil {
			return err
		}
		*out = append(*out, items...)
		return nil
	default:
		return nil
	}
}

func appendGroupMapRawTags(out *[]llmRawTag, gmap map[string]json.RawMessage, profile ProfileSpec) error {
	handled := make(map[string]bool, len(gmap))
	for _, g := range profile.Groups {
		for k, v := range gmap {
			if strings.EqualFold(strings.TrimSpace(k), strings.TrimSpace(g.Name)) {
				if err := appendGroupRawTags(out, g.Name, v); err != nil {
					return err
				}
				handled[k] = true
			}
		}
	}
	var extraKeys []string
	for k := range gmap {
		if !handled[k] {
			extraKeys = append(extraKeys, k)
		}
	}
	sort.Strings(extraKeys)
	for _, k := range extraKeys {
		if err := appendGroupRawTags(out, k, gmap[k]); err != nil {
			return err
		}
	}
	return nil
}

func parseLLMClassificationOutput(rawContent string, profile ProfileSpec, confidencePct int) ([]TagAssignment, error) {
	cleaned := extractJSONPayload(rawContent)

	var rawTags []llmRawTag
	if strings.HasPrefix(cleaned, "[") {
		items, err := parseFlexibleTagArray([]byte(cleaned), "", "")
		if err != nil {
			return nil, fmt.Errorf("failed parsing LLM JSON array: %w", err)
		}
		rawTags = items
	} else {
		var topMap map[string]json.RawMessage
		if err := json.Unmarshal([]byte(cleaned), &topMap); err != nil {
			return nil, fmt.Errorf("failed parsing LLM JSON: %w", err)
		}
		if groupsRaw, ok := topMap["groups"]; ok && len(bytes.TrimSpace(groupsRaw)) > 0 && bytes.TrimSpace(groupsRaw)[0] == '{' {
			var gmap map[string]json.RawMessage
			if err := json.Unmarshal(groupsRaw, &gmap); err != nil {
				return nil, fmt.Errorf("failed parsing LLM groups JSON: %w", err)
			}
			if err := appendGroupMapRawTags(&rawTags, gmap, profile); err != nil {
				return nil, fmt.Errorf("failed parsing LLM group entries: %w", err)
			}
		} else if tagsRaw, ok := topMap["tags"]; ok && len(bytes.TrimSpace(tagsRaw)) > 0 && bytes.TrimSpace(tagsRaw)[0] == '[' {
			items, err := parseFlexibleTagArray(tagsRaw, "", "")
			if err != nil {
				return nil, fmt.Errorf("failed parsing LLM tags JSON: %w", err)
			}
			rawTags = items
		} else {
			if err := appendGroupMapRawTags(&rawTags, topMap, profile); err != nil {
				return nil, fmt.Errorf("failed parsing LLM group map JSON: %w", err)
			}
		}
	}

	groupLimit := make(map[string]int, len(profile.Groups))
	tagByGroupAndName := make(map[string]TagSpec)
	tagByNameOnly := make(map[string][]struct {
		GroupName string
		Tag       TagSpec
	})

	for _, g := range profile.Groups {
		limit := g.MaxResults
		switch strings.ToLower(strings.TrimSpace(g.SelectionMode)) {
		case "single":
			limit = 1
		case "multi":
			limit = 1000
		default:
			if limit <= 0 {
				limit = 1
			}
		}
		groupLimit[g.Name] = limit
		for _, t := range g.Tags {
			key := strings.ToLower(strings.TrimSpace(g.Name)) + "::" + strings.ToLower(strings.TrimSpace(t.Name))
			tagByGroupAndName[key] = t
			normTag := strings.ToLower(strings.TrimSpace(t.Name))
			tagByNameOnly[normTag] = append(tagByNameOnly[normTag], struct {
				GroupName string
				Tag       TagSpec
			}{GroupName: g.Name, Tag: t})
		}
	}

	resolveGroupAndTag := func(rawGroup, rawTag string) (string, TagSpec, bool) {
		tryResolve := func(grp, tg string) (string, TagSpec, bool) {
			if grp != "" {
				exactKey := strings.ToLower(grp) + "::" + strings.ToLower(tg)
				if ts, ok := tagByGroupAndName[exactKey]; ok {
					for _, g := range profile.Groups {
						if strings.EqualFold(g.Name, grp) {
							return g.Name, ts, true
						}
					}
					return grp, ts, true
				}
			}
			if matches := tagByNameOnly[strings.ToLower(tg)]; len(matches) == 1 {
				return matches[0].GroupName, matches[0].Tag, true
			} else if len(matches) > 1 && grp == "" {
				return matches[0].GroupName, matches[0].Tag, true
			}
			return "", TagSpec{}, false
		}

		if grp, ts, ok := tryResolve(rawGroup, rawTag); ok {
			return grp, ts, true
		}
		// Only split "Group/Tag" if verbatim match failed so tag names with "/" (e.g. "快讯/消息") stay intact.
		if strings.Contains(rawTag, "/") {
			parts := strings.SplitN(rawTag, "/", 2)
			if len(parts) == 2 {
				partGroup := strings.TrimSpace(parts[0])
				partTag := strings.TrimSpace(parts[1])
				if partGroup != "" && partTag != "" && (rawGroup == "" || strings.EqualFold(rawGroup, partGroup)) {
					exactKey := strings.ToLower(partGroup) + "::" + strings.ToLower(partTag)
					if ts, ok := tagByGroupAndName[exactKey]; ok {
						for _, g := range profile.Groups {
							if strings.EqualFold(g.Name, partGroup) {
								return g.Name, ts, true
							}
						}
						return partGroup, ts, true
					}
				}
			}
		}
		return "", TagSpec{}, false
	}

	threshold := float64(confidencePct) / 100.0
	if confidencePct <= 0 {
		threshold = 0.6
	}

	type candidateAssignment struct {
		Assignment TagAssignment
		Priority   int
	}

	var candidates []candidateAssignment
	seenTagID := make(map[uint]bool)
	seenKey := make(map[string]bool)
	belowThresholdCount := 0
	unmatchedCount := 0
	bestFilteredTag := ""
	bestFilteredConf := -1.0

	for _, item := range rawTags {
		rawGroup := strings.TrimSpace(item.Group)
		if rawGroup == "" {
			rawGroup = strings.TrimSpace(item.GroupName)
		}
		rawTag := strings.TrimSpace(item.Tag)
		if rawTag == "" {
			rawTag = strings.TrimSpace(item.TagName)
		}
		if rawTag == "" {
			continue
		}

		var conf float64
		if !item.Confidence.Set {
			conf = 0.85
		} else {
			conf = item.Confidence.Value
			if conf < 0 {
				conf = 0
			} else if conf > 1.0 && conf <= 100.0 {
				conf = conf / 100.0
			} else if conf > 100.0 {
				conf = 1.0
			}
		}

		resolvedGroup, resolvedTag, found := resolveGroupAndTag(rawGroup, rawTag)
		if !found {
			unmatchedCount++
			continue
		}

		if conf < threshold {
			belowThresholdCount++
			if conf > bestFilteredConf {
				bestFilteredConf = conf
				bestFilteredTag = resolvedGroup + "/" + resolvedTag.Name
			}
			continue
		}

		dedupeKey := resolvedGroup + "::" + resolvedTag.Name
		if (resolvedTag.ID > 0 && seenTagID[resolvedTag.ID]) || seenKey[dedupeKey] {
			continue
		}
		if resolvedTag.ID > 0 {
			seenTagID[resolvedTag.ID] = true
		}
		seenKey[dedupeKey] = true

		candidates = append(candidates, candidateAssignment{
			Assignment: TagAssignment{
				TagID:      resolvedTag.ID,
				GroupName:  resolvedGroup,
				TagName:    resolvedTag.Name,
				Confidence: math.Round(conf*10000) / 10000,
				Provider:   "llm",
				Reason:     strings.TrimSpace(item.Reason),
			},
			Priority: resolvedTag.Priority,
		})
	}

	const prioritySortWeight = 0.005
	sort.SliceStable(candidates, func(i, j int) bool {
		scoreI := candidates[i].Assignment.Confidence + float64(candidates[i].Priority)*prioritySortWeight
		scoreJ := candidates[j].Assignment.Confidence + float64(candidates[j].Priority)*prioritySortWeight
		if scoreI != scoreJ {
			return scoreI > scoreJ
		}
		if candidates[i].Priority != candidates[j].Priority {
			return candidates[i].Priority > candidates[j].Priority
		}
		return candidates[i].Assignment.Confidence > candidates[j].Assignment.Confidence
	})

	groupCounts := make(map[string]int)
	var final []TagAssignment
	for _, c := range candidates {
		gName := c.Assignment.GroupName
		limit := groupLimit[gName]
		if limit <= 0 {
			limit = 1
		}
		if groupCounts[gName] >= limit {
			continue
		}
		groupCounts[gName]++
		final = append(final, c.Assignment)
	}

	if len(final) == 0 {
		switch {
		case belowThresholdCount > 0:
			return nil, fmt.Errorf("no tags met confidence threshold %.2f (%d candidate(s) below threshold, highest: %s %.2f)", threshold, belowThresholdCount, bestFilteredTag, bestFilteredConf)
		case unmatchedCount > 0:
			return nil, fmt.Errorf("LLM returned %d tag(s), none matched configured taxonomy", unmatchedCount)
		default:
			return nil, errors.New("LLM returned 0 classification tags")
		}
	}
	return final, nil
}

func (l *LLMClassifier) ListOllamaModels(ctx context.Context, baseURL string) ([]OllamaModelInfo, error) {
	baseURL = strings.TrimRight(strings.TrimSpace(baseURL), "/")
	if baseURL == "" {
		baseURL = "http://127.0.0.1:11434"
	}
	endpoint := strings.TrimSuffix(baseURL, "/api/chat")
	endpoint = strings.TrimSuffix(endpoint, "/v1")
	endpoint += "/api/tags"

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, err
	}
	resp, err := l.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode >= 300 {
		raw, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
		return nil, fmt.Errorf("HTTP %d: %s", resp.StatusCode, strings.TrimSpace(string(raw)))
	}
	var parsed struct {
		Models []struct {
			Name       string `json:"name"`
			Size       int64  `json:"size"`
			ModifiedAt string `json:"modified_at"`
			Details    struct {
				Family            string `json:"family"`
				ParameterSize     string `json:"parameter_size"`
				QuantizationLevel string `json:"quantization_level"`
			} `json:"details"`
		} `json:"models"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&parsed); err != nil {
		return nil, err
	}
	out := make([]OllamaModelInfo, 0, len(parsed.Models))
	for _, m := range parsed.Models {
		out = append(out, OllamaModelInfo{
			Name:       m.Name,
			Size:       m.Size,
			ModifiedAt: m.ModifiedAt,
			Family:     m.Details.Family,
			ParamSize:  m.Details.ParameterSize,
			QuantLevel: m.Details.QuantizationLevel,
		})
	}
	return out, nil
}

func (l *LLMClassifier) PullOllamaModel(ctx context.Context, baseURL, modelName string, onProgress func(ProgressUpdate)) error {
	baseURL = strings.TrimRight(strings.TrimSpace(baseURL), "/")
	if baseURL == "" {
		baseURL = "http://127.0.0.1:11434"
	}
	modelName = strings.TrimSpace(modelName)
	if modelName == "" {
		return errors.New("model name is required")
	}
	endpoint := strings.TrimSuffix(baseURL, "/api/chat")
	endpoint = strings.TrimSuffix(endpoint, "/v1")
	endpoint += "/api/pull"

	pullCtx, cancel := context.WithTimeout(ctx, 60*time.Minute)
	defer cancel()

	if onProgress != nil {
		onProgress(ProgressUpdate{
			Stage:         "resolving",
			Message:       "pulling manifest",
			CurrentItem:   modelName,
			Percent:       -1,
			Indeterminate: true,
		})
	}

	pullClient := &http.Client{Timeout: 0}
	body, _ := json.Marshal(map[string]any{
		"name":   modelName,
		"stream": true,
	})
	req, err := http.NewRequestWithContext(pullCtx, http.MethodPost, endpoint, bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := pullClient.Do(req)
	if err != nil {
		return err
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode >= 300 {
		raw, _ := io.ReadAll(io.LimitReader(resp.Body, 64<<10))
		return fmt.Errorf("HTTP %d: %s", resp.StatusCode, strings.TrimSpace(string(raw)))
	}

	type layerState struct {
		total     int64
		completed int64
	}
	layers := make(map[string]layerState)
	lastStage := ""
	lastEmit := time.Time{}

	scanner := bufio.NewScanner(resp.Body)
	scanner.Buffer(make([]byte, 0, 64*1024), 2*1024*1024)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		var chunk struct {
			Status    string `json:"status"`
			Digest    string `json:"digest"`
			Total     int64  `json:"total"`
			Completed int64  `json:"completed"`
			Error     string `json:"error"`
		}
		if err := json.Unmarshal([]byte(line), &chunk); err != nil {
			continue
		}
		if strings.TrimSpace(chunk.Error) != "" {
			return errors.New(strings.TrimSpace(chunk.Error))
		}

		if chunk.Digest != "" && chunk.Total > 0 {
			comp := chunk.Completed
			if comp < 0 {
				comp = 0
			}
			if comp > chunk.Total {
				comp = chunk.Total
			}
			prev := layers[chunk.Digest]
			if chunk.Total > prev.total {
				prev.total = chunk.Total
			}
			if comp > prev.completed {
				prev.completed = comp
			}
			layers[chunk.Digest] = prev
		}

		var sumTotal, sumCompleted int64
		for _, ls := range layers {
			sumTotal += ls.total
			sumCompleted += ls.completed
		}

		lowerStatus := strings.ToLower(strings.TrimSpace(chunk.Status))
		stage := "downloading"
		pct := float64(-1)
		indeterminate := sumTotal <= 0

		switch {
		case lowerStatus == "success":
			stage = "ready"
			pct = 100
			indeterminate = false
		case strings.Contains(lowerStatus, "verifying"):
			stage = "verifying"
			if sumTotal > 0 {
				pct = 99
				indeterminate = false
			} else {
				indeterminate = true
			}
		case strings.Contains(lowerStatus, "writing manifest") || strings.Contains(lowerStatus, "removing"):
			stage = "finalizing"
			if sumTotal > 0 {
				pct = 99.5
				indeterminate = false
			} else {
				indeterminate = true
			}
		case strings.Contains(lowerStatus, "manifest") || (sumTotal <= 0 && !strings.HasPrefix(lowerStatus, "pulling ")):
			stage = "resolving"
			indeterminate = true
		default:
			stage = "downloading"
			if sumTotal > 0 {
				pct = (float64(sumCompleted) / float64(sumTotal)) * 99.0
				if pct > 99.0 {
					pct = 99.0
				}
				indeterminate = false
			}
		}

		currentItem := strings.TrimSpace(chunk.Digest)
		if currentItem != "" {
			short := strings.TrimPrefix(currentItem, "sha256:")
			if len(short) > 12 {
				short = short[:12]
			}
			currentItem = short
		} else {
			currentItem = modelName
		}

		now := time.Now()
		if onProgress != nil && (stage != lastStage || stage == "ready" || now.Sub(lastEmit) >= 120*time.Millisecond) {
			lastStage = stage
			lastEmit = now
			onProgress(ProgressUpdate{
				Stage:          stage,
				Message:        strings.TrimSpace(chunk.Status),
				CurrentItem:    currentItem,
				CompletedBytes: sumCompleted,
				TotalBytes:     sumTotal,
				Percent:        pct,
				Indeterminate:  indeterminate,
			})
		}
	}
	if err := scanner.Err(); err != nil {
		if pullCtx.Err() != nil {
			return pullCtx.Err()
		}
		return err
	}
	if pullCtx.Err() != nil {
		return pullCtx.Err()
	}
	return nil
}
