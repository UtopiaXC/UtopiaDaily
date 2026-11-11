package scrapers

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"html"
	"io"
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/UtopiaXC/UtopiaDaily/internal/module"
)

const defaultBangumiUA = "UtopiaXC/UtopiaDaily/0.2.0 (https://github.com/UtopiaXC/UtopiaDaily)"

type BangumiScraper struct {
	client *http.Client
}

func NewBangumiScraper() *BangumiScraper {
	return &BangumiScraper{
		client: &http.Client{Timeout: 20 * time.Second},
	}
}

func (b *BangumiScraper) Meta() module.Metadata {
	return module.Metadata{
		ID:      "bangumi_anime",
		Kind:    module.KindScraper,
		Source:  module.SourceBuiltIn,
		Version: module.DefaultVersion,
		Author:  "UtopiaXC",
	}
}

func (b *BangumiScraper) ConfigSchema() []module.ConfigFieldSpec {
	return []module.ConfigFieldSpec{
		{Key: "username", Type: "string", DefaultValue: "", SortOrder: 1},
		{Key: "access_token", Type: "password", Secret: true, DefaultValue: "", SortOrder: 2},
		{Key: "include_watching", Type: "boolean", DefaultValue: "true", SortOrder: 3},
		{Key: "include_wish", Type: "boolean", DefaultValue: "false", SortOrder: 4},
		{Key: "include_collect_airing", Type: "boolean", DefaultValue: "true", SortOrder: 5},
		{Key: "user_agent", Type: "string", DefaultValue: defaultBangumiUA, SortOrder: 6},
	}
}

func (b *BangumiScraper) Tasks() []module.TaskSpec {
	return []module.TaskSpec{
		{Key: "fetch_calendar", DefaultCron: "0 8 * * *", DefaultEnabled: false},
		{Key: "fetch_collections", DefaultCron: "0 */3 * * *", DefaultEnabled: false},
	}
}

func (b *BangumiScraper) Locales() map[string]map[string]string {
	return map[string]map[string]string{
		"zh_CN": {
			"module.bangumi_anime.name":                               "Bangumi 番组计划",
			"module.bangumi_anime.description":                        "抓取 Bangumi 每日放送表及用户关注番剧的更新进度",
			"module.bangumi_anime.config.username.desc":               "Bangumi 用户名",
			"module.bangumi_anime.config.username.hint":               "用户主页 URL 中的用户名或 UID，抓取个人在看或想看列表时必填",
			"module.bangumi_anime.config.access_token.desc":           "Access Token",
			"module.bangumi_anime.config.access_token.hint":           "可选个人访问令牌，用于访问私密收藏与提高限额",
			"module.bangumi_anime.config.include_watching.desc":       "追踪在看动画",
			"module.bangumi_anime.config.include_watching.hint":       "包含用户状态为在看的连载或未完结动画",
			"module.bangumi_anime.config.include_wish.desc":           "追踪想看动画",
			"module.bangumi_anime.config.include_wish.hint":           "包含用户状态为想看的新番更新动态",
			"module.bangumi_anime.config.include_collect_airing.desc": "追踪看过但未完结动画",
			"module.bangumi_anime.config.include_collect_airing.hint": "包含已标记看过且仍在放送新话数的动画",
			"module.bangumi_anime.config.user_agent.desc":             "User-Agent",
			"module.bangumi_anime.config.user_agent.hint":             "请求 api.bgm.tv 时附带的客户端标识",
			"module.bangumi_anime.task.fetch_calendar.name":           "抓取今日放送",
			"module.bangumi_anime.task.fetch_calendar.desc":           "抓取 Bangumi 今日新番放送列表",
			"module.bangumi_anime.task.fetch_collections.name":        "抓取追番动态",
			"module.bangumi_anime.task.fetch_collections.desc":        "检查已配置用户收藏动画的最新已播剧集",
		},
		"en_US": {
			"module.bangumi_anime.name":                               "Bangumi Anime Tracker",
			"module.bangumi_anime.description":                        "Collects Bangumi daily airing schedule and user anime collection episode updates",
			"module.bangumi_anime.config.username.desc":               "Bangumi Username",
			"module.bangumi_anime.config.username.hint":               "Username or UID used when fetching user anime collections",
			"module.bangumi_anime.config.access_token.desc":           "Access Token",
			"module.bangumi_anime.config.access_token.hint":           "Optional personal access token for private collections",
			"module.bangumi_anime.config.include_watching.desc":       "Track Watching",
			"module.bangumi_anime.config.include_watching.hint":       "Include anime marked as currently watching",
			"module.bangumi_anime.config.include_wish.desc":           "Track Wish List",
			"module.bangumi_anime.config.include_wish.hint":           "Include anime marked as plan to watch",
			"module.bangumi_anime.config.include_collect_airing.desc": "Track Airing Completed Entries",
			"module.bangumi_anime.config.include_collect_airing.hint": "Include entries marked as watched that are still airing",
			"module.bangumi_anime.config.user_agent.desc":             "User-Agent",
			"module.bangumi_anime.config.user_agent.hint":             "Client identifier sent to api.bgm.tv",
			"module.bangumi_anime.task.fetch_calendar.name":           "Fetch Daily Calendar",
			"module.bangumi_anime.task.fetch_calendar.desc":           "Collect today's airing anime from Bangumi calendar",
			"module.bangumi_anime.task.fetch_collections.name":        "Fetch Collection Updates",
			"module.bangumi_anime.task.fetch_collections.desc":        "Check latest aired episodes for user anime collections",
		},
	}
}

