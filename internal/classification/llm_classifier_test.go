package classification

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func testClassificationProfile() ProfileSpec {
	return ProfileSpec{
		ID:   1,
		Name: "Default",
		Groups: []GroupSpec{
			{
				ID:            10,
				Name:          "Topic",
				SelectionMode: "single",
				MaxResults:    1,
				Tags: []TagSpec{
					{
						ID:          101,
						Name:        "Tech",
						Description: "Computing, AI and software engineering",
						Keywords:    []string{"chip", "GPU"},
					},
					{
						ID:          102,
						Name:        "Finance",
						Description: "Economy, central banks and markets",
						Keywords:    []string{"GDP", "Fed"},
					},
				},
			},
		},
	}
}

func testRichNewsProfile() ProfileSpec {
	return ProfileSpec{
		ID:   2,
		Name: "News",
		Groups: []GroupSpec{
			{
				ID:            1,
				Name:          "体裁",
				SelectionMode: "single",
				MaxResults:    1,
				Tags: []TagSpec{
					{ID: 11, Name: "快讯/消息", Priority: 10, Keywords: []string{"据悉", "消息"}},
					{ID: 12, Name: "深度/特稿", Priority: 9, Keywords: []string{"调查", "专访"}},
				},
			},
			{
				ID:            2,
				Name:          "领域",
				SelectionMode: "limit",
				MaxResults:    2,
				Tags: []TagSpec{
					{ID: 21, Name: "国际政治", Priority: 11, Keywords: []string{"外交", "领事"}},
					{ID: 22, Name: "Hard/Tech", Priority: 0, Keywords: []string{"chip"}},
				},
			},
			{
				ID:            3,
				Name:          "区域",
				SelectionMode: "limit",
				MaxResults:    2,
				Tags: []TagSpec{
					{ID: 31, Name: "中国", Priority: 10, Keywords: []string{"k1", "k2", "k3", "k4", "k5", "k6", "k7", "k8", "北京", "上海"}},
					{ID: 32, Name: "亚太", Priority: 9, Keywords: []string{"日本", "东京", "大阪", "韩国", "首尔", "印度", "东盟", "澳大利亚", "新加坡", "半岛"}},
				},
			},
		},
	}
}

func TestBuildLLMPromptsMarksKeywordsAsOptionalAndNoTruncation(t *testing.T) {
	rich := testRichNewsProfile()
	sys, usr := buildLLMPrompts("Quantum Processor Breakthrough", "TechNews", "Researchers built a fault-tolerant quantum circuit.", rich, "")
	if !strings.Contains(sys, "optional_keywords") {
		t.Fatalf("expected optional_keywords marker in prompt: %s", sys)
	}
	if !strings.Contains(sys, "NOT required match conditions") {
		t.Fatalf("expected explicit instruction that keywords are optional: %s", sys)
	}
	if !strings.Contains(sys, "[priority: 10]") || !strings.Contains(sys, "[priority: 9]") {
		t.Fatalf("expected nonzero tag priorities in prompt: %s", sys)
	}
	if strings.Contains(sys, "[priority: 0]") {
		t.Fatalf("did not expect [priority: 0] clutter in prompt: %s", sys)
	}
	if !strings.Contains(sys, "Prefer higher-priority and more specific tags over broader umbrella tags") {
		t.Fatalf("expected priority specificity guidance in prompt: %s", sys)
	}
	if strings.Contains(sys, "If no tag in a group clearly fits, omit that group.") {
		t.Fatalf("prompt should not encourage omitting relevant groups: %s", sys)
	}
	if !strings.Contains(sys, "北京") || !strings.Contains(sys, "上海") || !strings.Contains(sys, "新加坡") || !strings.Contains(sys, "半岛") {
		t.Fatalf("expected all configured keywords past the 8th item to appear in prompt: %s", sys)
	}
	if !strings.Contains(sys, `"groups"`) || !strings.Contains(sys, `<exact tag name from group 区域>`) {
		t.Fatalf("expected groups example structure with neutral placeholder in prompt: %s", sys)
	}
	if !strings.Contains(usr, "Quantum Processor Breakthrough") {
		t.Fatalf("expected title in user prompt: %s", usr)
	}
}

func TestBuildOllamaTaxonomySchema(t *testing.T) {
	rich := testRichNewsProfile()
	schema := buildOllamaTaxonomySchema(rich)
	if schema == nil {
		t.Fatal("expected non-nil schema")
	}
	props := schema["properties"].(map[string]any)
	groupsObj := props["groups"].(map[string]any)
	groupProps := groupsObj["properties"].(map[string]any)

	genreTags := groupProps["体裁"].(map[string]any)["properties"].(map[string]any)["tags"].(map[string]any)
	if genreTags["maxItems"] != 1 {
		t.Fatalf("expected single group maxItems=1, got %v", genreTags["maxItems"])
	}
	regionTags := groupProps["区域"].(map[string]any)["properties"].(map[string]any)["tags"].(map[string]any)
	if regionTags["maxItems"] != 2 {
		t.Fatalf("expected limit group maxItems=2, got %v", regionTags["maxItems"])
	}
	regionItemProps := regionTags["items"].(map[string]any)["properties"].(map[string]any)
	enumList := regionItemProps["tag"].(map[string]any)["enum"].([]string)
	if len(enumList) != 2 || enumList[0] != "中国" || enumList[1] != "亚太" {
		t.Fatalf("unexpected region enum list: %v", enumList)
	}

	emptySchema := buildOllamaTaxonomySchema(ProfileSpec{Groups: []GroupSpec{{Name: "Empty"}}})
	if emptySchema != nil {
		t.Fatalf("expected nil schema when no groups have tags, got %v", emptySchema)
	}
}

func TestParseLLMClassificationOutputDiagnosticsAndFormats(t *testing.T) {
	profile := testClassificationProfile()

	t.Run("accepts bare array and percentage confidence", func(t *testing.T) {
		raw := `[{"group":"Topic","tag":"Tech","confidence":"85%","reason":"Quantum hardware advance"}]`
		tags, err := parseLLMClassificationOutput(raw, profile, 60)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(tags) != 1 || tags[0].TagName != "Tech" || tags[0].Confidence != 0.85 {
			t.Fatalf("unexpected parsed tags: %+v", tags)
		}
	})

	t.Run("parses group-level reason-first schema and preserves slashes in tag names", func(t *testing.T) {
		rich := testRichNewsProfile()
		raw := `[Result] {"groups":{"体裁":{"reason":"突发外交消息报道","tags":[{"tag":"快讯/消息","confidence":0.95}]},"领域":{"reason":"涉及中日外交会谈","tags":[{"tag":"国际政治","confidence":0.96}]},"区域":{"reason":"涉及日本首相与中方在北京会谈","tags":[{"tag":"亚太","confidence":0.92},{"tag":"中国","confidence":0.90},{"tag":"亚太","confidence":0.85}]}}}`
		tags, err := parseLLMClassificationOutput(raw, rich, 60)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(tags) != 4 {
			t.Fatalf("expected 4 deduped tags across 3 groups, got %d: %+v", len(tags), tags)
		}
		foundSlashTag := false
		foundAP := false
		foundCN := false
		for _, tg := range tags {
			if tg.GroupName == "体裁" && tg.TagName == "快讯/消息" && tg.Reason == "突发外交消息报道" {
				foundSlashTag = true
			}
			if tg.GroupName == "区域" && tg.TagName == "亚太" && tg.Reason == "涉及日本首相与中方在北京会谈" {
				foundAP = true
			}
			if tg.GroupName == "区域" && tg.TagName == "中国" {
				foundCN = true
			}
		}
		if !foundSlashTag || !foundAP || !foundCN {
			t.Fatalf("expected slash tag and both bilateral region tags inheriting group reason, got %+v", tags)
		}
	})

	t.Run("parses unwrapped top-level group map with plain string tags", func(t *testing.T) {
		raw := `{"Topic":{"reason":"Discusses AI chips","tags":["Tech"]}}`
		tags, err := parseLLMClassificationOutput(raw, profile, 60)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(tags) != 1 || tags[0].TagName != "Tech" || tags[0].Confidence != 0.85 || tags[0].Reason != "Discusses AI chips" {
			t.Fatalf("unexpected unwrapped parsed tags: %+v", tags)
		}
	})

	t.Run("still supports legacy Group/Tag shorthand when verbatim match fails", func(t *testing.T) {
		raw := `{"tags":[{"tag":"Topic/Tech","confidence":0.9}]}`
		tags, err := parseLLMClassificationOutput(raw, profile, 60)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(tags) != 1 || tags[0].GroupName != "Topic" || tags[0].TagName != "Tech" {
			t.Fatalf("expected Topic/Tech shorthand resolution, got %+v", tags)
		}
	})

	t.Run("preserves explicit zero confidence and reports threshold failure", func(t *testing.T) {
		raw := `{"tags":[{"group":"Topic","tag":"Tech","confidence":0,"reason":"uncertain"}]}`
		_, err := parseLLMClassificationOutput(raw, profile, 60)
		if err == nil || !strings.Contains(err.Error(), "confidence threshold") {
			t.Fatalf("expected threshold failure error, got: %v", err)
		}
	})

	t.Run("defaults omitted confidence to 0.85", func(t *testing.T) {
		raw := `{"tags":[{"group":"Topic","tag":"Finance","reason":"central bank policy"}]}`
		tags, err := parseLLMClassificationOutput(raw, profile, 60)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(tags) != 1 || tags[0].Confidence != 0.85 {
			t.Fatalf("expected 0.85 default confidence, got %+v", tags)
		}
	})

	t.Run("reports zero tags error rather than silent empty success", func(t *testing.T) {
		_, err := parseLLMClassificationOutput(`{"tags":[]}`, profile, 60)
		if err == nil || !strings.Contains(err.Error(), "0 classification tags") {
			t.Fatalf("expected 0 tags error, got %v", err)
		}
	})

	t.Run("reports unmatched taxonomy error", func(t *testing.T) {
		raw := `{"tags":[{"group":"Topic","tag":"UnknownTag","confidence":0.9}]}`
		_, err := parseLLMClassificationOutput(raw, profile, 60)
		if err == nil || !strings.Contains(err.Error(), "none matched configured taxonomy") {
			t.Fatalf("expected unmatched taxonomy error, got %v", err)
		}
	})
}