func (b *BangumiScraper) Scrape(ctx context.Context, taskKey string, cfg module.ConfigAccessor) ([]module.ScrapedItem, error) {
	ua := strings.TrimSpace(cfg.GetString("user_agent"))
	if ua == "" {
		ua = defaultBangumiUA
	}
	token := strings.TrimSpace(cfg.GetString("access_token"))

	switch taskKey {
	case "fetch_calendar":
		return b.fetchCalendar(ctx, ua)
	case "fetch_collections":
		username := strings.TrimSpace(cfg.GetString("username"))
		if username == "" {
			return nil, errors.New("bangumi username is not configured")
		}
		return b.fetchCollections(
			ctx,
			ua,
			username,
			token,
			cfg.GetBool("include_watching"),
			cfg.GetBool("include_wish"),
			cfg.GetBool("include_collect_airing"),
		)
	default:
		return nil, fmt.Errorf("unknown task: %s", taskKey)
	}
}

func (b *BangumiScraper) TestConfig(ctx context.Context, cfg module.ConfigAccessor) module.TestResult {
	ua := strings.TrimSpace(cfg.GetString("user_agent"))
	if ua == "" {
		ua = defaultBangumiUA
	}
	token := strings.TrimSpace(cfg.GetString("access_token"))

	if _, err := b.getJSON(ctx, "https://api.bgm.tv/calendar", ua, ""); err != nil {
		return module.TestResult{Success: false, MessageKey: "module.test.failed", Detail: err.Error()}
	}

	username := strings.TrimSpace(cfg.GetString("username"))
	if username != "" {
		uEndpoint := "https://api.bgm.tv/v0/users/" + url.PathEscape(username)
		if _, err := b.getJSON(ctx, uEndpoint, ua, token); err != nil {
			return module.TestResult{Success: false, MessageKey: "module.test.failed", Detail: err.Error()}
		}
	}
	return module.TestResult{Success: true, MessageKey: "module.test.passed"}
}

func (b *BangumiScraper) getJSON(ctx context.Context, endpoint, ua, token string) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", ua)
	req.Header.Set("Accept", "application/json")
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	resp, err := b.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()
	raw, _ := io.ReadAll(io.LimitReader(resp.Body, 8<<20))
	if resp.StatusCode >= 300 {
		return nil, fmt.Errorf("bangumi API HTTP %d: %s", resp.StatusCode, strings.TrimSpace(string(raw)))
	}
	return raw, nil
}