func TestLLMClassifyItemOllamaSchemaAndFallback(t *testing.T) {
	t.Run("sends structured schema format and think=false on Ollama", func(t *testing.T) {
		var sawSchema bool
		var sawThinkFalse bool
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			var body map[string]any
			_ = json.NewDecoder(r.Body).Decode(&body)
			if _, ok := body["format"].(map[string]any); ok {
				sawSchema = true
			}
			if th, ok := body["think"].(bool); ok && !th {
				sawThinkFalse = true
			}
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"message":{"content":"{\"groups\":{\"Topic\":{\"reason\":\"Monetary tightening affects bond yields\",\"tags\":[{\"tag\":\"Finance\",\"confidence\":0.82}]}}}"}}`))
		}))
		defer srv.Close()

		cls := NewLLMClassifier()
		tags, err := cls.ClassifyItem(context.Background(), InputItem{
			Title:   "Treasury yields rise after policy statement",
			Content: "Fixed-income investors repriced long-duration sovereign debt overnight.",
		}, EngineOptions{
			Provider:    "llm",
			LLMProtocol: "ollama",
			LLMURL:      srv.URL,
			LLMModel:    "qwen3.5:2b",
			Confidence:  60,
			Profile:     testClassificationProfile(),
		})
		if err != nil {
			t.Fatalf("ClassifyItem failed: %v", err)
		}
		if !sawSchema || !sawThinkFalse {
			t.Fatalf("expected schema format and think=false, got schema=%v think=%v", sawSchema, sawThinkFalse)
		}
		if len(tags) != 1 || tags[0].TagName != "Finance" || tags[0].Reason == "" {
			t.Fatalf("expected semantic Finance tag with group reason, got %+v", tags)
		}
	})

	t.Run("falls back on HTTP 400 to json format without think", func(t *testing.T) {
		attempts := 0
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			attempts++
			var body map[string]any
			_ = json.NewDecoder(r.Body).Decode(&body)
			_, hasThink := body["think"]
			_, isSchema := body["format"].(map[string]any)
			if hasThink || isSchema {
				w.WriteHeader(http.StatusBadRequest)
				_, _ = w.Write([]byte(`{"error":"unsupported format schema or think field"}`))
				return
			}
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"message":{"content":"{\"Topic\":{\"reason\":\"Software engineering topic\",\"tags\":[\"Tech\"]}}"}}`))
		}))
		defer srv.Close()

		cls := NewLLMClassifier()
		tags, err := cls.ClassifyItem(context.Background(), InputItem{
			Title:   "Compiler Optimization Released",
			Content: "A new LLVM backend pass reduces memory pressure.",
		}, EngineOptions{
			Provider:    "llm",
			LLMProtocol: "ollama",
			LLMURL:      srv.URL,
			LLMModel:    "legacy-model",
			Confidence:  60,
			Profile:     testClassificationProfile(),
		})
		if err != nil {
			t.Fatalf("expected HTTP 400 fallback to succeed, got error: %v", err)
		}
		if attempts != 4 {
			t.Fatalf("expected 4 progressive attempts ending in plain json without think, got %d", attempts)
		}
		if len(tags) != 1 || tags[0].TagName != "Tech" || tags[0].Confidence != 0.85 {
			t.Fatalf("unexpected fallback tags: %+v", tags)
		}
	})
}