type bgmCalendarDay struct {
	Weekday struct {
		ID int    `json:"id"`
		CN string `json:"cn"`
		EN string `json:"en"`
	} `json:"weekday"`
	Items []struct {
		ID      int    `json:"id"`
		URL     string `json:"url"`
		Name    string `json:"name"`
		NameCN  string `json:"name_cn"`
		AirDate string `json:"air_date"`
		Rating  *struct {
			Score float64 `json:"score"`
			Total int     `json:"total"`
		} `json:"rating"`
		Rank   int `json:"rank"`
		Images *struct {
			Large  string `json:"large"`
			Common string `json:"common"`
			Medium string `json:"medium"`
		} `json:"images"`
	} `json:"items"`
}

func (b *BangumiScraper) fetchCalendar(ctx context.Context, ua string) ([]module.ScrapedItem, error) {
	raw, err := b.getJSON(ctx, "https://api.bgm.tv/calendar", ua, "")
	if err != nil {
		return nil, err
	}
	var days []bgmCalendarDay
	if err := json.Unmarshal(raw, &days); err != nil {
		return nil, err
	}

	nowJST := time.Now().UTC().Add(8 * time.Hour)
	bgmWeekday := int(nowJST.Weekday())
	if bgmWeekday == 0 {
		bgmWeekday = 7
	}
	dateKey := nowJST.Format("2006-01-02")

	var items []module.ScrapedItem
	for _, day := range days {
		if day.Weekday.ID != bgmWeekday {
			continue
		}
		for _, sub := range day.Items {
			title := strings.TrimSpace(sub.NameCN)
			if title == "" {
				title = strings.TrimSpace(sub.Name)
			}
			if title == "" {
				continue
			}
			pageURL := strings.TrimSpace(sub.URL)
			if pageURL == "" {
				pageURL = fmt.Sprintf("https://bgm.tv/subject/%d", sub.ID)
			}
			var imgURL string
			if sub.Images != nil {
				imgURL = pickFirstNonEmpty(sub.Images.Large, sub.Images.Common, sub.Images.Medium)
			}
			scoreStr := "-"
			if sub.Rating != nil && sub.Rating.Score > 0 {
				scoreStr = fmt.Sprintf("%.1f", sub.Rating.Score)
			}
			desc := fmt.Sprintf("[%s] %s | Score: %s", day.Weekday.CN, sub.Name, scoreStr)
			if sub.Rank > 0 {
				desc += fmt.Sprintf(" | Rank #%d", sub.Rank)
			}
			if sub.AirDate != "" {
				desc += fmt.Sprintf(" | Air: %s", sub.AirDate)
			}
			bodyHTML := buildBangumiHTML(title, sub.Name, desc, pageURL, imgURL)
			sum := sha256.Sum256([]byte(fmt.Sprintf("bangumi_cal|%s|%d", dateKey, sub.ID)))

			var media []module.ScrapedMedia
			if imgURL != "" {
				media = append(media, module.ScrapedMedia{Type: "image", URL: imgURL})
			}

			items = append(items, module.ScrapedItem{
				Title:       title,
				Summary:     desc,
				Content:     bodyHTML,
				ContentType: "html",
				RenderType:  "html",
				SourceName:  "Bangumi Daily",
				Author:      day.Weekday.CN,
				ExternalURL: pageURL,
				PublishedAt: time.Now().UTC(),
				Fingerprint: hex.EncodeToString(sum[:]),
				Media:       media,
			})
		}
	}
	return items, nil
}

func (b *BangumiScraper) fetchCollections(ctx context.Context, ua, username, token string, watching, wish, collectAiring bool) ([]module.ScrapedItem, error) {
	type targetCol struct {
		colType int
		label   string
	}
	var targets []targetCol
	if watching {
		targets = append(targets, targetCol{colType: 3, label: "Watching"})
	}
	if wish {
		targets = append(targets, targetCol{colType: 1, label: "Wish"})
	}
	if collectAiring {
		targets = append(targets, targetCol{colType: 2, label: "Airing"})
	}
	if len(targets) == 0 {
		return nil, nil
	}

	nowUTC := time.Now().UTC()
	todayStr := nowUTC.Add(8 * time.Hour).Format("2006-01-02")
	recentWindow := nowUTC.AddDate(0, 0, -21).Format("2006-01-02")

	var out []module.ScrapedItem
	seenSubject := make(map[int]bool)

	for _, tg := range targets {
		endpoint := fmt.Sprintf("https://api.bgm.tv/v0/users/%s/collections?subject_type=2&type=%d&limit=30", url.PathEscape(username), tg.colType)
		raw, err := b.getJSON(ctx, endpoint, ua, token)
		if err != nil {
			return nil, err
		}
		var parsed struct {
			Data []struct {
				SubjectID int    `json:"subject_id"`
				EpStatus  int    `json:"ep_status"`
				Type      int    `json:"type"`
				UpdatedAt string `json:"updated_at"`
				Subject   struct {
					ID     int     `json:"id"`
					Name   string  `json:"name"`
					NameCN string  `json:"name_cn"`
					Date   string  `json:"date"`
					Eps    int     `json:"eps"`
					Score  float64 `json:"score"`
					Short  string  `json:"short_summary"`
					Images struct {
						Large  string `json:"large"`
						Common string `json:"common"`
						Medium string `json:"medium"`
					} `json:"images"`
				} `json:"subject"`
			} `json:"data"`
		}
		if err := json.Unmarshal(raw, &parsed); err != nil {
			return nil, err
		}

		for _, entry := range parsed.Data {
			sid := entry.SubjectID
			if sid == 0 || seenSubject[sid] {
				continue
			}

			latestEp, hasFutureOrRecent, err := b.fetchLatestEpisode(ctx, ua, token, sid, todayStr, recentWindow)
			if err != nil {
				continue
			}
			// For "Watched / Collected" (type 2), strictly keep ongoing/airing shows only.
			if tg.colType == 2 && !hasFutureOrRecent && (entry.Subject.Eps > 0 && latestEp != nil && int(latestEp.Sort) >= entry.Subject.Eps) {
				continue
			}

			seenSubject[sid] = true
			title := pickFirstNonEmpty(entry.Subject.NameCN, entry.Subject.Name)
			if title == "" {
				continue
			}
			imgURL := pickFirstNonEmpty(entry.Subject.Images.Large, entry.Subject.Images.Common, entry.Subject.Images.Medium)
			pageURL := fmt.Sprintf("https://bgm.tv/subject/%d", sid)

			fpKey := fmt.Sprintf("bangumi_col|%d|base", sid)
			pubTime := nowUTC
			epSummary := fmt.Sprintf("[%s] Progress %d/%d", tg.label, entry.EpStatus, entry.Subject.Eps)

			if latestEp != nil {
				epSort := strings.TrimRight(strings.TrimRight(fmt.Sprintf("%.1f", latestEp.Sort), "0"), ".")
				epTitle := pickFirstNonEmpty(latestEp.NameCN, latestEp.Name)
				if epTitle != "" {
					title = fmt.Sprintf("%s EP %s: %s", title, epSort, epTitle)
				} else {
					title = fmt.Sprintf("%s EP %s", title, epSort)
				}
				fpKey = fmt.Sprintf("bangumi_col|%d|ep|%s", sid, epSort)
				epSummary = fmt.Sprintf("[%s] EP %s aired %s | Progress %d/%d", tg.label, epSort, latestEp.AirDate, entry.EpStatus, entry.Subject.Eps)
				if parsedDate, dErr := time.Parse("2006-01-02", latestEp.AirDate); dErr == nil {
					pubTime = parsedDate.UTC()
				}
			}

			desc := epSummary
			if entry.Subject.Short != "" {
				desc += " — " + entry.Subject.Short
			}
			bodyHTML := buildBangumiHTML(title, entry.Subject.Name, desc, pageURL, imgURL)
			sum := sha256.Sum256([]byte(fpKey))

			var media []module.ScrapedMedia
			if imgURL != "" {
				media = append(media, module.ScrapedMedia{Type: "image", URL: imgURL})
			}

			out = append(out, module.ScrapedItem{
				Title:       title,
				Summary:     buildSummary(desc),
				Content:     bodyHTML,
				ContentType: "html",
				RenderType:  "html",
				SourceName:  "Bangumi Collections",
				Author:      username,
				ExternalURL: pageURL,
				PublishedAt: pubTime,
				Fingerprint: hex.EncodeToString(sum[:]),
				Media:       media,
			})
		}
	}
	return out, nil
}