func TestParseLLMClassificationOutput_PriorityOrderingAndThreshold(t *testing.T) {
	profile := ProfileSpec{
		ID:   3,
		Name: "RegionPriority",
		Groups: []GroupSpec{
			{
				ID:            1,
				Name:          "区域",
				SelectionMode: "limit",
				MaxResults:    2,
				Tags: []TagSpec{
					{ID: 101, Name: "亚太", Priority: 6},
					{ID: 102, Name: "中国", Priority: 12},
					{ID: 103, Name: "日本", Priority: 12},
					{ID: 104, Name: "全球", Priority: 14},
				},
			},
		},
	}

	// 亚太 has slightly higher raw confidence (0.91) than 中国 (0.90) & 日本 (0.90),
	// but lower priority (6 vs 12): (0.90+0.06=0.96) > (0.91+0.03=0.94).
	// Meanwhile 全球 has highest priority (14) but is below threshold 0.60 so MUST be rejected!
	raw := `{"groups":{"区域":{"reason":"中日外交","tags":[{"tag":"亚太","confidence":0.91},{"tag":"中国","confidence":0.90},{"tag":"日本","confidence":0.90},{"tag":"全球","confidence":0.40}]}}}`
	tags, err := parseLLMClassificationOutput(raw, profile, 60)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(tags) != 2 {
		t.Fatalf("expected 2 selected tags, got %+v", tags)
	}
	if tags[0].TagName != "中国" || tags[1].TagName != "日本" {
		t.Fatalf("expected higher-priority 中国 and 日本 to outrank 亚太, got %+v", tags)
	}
	if tags[0].Confidence != 0.90 || tags[1].Confidence != 0.90 {
		t.Fatalf("expected uninflated confidence 0.90 preserved, got %+v", tags)
	}
}

func TestLLMClassifyItemRespectsMaxCharsAndStrategy(t *testing.T) {
	var seenUsers []string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var body struct {
			Messages []struct {
				Role    string `json:"role"`
				Content string `json:"content"`
			} `json:"messages"`
		}
		_ = json.NewDecoder(r.Body).Decode(&body)
		for _, m := range body.Messages {
			if m.Role == "user" {
				seenUsers = append(seenUsers, m.Content)
			}
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"message":{"content":"{\"groups\":{\"Topic\":{\"reason\":\"Matches finance\",\"tags\":[{\"tag\":\"Finance\",\"confidence\":0.9}]}}}"}}`))
	}))
	defer srv.Close()

	cls := NewLLMClassifier()
	item := InputItem{
		Title:   "HEAD_TITLE",
		Content: "HEAD_BODY_" + strings.Repeat("填充正文词句", 50) + "_TAIL_BODY",
	}

	// Head strategy with max 40 chars
	if _, err := cls.ClassifyItem(context.Background(), item, EngineOptions{
		Provider:         "llm",
		LLMProtocol:      "ollama",
		LLMURL:           srv.URL,
		LLMModel:         "qwen3.5:0.8b",
		Confidence:       60,
		MaxInputChars:    40,
		TruncateStrategy: TruncateStrategyHead,
		Profile:          testClassificationProfile(),
	}); err != nil {
		t.Fatalf("head ClassifyItem failed: %v", err)
	}

	// Tail strategy with max 40 chars
	if _, err := cls.ClassifyItem(context.Background(), item, EngineOptions{
		Provider:         "llm",
		LLMProtocol:      "ollama",
		LLMURL:           srv.URL,
		LLMModel:         "qwen3.5:0.8b",
		Confidence:       60,
		MaxInputChars:    40,
		TruncateStrategy: TruncateStrategyTail,
		Profile:          testClassificationProfile(),
	}); err != nil {
		t.Fatalf("tail ClassifyItem failed: %v", err)
	}

	// 0 = full text
	if _, err := cls.ClassifyItem(context.Background(), item, EngineOptions{
		Provider:      "llm",
		LLMProtocol:   "ollama",
		LLMURL:        srv.URL,
		LLMModel:      "qwen3.5:0.8b",
		Confidence:    60,
		MaxInputChars: 0,
		Profile:       testClassificationProfile(),
	}); err != nil {
		t.Fatalf("unlimited ClassifyItem failed: %v", err)
	}

	if len(seenUsers) != 3 {
		t.Fatalf("expected 3 captured user prompts, got %d", len(seenUsers))
	}
	if !strings.Contains(seenUsers[0], "HEAD_BODY") || strings.Contains(seenUsers[0], "_TAIL_BODY") {
		t.Fatalf("head user prompt unexpected: %s", seenUsers[0])
	}
	if !strings.Contains(seenUsers[1], "_TAIL_BODY") || strings.Contains(seenUsers[1], "HEAD_BODY") {
		t.Fatalf("tail user prompt unexpected: %s", seenUsers[1])
	}
	if !strings.Contains(seenUsers[2], "HEAD_BODY") || !strings.Contains(seenUsers[2], "_TAIL_BODY") {
		t.Fatalf("unlimited user prompt missing head or tail: %s", seenUsers[2])
	}
}