type bgmEpisode struct {
	ID      int     `json:"id"`
	Sort    float64 `json:"sort"`
	Name    string  `json:"name"`
	NameCN  string  `json:"name_cn"`
	AirDate string  `json:"airdate"`
}

func (b *BangumiScraper) fetchLatestEpisode(ctx context.Context, ua, token string, subjectID int, todayStr, recentWindow string) (*bgmEpisode, bool, error) {
	endpoint := fmt.Sprintf("https://api.bgm.tv/v0/episodes?subject_id=%d&type=0&limit=100", subjectID)
	raw, err := b.getJSON(ctx, endpoint, ua, token)
	if err != nil {
		return nil, false, err
	}
	var parsed struct {
		Data []bgmEpisode `json:"data"`
	}
	if err := json.Unmarshal(raw, &parsed); err != nil {
		return nil, false, err
	}
	if len(parsed.Data) == 0 {
		return nil, false, nil
	}

	sort.Slice(parsed.Data, func(i, j int) bool {
		return parsed.Data[i].Sort < parsed.Data[j].Sort
	})

	var latestAired *bgmEpisode
	hasFuture := false
	for idx := range parsed.Data {
		ep := &parsed.Data[idx]
		d := strings.TrimSpace(ep.AirDate)
		if d == "" {
			hasFuture = true
			continue
		}
		if d <= todayStr {
			latestAired = ep
		} else {
			hasFuture = true
		}
	}
	isRecent := latestAired != nil && latestAired.AirDate >= recentWindow
	return latestAired, hasFuture || isRecent, nil
}

func buildBangumiHTML(title, origName, summary, pageURL, imgURL string) string {
	var sb strings.Builder
	sb.WriteString(`<div class="bangumi-card">`)
	if imgURL != "" {
		sb.WriteString(fmt.Sprintf(`<p><img src="%s" alt="%s" style="max-width:240px;border-radius:4px;" /></p>`, html.EscapeString(imgURL), html.EscapeString(title)))
	}
	if origName != "" && origName != title {
		sb.WriteString(fmt.Sprintf(`<p style="color:#666;font-size:13px;">%s</p>`, html.EscapeString(origName)))
	}
	if summary != "" {
		sb.WriteString(fmt.Sprintf(`<p>%s</p>`, html.EscapeString(summary)))
	}
	if pageURL != "" {
		sb.WriteString(fmt.Sprintf(`<p><a href="%s" target="_blank" rel="noopener noreferrer">Bangumi Subject #%s</a></p>`, html.EscapeString(pageURL), html.EscapeString(strconv.Itoa(extractSubjectID(pageURL)))))
	}
	sb.WriteString(`</div>`)
	return sb.String()
}

func extractSubjectID(u string) int {
	parts := strings.Split(strings.TrimRight(u, "/"), "/")
	if len(parts) > 0 {
		if id, err := strconv.Atoi(parts[len(parts)-1]); err == nil {
			return id
		}
	}
	return 0
}

func pickFirstNonEmpty(vals ...string) string {
	for _, v := range vals {
		if t := strings.TrimSpace(v); t != "" {
			return t
		}
	}
	return ""
}
