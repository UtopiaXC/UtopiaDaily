package scrapers

import (
	"context"
	"crypto/hmac"
	"crypto/md5"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"html"
	"io"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"path"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/UtopiaXC/UtopiaDaily/internal/module"
)

const biliDesktopUA = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/131.0.0.0 Safari/537.36"

var wbiMixinKeyEncTab = []int{
	46, 47, 18, 2, 53, 8, 23, 32, 15, 50, 10, 31, 58, 3, 45, 35, 27, 43, 5, 49,
	33, 9, 42, 19, 29, 28, 14, 39, 12, 38, 41, 13, 37, 48, 7, 16, 24, 55, 40,
	61, 26, 17, 0, 1, 60, 51, 30, 4, 22, 25, 54, 21, 56, 59, 6, 63, 57, 62, 11,
	36, 20, 34, 44, 52,
}

type BilibiliScraper struct{}

func NewBilibiliScraper() *BilibiliScraper {
	return &BilibiliScraper{}
}

func (b *BilibiliScraper) Meta() module.Metadata {
	return module.Metadata{
		ID:      "bilibili_ranking_and_update",
		Kind:    module.KindScraper,
		Source:  module.SourceBuiltIn,
		Version: module.DefaultVersion,
		Author:  "UtopiaXC",
	}
}

func (b *BilibiliScraper) ConfigSchema() []module.ConfigFieldSpec {
	return []module.ConfigFieldSpec{
		{Key: "uploader_id", Type: "array", DefaultItems: []string{}, SortOrder: 1},
		{Key: "ranking_limit", Type: "int", DefaultValue: "20", SortOrder: 2},
		{Key: "cookie", Type: "password", Secret: true, DefaultValue: "", SortOrder: 3},
	}
}

func (b *BilibiliScraper) Tasks() []module.TaskSpec {
	return []module.TaskSpec{
		{Key: "fetch_ranking", DefaultCron: "0 */2 * * *", DefaultEnabled: false},
		{Key: "fetch_update", DefaultCron: "*/30 * * * *", DefaultEnabled: false},
	}
}

func (b *BilibiliScraper) Locales() map[string]map[string]string {
	return map[string]map[string]string{
		"zh_CN": {
			"module.bilibili_ranking_and_update.name":                      "Bilibili 热榜与更新",
			"module.bilibili_ranking_and_update.description":               "抓取 Bilibili 全站热门排行及关注 UP 主投稿更新",
			"module.bilibili_ranking_and_update.config.uploader_id.desc":   "UP 主 UID",
			"module.bilibili_ranking_and_update.config.uploader_id.hint":   "输入 Bilibili UP 主数字 UID",
			"module.bilibili_ranking_and_update.config.ranking_limit.desc": "热榜条目数",
			"module.bilibili_ranking_and_update.config.ranking_limit.hint": "单次保留的热榜视频上限",
			"module.bilibili_ranking_and_update.config.cookie.desc":        "账号 Cookie",
			"module.bilibili_ranking_and_update.config.cookie.hint":        "可选配置或通过下方扫码登录自动写入，用于降低风控拦截",
			"module.bilibili_ranking_and_update.task.fetch_ranking.name":   "抓取热榜",
			"module.bilibili_ranking_and_update.task.fetch_ranking.desc":   "抓取 Bilibili 全站热门视频排行榜",
			"module.bilibili_ranking_and_update.task.fetch_update.name":    "抓取投稿更新",
			"module.bilibili_ranking_and_update.task.fetch_update.desc":    "抓取已配置 UP 主的最新视频",
		},
		"en_US": {
			"module.bilibili_ranking_and_update.name":                      "Bilibili Ranking & Updates",
			"module.bilibili_ranking_and_update.description":               "Collects Bilibili hot rankings and uploader latest videos",
			"module.bilibili_ranking_and_update.config.uploader_id.desc":   "Uploader UIDs",
			"module.bilibili_ranking_and_update.config.uploader_id.hint":   "Numeric Bilibili uploader user IDs",
			"module.bilibili_ranking_and_update.config.ranking_limit.desc": "Ranking Limit",
			"module.bilibili_ranking_and_update.config.ranking_limit.hint": "Maximum videos kept from hot ranking",
			"module.bilibili_ranking_and_update.config.cookie.desc":        "Account Cookie",
			"module.bilibili_ranking_and_update.config.cookie.hint":        "Optional Cookie header populated automatically by QR sign-in to reduce anti-bot blocks",
			"module.bilibili_ranking_and_update.task.fetch_ranking.name":   "Fetch Ranking",
			"module.bilibili_ranking_and_update.task.fetch_ranking.desc":   "Collect Bilibili overall hot video ranking",
			"module.bilibili_ranking_and_update.task.fetch_update.name":    "Fetch Updates",
			"module.bilibili_ranking_and_update.task.fetch_update.desc":    "Collect latest videos from configured uploaders",
		},
	}
}

func (b *BilibiliScraper) Scrape(ctx context.Context, taskKey string, cfg module.ConfigAccessor) ([]module.ScrapedItem, error) {
	client := b.newSessionClient(ctx, cfg.GetString("cookie"))
	switch taskKey {
	case "fetch_ranking":
		limit := cfg.GetInt("ranking_limit", 20)
		if limit <= 0 {
			limit = 20
		}
		return b.fetchRanking(ctx, client, limit)
	case "fetch_update":
		uids := cfg.GetStringSlice("uploader_id")
		return b.fetchUpdates(ctx, client, uids)
	default:
		return nil, fmt.Errorf("unknown task: %s", taskKey)
	}
}

func (b *BilibiliScraper) TestConfig(ctx context.Context, cfg module.ConfigAccessor) module.TestResult {
	client := b.newSessionClient(ctx, cfg.GetString("cookie"))
	req, err := b.newRequest(ctx, "https://api.bilibili.com/x/web-interface/ranking/v2?rid=0&type=all", "")
	if err != nil {
		return module.TestResult{Success: false, MessageKey: "module.test.failed", Detail: err.Error()}
	}
	resp, err := client.Do(req)
	if err != nil {
		return module.TestResult{Success: false, MessageKey: "module.test.failed", Detail: err.Error()}
	}
	var payload struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
	}
	_ = json.NewDecoder(resp.Body).Decode(&payload)
	_ = resp.Body.Close()
	if payload.Code != 0 {
		return module.TestResult{Success: false, MessageKey: "module.test.failed", Detail: fmt.Sprintf("HTTP %d code %d %s", resp.StatusCode, payload.Code, payload.Message)}
	}

	uids := cfg.GetStringSlice("uploader_id")
	for _, rawUID := range uids {
		uid := strings.TrimSpace(rawUID)
		if uid == "" {
			continue
		}
		if _, updErr := b.fetchSingleUploader(ctx, client, uid, "", ""); updErr != nil {
			return module.TestResult{Success: false, MessageKey: "module.test.failed", Detail: fmt.Sprintf("UID %s check failed: %v", uid, updErr)}
		}
		break
	}

	return module.TestResult{Success: true, MessageKey: "module.test.passed"}
}

type parsedCookiePair struct {
	Name  string
	Value string
}

var biliCookieAttributes = map[string]bool{
	"path":     true,
	"domain":   true,
	"expires":  true,
	"max-age":  true,
	"samesite": true,
	"secure":   true,
	"httponly": true,
	"priority": true,
}

var biliBrowserFingerprintCookies = map[string]bool{
	"buvid3":              true,
	"buvid4":              true,
	"buvid_fp":            true,
	"buvid_fp_plain":      true,
	"b_nut":               true,
	"b_lsid":              true,
	"_uuid":               true,
	"bili_ticket":         true,
	"bili_ticket_expires": true,
	"rpdid":               true,
	"browser_resolution":  true,
	"enable_web_push":     true,
}

func cleanCookieToken(s string) string {
	s = strings.TrimSpace(s)
	return strings.Map(func(r rune) rune {
		if r < 0x20 || r == 0x7f {
			return -1
		}
		return r
	}, s)
}

func unwrapCookieValue(v string) string {
	v = cleanCookieToken(v)
	if len(v) >= 2 {
		if (v[0] == '"' && v[len(v)-1] == '"') || (v[0] == '\'' && v[len(v)-1] == '\'') {
			v = cleanCookieToken(v[1 : len(v)-1])
		}
	}
	return v
}

func stripCookieHeaderPrefix(s string) string {
	trimmed := strings.TrimSpace(s)
	lower := strings.ToLower(trimmed)
	if strings.HasPrefix(lower, "cookie:") {
		return strings.TrimSpace(trimmed[len("cookie:"):])
	}
	return trimmed
}

// parseCookiePairs supports standard Cookie headers, optional "Cookie:" prefixes,
// multi-line header dumps, quoted wrappers, and DevTools tab-separated cookie tables.
func parseCookiePairs(rawCookie string) []parsedCookiePair {
	cleaned := strings.TrimPrefix(rawCookie, "\uFEFF")
	cleaned = strings.TrimSpace(cleaned)
	if cleaned == "" {
		return nil
	}
	if len(cleaned) >= 2 &&
		((cleaned[0] == '"' && cleaned[len(cleaned)-1] == '"' && strings.Count(cleaned, `"`) == 2) ||
			(cleaned[0] == '\'' && cleaned[len(cleaned)-1] == '\'' && strings.Count(cleaned, `'`) == 2)) {
		cleaned = strings.TrimSpace(cleaned[1 : len(cleaned)-1])
	}
	cleaned = strings.ReplaceAll(cleaned, "\r\n", "\n")
	cleaned = strings.ReplaceAll(cleaned, "\r", "\n")

	order := make([]string, 0)
	values := make(map[string]string)
	recordPair := func(k, v string) {
		k = cleanCookieToken(k)
		v = unwrapCookieValue(v)
		if k == "" || v == "" || biliCookieAttributes[strings.ToLower(k)] {
			return
		}
		if _, exists := values[k]; !exists {
			order = append(order, k)
		}
		values[k] = v
	}

	for _, rawLine := range strings.Split(cleaned, "\n") {
		line := stripCookieHeaderPrefix(rawLine)
		if line == "" {
			continue
		}
		if !strings.Contains(line, ";") && strings.Contains(line, "\t") && !strings.Contains(line, "=") {
			cols := strings.Split(line, "\t")
			if len(cols) >= 2 {
				k := cleanCookieToken(cols[0])
				if !strings.EqualFold(k, "name") {
					recordPair(k, cols[1])
				}
			}
			continue
		}
		for _, part := range strings.Split(line, ";") {
			part = stripCookieHeaderPrefix(part)
			if part == "" {
				continue
			}
			eq := strings.IndexByte(part, '=')
			if eq <= 0 {
				continue
			}
			recordPair(part[:eq], part[eq+1:])
		}
	}

	out := make([]parsedCookiePair, 0, len(order))
	for _, k := range order {
		out = append(out, parsedCookiePair{Name: k, Value: values[k]})
	}
	return out
}

func hasCookieInJar(jar http.CookieJar, targetURL *url.URL, name string) bool {
	if jar == nil || targetURL == nil {
		return false
	}
	for _, ck := range jar.Cookies(targetURL) {
		if ck.Name == name && strings.TrimSpace(ck.Value) != "" {
			return true
		}
	}
	return false
}

func (b *BilibiliScraper) newSessionClient(ctx context.Context, rawCookie string) *http.Client {
	return b.newSessionClientWithTransport(ctx, rawCookie, nil)
}

func (b *BilibiliScraper) newSessionClientWithTransport(ctx context.Context, rawCookie string, rt http.RoundTripper) *http.Client {
	jar, _ := cookiejar.New(nil)
	client := &http.Client{
		Timeout:   20 * time.Second,
		Jar:       jar,
		Transport: rt,
	}

	rootURL, _ := url.Parse("https://www.bilibili.com/")
	pairs := parseCookiePairs(rawCookie)
	cookieMap := make(map[string]string, len(pairs))
	for _, p := range pairs {
		cookieMap[p.Name] = p.Value
	}

	hasSESSDATA := cookieMap["SESSDATA"] != ""
	biliJCT := cookieMap["bili_jct"]

	// When authenticated cookies (SESSDATA) are supplied by pasting from a browser,
	// exclude browser-bound fingerprint tokens (bili_ticket, buvid3, buvid4, b_nut, etc.)
	// upfront so this Go HTTP client mints matching fresh tokens via Bilibili's web APIs.
	var initialCookies []*http.Cookie
	for _, p := range pairs {
		if hasSESSDATA && biliBrowserFingerprintCookies[strings.ToLower(p.Name)] {
			continue
		}
		initialCookies = append(initialCookies, &http.Cookie{
			Name:   p.Name,
			Value:  p.Value,
			Domain: ".bilibili.com",
			Path:   "/",
		})
	}
	if len(initialCookies) > 0 {
		jar.SetCookies(rootURL, initialCookies)
	}

	if req, err := b.newRequest(ctx, "https://www.bilibili.com/", ""); err == nil {
		if resp, err := client.Do(req); err == nil {
			_ = resp.Body.Close()
		}
	}

	if req, err := b.newRequest(ctx, "https://api.bilibili.com/x/frontend/finger/spi", ""); err == nil {
		if resp, err := client.Do(req); err == nil {
			var spi struct {
				Code int `json:"code"`
				Data struct {
					B3 string `json:"b_3"`
					B4 string `json:"b_4"`
				} `json:"data"`
			}
			_ = json.NewDecoder(resp.Body).Decode(&spi)
			_ = resp.Body.Close()
			if spi.Code == 0 && spi.Data.B3 != "" {
				var spiCookies []*http.Cookie
				spiCookies = append(spiCookies, &http.Cookie{
					Name:   "buvid3",
					Value:  spi.Data.B3,
					Domain: ".bilibili.com",
					Path:   "/",
				})
				if spi.Data.B4 != "" {
					spiCookies = append(spiCookies, &http.Cookie{
						Name:   "buvid4",
						Value:  spi.Data.B4,
						Domain: ".bilibili.com",
						Path:   "/",
					})
				}
				jar.SetCookies(rootURL, spiCookies)
			}
		}
	}

	// Fall back to user-supplied buvid3/buvid4 only if /spi failed to populate buvid3.
	if !hasCookieInJar(jar, rootURL, "buvid3") && cookieMap["buvid3"] != "" {
		fallbackBuvid := []*http.Cookie{{
			Name:   "buvid3",
			Value:  cookieMap["buvid3"],
			Domain: ".bilibili.com",
			Path:   "/",
		}}
		if cookieMap["buvid4"] != "" {
			fallbackBuvid = append(fallbackBuvid, &http.Cookie{
				Name:   "buvid4",
				Value:  cookieMap["buvid4"],
				Domain: ".bilibili.com",
				Path:   "/",
			})
		}
		jar.SetCookies(rootURL, fallbackBuvid)
	}

	// Always attempt to generate a fresh web ticket bound to this client session;
	// fall back to pasted bili_ticket only if ticket generation fails.
	if !b.obtainWebTicket(ctx, client, rootURL, biliJCT) && cookieMap["bili_ticket"] != "" {
		jar.SetCookies(rootURL, []*http.Cookie{{
			Name:   "bili_ticket",
			Value:  cookieMap["bili_ticket"],
			Domain: ".bilibili.com",
			Path:   "/",
		}})
	}

	return client
}

func (b *BilibiliScraper) obtainWebTicket(ctx context.Context, client *http.Client, rootURL *url.URL, csrf string) bool {
	ts := strconv.FormatInt(time.Now().Unix(), 10)
	mac := hmac.New(sha256.New, []byte("XgwSnGZ1p"))
	_, _ = mac.Write([]byte("ts" + ts))
	hexSign := hex.EncodeToString(mac.Sum(nil))

	q := url.Values{}
	q.Set("key_id", "ec02")
	q.Set("hexsign", hexSign)
	q.Set("context[ts]", ts)
	q.Set("csrf", csrf)

	endpoint := "https://api.bilibili.com/bapis/bilibili.api.ticket.v1.Ticket/GenWebTicket?" + q.Encode()
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, nil)
	if err != nil {
		return false
	}
	req.Header.Set("User-Agent", biliDesktopUA)
	req.Header.Set("Referer", "https://www.bilibili.com/")
	req.Header.Set("Origin", "https://www.bilibili.com")

	resp, err := client.Do(req)
	if err != nil {
		return false
	}
	defer func() { _ = resp.Body.Close() }()
	var parsed struct {
		Code int `json:"code"`
		Data struct {
			Ticket string `json:"ticket"`
		} `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&parsed); err == nil && parsed.Code == 0 && parsed.Data.Ticket != "" {
		client.Jar.SetCookies(rootURL, []*http.Cookie{
			{
				Name:   "bili_ticket",
				Value:  parsed.Data.Ticket,
				Domain: ".bilibili.com",
				Path:   "/",
			},
			{
				Name:   "bili_ticket_expires",
				Value:  strconv.FormatInt(time.Now().Unix()+259200, 10),
				Domain: ".bilibili.com",
				Path:   "/",
			},
		})
		return true
	}
	return false
}

func (b *BilibiliScraper) newRequest(ctx context.Context, targetURL, referer string) (*http.Request, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, targetURL, nil)
	if err != nil {
		return nil, err
	}
	if referer == "" {
		referer = "https://www.bilibili.com/"
	}
	req.Header.Set("User-Agent", biliDesktopUA)
	req.Header.Set("Referer", referer)
	req.Header.Set("Origin", "https://www.bilibili.com")
	req.Header.Set("Accept", "application/json, text/plain, */*")
	req.Header.Set("Accept-Language", "zh-CN,zh;q=0.9,en;q=0.8")
	return req, nil
}

func (b *BilibiliScraper) fetchRanking(ctx context.Context, client *http.Client, limit int) ([]module.ScrapedItem, error) {
	req, err := b.newRequest(ctx, "https://api.bilibili.com/x/web-interface/ranking/v2?rid=0&type=all", "")
	if err != nil {
		return nil, err
	}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()

	var parsed struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
		Data    struct {
			List []struct {
				Bvid    string `json:"bvid"`
				Title   string `json:"title"`
				Pic     string `json:"pic"`
				Desc    string `json:"desc"`
				Pubdate int64  `json:"pubdate"`
				Owner   struct {
					Name string `json:"name"`
				} `json:"owner"`
				Stat struct {
					View  int `json:"view"`
					Like  int `json:"like"`
					Coin  int `json:"coin"`
					Reply int `json:"reply"`
				} `json:"stat"`
			} `json:"list"`
		} `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&parsed); err != nil {
		return nil, err
	}
	if parsed.Code != 0 {
		return nil, fmt.Errorf("bilibili ranking api error: %d %s", parsed.Code, parsed.Message)
	}

	items := make([]module.ScrapedItem, 0, limit)
	for i, v := range parsed.Data.List {
		if i >= limit {
			break
		}
		videoURL := "https://www.bilibili.com/video/" + v.Bvid
		pubTime := time.Unix(v.Pubdate, 0).UTC()
		bodyHTML := fmt.Sprintf(
			`<div class="bilibili-video"><div style="position:relative;padding-bottom:56.25%%;height:0;overflow:hidden;border-radius:8px;margin-bottom:10px;"><iframe src="https://player.bilibili.com/player.html?bvid=%s&autoplay=0" scrolling="no" frameborder="0" allowfullscreen="true" style="position:absolute;top:0;left:0;width:100%%;height:100%%;"></iframe></div><p style="font-size:13px;color:#555;">Views: %d | Likes: %d | Coins: %d | Comments: %d</p><p>%s</p></div>`,
			html.EscapeString(v.Bvid), v.Stat.View, v.Stat.Like, v.Stat.Coin, v.Stat.Reply, html.EscapeString(v.Desc),
		)

		sum := sha256.Sum256([]byte("bilibili_ranking|" + v.Bvid))
		var media []module.ScrapedMedia
		if v.Pic != "" {
			media = append(media, module.ScrapedMedia{Type: "image", URL: normalizeBiliMediaURL(v.Pic)})
		}
		items = append(items, module.ScrapedItem{
			Title:       v.Title,
			Summary:     buildSummary(v.Desc),
			Content:     bodyHTML,
			ContentType: "html",
			RenderType:  "html",
			SourceName:  "Bilibili Hot",
			Author:      v.Owner.Name,
			ExternalURL: videoURL,
			PublishedAt: pubTime,
			Fingerprint: hex.EncodeToString(sum[:]),
			Media:       media,
		})
	}
	return items, nil
}

func (b *BilibiliScraper) fetchWBIKeys(ctx context.Context, client *http.Client) (imgKey, subKey string, err error) {
	req, err := b.newRequest(ctx, "https://api.bilibili.com/x/web-interface/nav", "")
	if err != nil {
		return "", "", err
	}
	resp, err := client.Do(req)
	if err != nil {
		return "", "", err
	}
	defer func() { _ = resp.Body.Close() }()

	raw, err := io.ReadAll(io.LimitReader(resp.Body, 2<<20))
	if err != nil {
		return "", "", err
	}
	var parsed struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
		Data    struct {
			WbiImg struct {
				ImgURL string `json:"img_url"`
				SubURL string `json:"sub_url"`
			} `json:"wbi_img"`
		} `json:"data"`
	}
	if err := json.Unmarshal(raw, &parsed); err != nil {
		return "", "", fmt.Errorf("HTTP %d invalid nav json: %w", resp.StatusCode, err)
	}
	imgURL := strings.TrimSpace(parsed.Data.WbiImg.ImgURL)
	subURL := strings.TrimSpace(parsed.Data.WbiImg.SubURL)
	if imgURL == "" || subURL == "" {
		return "", "", fmt.Errorf("missing wbi_img keys in nav response (HTTP %d, code %d: %s)", resp.StatusCode, parsed.Code, parsed.Message)
	}
	imgKey = strings.TrimSuffix(path.Base(imgURL), path.Ext(imgURL))
	subKey = strings.TrimSuffix(path.Base(subURL), path.Ext(subURL))
	if imgKey == "" || imgKey == "." || subKey == "" || subKey == "." {
		return "", "", fmt.Errorf("invalid wbi_img urls (%q, %q)", imgURL, subURL)
	}
	return imgKey, subKey, nil
}

func signWBI(params url.Values, imgKey, subKey string) url.Values {
	rawKey := imgKey + subKey
	var mixin strings.Builder
	for _, idx := range wbiMixinKeyEncTab {
		if idx < len(rawKey) {
			mixin.WriteByte(rawKey[idx])
		}
	}
	mixinKey := mixin.String()
	if len(mixinKey) > 32 {
		mixinKey = mixinKey[:32]
	}

	params.Set("wts", strconv.FormatInt(time.Now().Unix(), 10))
	keys := make([]string, 0, len(params))
	for k := range params {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	ordered := url.Values{}
	for _, k := range keys {
		val := strings.Map(func(r rune) rune {
			if strings.ContainsRune("!'()*", r) {
				return -1
			}
			return r
		}, params.Get(k))
		ordered.Set(k, val)
	}
	query := ordered.Encode()
	sum := md5.Sum([]byte(query + mixinKey))
	ordered.Set("w_rid", hex.EncodeToString(sum[:]))
	return ordered
}

func (b *BilibiliScraper) fetchUpdates(ctx context.Context, client *http.Client, uids []string) ([]module.ScrapedItem, error) {
	if len(uids) == 0 {
		return nil, nil
	}
	imgKey, subKey, wbiErr := b.fetchWBIKeys(ctx, client)

	var all []module.ScrapedItem
	var lastErr error
	attempted := 0

	for _, rawUID := range uids {
		uid := strings.TrimSpace(rawUID)
		if uid == "" {
			continue
		}
		attempted++
		items, err := b.fetchSingleUploader(ctx, client, uid, imgKey, subKey)
		if err != nil {
			lastErr = err
		} else {
			all = append(all, items...)
		}
		time.Sleep(600 * time.Millisecond)
	}
	if len(all) == 0 && attempted > 0 && lastErr != nil {
		if wbiErr != nil {
			return nil, fmt.Errorf("wbi: %v; uploader: %w", wbiErr, lastErr)
		}
		return nil, lastErr
	}
	return all, nil
}

func (b *BilibiliScraper) fetchSingleUploader(ctx context.Context, client *http.Client, uid, imgKey, subKey string) ([]module.ScrapedItem, error) {
	var arcErr error
	if imgKey != "" && subKey != "" {
		items, err := b.fetchUploaderWBI(ctx, client, uid, imgKey, subKey)
		if err == nil {
			return items, nil
		}
		arcErr = err
	} else {
		// Ensure keys attempted once when invoked directly from TestConfig
		ik, sk, kErr := b.fetchWBIKeys(ctx, client)
		if kErr == nil && ik != "" && sk != "" {
			items, err := b.fetchUploaderWBI(ctx, client, uid, ik, sk)
			if err == nil {
				return items, nil
			}
			arcErr = err
		} else if kErr != nil {
			arcErr = kErr
		}
	}

	dynItems, dynErr := b.fetchUploaderDynamics(ctx, client, uid)
	if dynErr == nil {
		return dynItems, nil
	}
	if arcErr != nil {
		return nil, fmt.Errorf("arc/search (%v); dynamic (%v)", arcErr, dynErr)
	}
	return nil, dynErr
}

func (b *BilibiliScraper) fetchUploaderWBI(ctx context.Context, client *http.Client, uid, imgKey, subKey string) ([]module.ScrapedItem, error) {
	q := url.Values{}
	q.Set("mid", uid)
	q.Set("ps", "15")
	q.Set("pn", "1")
	q.Set("tid", "0")
	q.Set("keyword", "")
	q.Set("order", "pubdate")
	q.Set("platform", "web")
	q.Set("web_location", "1550101")
	q.Set("order_avoided", "true")
	q.Set("dm_img_list", "[]")
	q.Set("dm_img_str", "V2ViR0wgMS4wIChPcGVuR0wgRVMgMi4wIENocm9taXVtKQ")
	q.Set("dm_cover_img_str", "QU5HTEUgKEdvb2dsZSwgVnVsa2FuIDEuMy4wIChTd2lmdFNoYWRlciBEZXZpY2UgKFN1Ynplcm8pICgweDAwMDBDMERFKSksIFN3aWZ0U2hhZGVyIGRyaXZlcik")
	q.Set("dm_img_inter", `{"ds":[],"wh":[3557,1574,104],"of":[153,306,153]}`)
	signed := signWBI(q, imgKey, subKey)

	referer := fmt.Sprintf("https://space.bilibili.com/%s/video", uid)
	req, err := b.newRequest(ctx, "https://api.bilibili.com/x/space/wbi/arc/search?"+signed.Encode(), referer)
	if err != nil {
		return nil, err
	}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()

	raw, err := io.ReadAll(io.LimitReader(resp.Body, 4<<20))
	if err != nil {
		return nil, err
	}
	var parsed struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
		Data    struct {
			List struct {
				Vlist []struct {
					Bvid        string `json:"bvid"`
					Title       string `json:"title"`
					Description string `json:"description"`
					Author      string `json:"author"`
					Pic         string `json:"pic"`
					Created     int64  `json:"created"`
				} `json:"vlist"`
			} `json:"list"`
		} `json:"data"`
	}
	if err := json.Unmarshal(raw, &parsed); err != nil {
		return nil, fmt.Errorf("HTTP %d invalid json: %w", resp.StatusCode, err)
	}
	if parsed.Code != 0 {
		return nil, fmt.Errorf("HTTP %d code %d: %s", resp.StatusCode, parsed.Code, parsed.Message)
	}

	out := make([]module.ScrapedItem, 0, len(parsed.Data.List.Vlist))
	for _, v := range parsed.Data.List.Vlist {
		out = append(out, buildBiliVideoItem(v.Bvid, v.Title, v.Description, v.Author, v.Pic, time.Unix(v.Created, 0).UTC()))
	}
	return out, nil
}

func (b *BilibiliScraper) fetchUploaderDynamics(ctx context.Context, client *http.Client, uid string) ([]module.ScrapedItem, error) {
	q := url.Values{}
	q.Set("host_mid", uid)
	q.Set("offset", "")
	q.Set("timezone_offset", "-480")
	q.Set("platform", "web")
	q.Set("web_location", "333.1387")
	q.Set("features", "itemOpusStyle,listOnlyfans,opusBigCover,onlyfansVote")

	referer := fmt.Sprintf("https://space.bilibili.com/%s/dynamic", uid)
	req, err := b.newRequest(ctx, "https://api.bilibili.com/x/polymer/web-dynamic/v1/feed/space?"+q.Encode(), referer)
	if err != nil {
		return nil, err
	}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()

	raw, err := io.ReadAll(io.LimitReader(resp.Body, 4<<20))
	if err != nil {
		return nil, err
	}
	var parsed struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
		Data    struct {
			Items []struct {
				Type    string `json:"type"`
				Modules struct {
					ModuleAuthor struct {
						Name  string `json:"name"`
						PubTs int64  `json:"pub_ts"`
					} `json:"module_author"`
					ModuleDynamic struct {
						Major *struct {
							Type    string `json:"type"`
							Archive *struct {
								Bvid  string `json:"bvid"`
								Title string `json:"title"`
								Desc  string `json:"desc"`
								Cover string `json:"cover"`
							} `json:"archive"`
						} `json:"major"`
					} `json:"module_dynamic"`
				} `json:"modules"`
			} `json:"items"`
		} `json:"data"`
	}
	if err := json.Unmarshal(raw, &parsed); err != nil {
		return nil, fmt.Errorf("HTTP %d invalid json: %w", resp.StatusCode, err)
	}
	if parsed.Code != 0 {
		return nil, fmt.Errorf("HTTP %d code %d: %s", resp.StatusCode, parsed.Code, parsed.Message)
	}

	var out []module.ScrapedItem
	for _, it := range parsed.Data.Items {
		if it.Type != "DYNAMIC_TYPE_AV" || it.Modules.ModuleDynamic.Major == nil || it.Modules.ModuleDynamic.Major.Archive == nil {
			continue
		}
		arc := it.Modules.ModuleDynamic.Major.Archive
		if arc.Bvid == "" {
			continue
		}
		pubTime := time.Unix(it.Modules.ModuleAuthor.PubTs, 0).UTC()
		if it.Modules.ModuleAuthor.PubTs <= 0 {
			pubTime = time.Now().UTC()
		}
		author := strings.TrimSpace(it.Modules.ModuleAuthor.Name)
		if author == "" {
			author = "Bilibili UP " + uid
		}
		out = append(out, buildBiliVideoItem(arc.Bvid, arc.Title, arc.Desc, author, arc.Cover, pubTime))
	}
	return out, nil
}

func buildBiliVideoItem(bvid, title, desc, author, pic string, pubTime time.Time) module.ScrapedItem {
	videoURL := "https://www.bilibili.com/video/" + bvid
	bodyHTML := fmt.Sprintf(
		`<div class="bilibili-video"><div style="position:relative;padding-bottom:56.25%%;height:0;overflow:hidden;border-radius:8px;margin-bottom:10px;"><iframe src="https://player.bilibili.com/player.html?bvid=%s&autoplay=0" scrolling="no" frameborder="0" allowfullscreen="true" style="position:absolute;top:0;left:0;width:100%%;height:100%%;"></iframe></div><p>%s</p></div>`,
		html.EscapeString(bvid), html.EscapeString(desc),
	)
	sum := sha256.Sum256([]byte("bilibili_update|" + bvid))
	var media []module.ScrapedMedia
	if pic != "" {
		media = append(media, module.ScrapedMedia{Type: "image", URL: normalizeBiliMediaURL(pic)})
	}
	return module.ScrapedItem{
		Title:       title,
		Summary:     buildSummary(desc),
		Content:     bodyHTML,
		ContentType: "html",
		RenderType:  "html",
		SourceName:  author,
		Author:      author,
		ExternalURL: videoURL,
		PublishedAt: pubTime,
		Fingerprint: hex.EncodeToString(sum[:]),
		Media:       media,
	}
}

func normalizeBiliMediaURL(raw string) string {
	s := strings.TrimSpace(raw)
	if strings.HasPrefix(s, "//") {
		return "https:" + s
	}
	if strings.HasPrefix(s, "http://") {
		return "https://" + strings.TrimPrefix(s, "http://")
	}
	return s
}

type BilibiliQRSession struct {
	URL       string `json:"url"`
	QRCodeKey string `json:"qrcode_key"`
}

type BilibiliQRPollResult struct {
	Status  string `json:"status"` // waiting | scanned | confirmed | expired
	Code    int    `json:"code"`
	Message string `json:"message"`
	Cookie  string `json:"cookie,omitempty"`
}

func GenerateBilibiliQR(ctx context.Context) (*BilibiliQRSession, error) {
	client := &http.Client{Timeout: 15 * time.Second}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "https://passport.bilibili.com/x/passport-login/web/qrcode/generate", nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", biliDesktopUA)
	req.Header.Set("Referer", "https://passport.bilibili.com/login")

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()

	raw, _ := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	var parsed struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
		Data    struct {
			URL       string `json:"url"`
			QRCodeKey string `json:"qrcode_key"`
		} `json:"data"`
	}
	if err := json.Unmarshal(raw, &parsed); err != nil {
		return nil, err
	}
	if parsed.Code != 0 || parsed.Data.URL == "" || parsed.Data.QRCodeKey == "" {
		return nil, fmt.Errorf("qrcode generate failed (%d): %s", parsed.Code, parsed.Message)
	}
	return &BilibiliQRSession{
		URL:       parsed.Data.URL,
		QRCodeKey: parsed.Data.QRCodeKey,
	}, nil
}

func PollBilibiliQR(ctx context.Context, qrcodeKey string) (*BilibiliQRPollResult, error) {
	qrcodeKey = strings.TrimSpace(qrcodeKey)
	if qrcodeKey == "" {
		return nil, errors.New("qrcode_key is required")
	}

	client := &http.Client{Timeout: 15 * time.Second}
	endpoint := "https://passport.bilibili.com/x/passport-login/web/qrcode/poll?qrcode_key=" + url.QueryEscape(qrcodeKey)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", biliDesktopUA)
	req.Header.Set("Referer", "https://passport.bilibili.com/login")

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()

	raw, _ := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	var parsed struct {
		Code int `json:"code"`
		Data struct {
			URL          string `json:"url"`
			RefreshToken string `json:"refresh_token"`
			Timestamp    int64  `json:"timestamp"`
			Code         int    `json:"code"`
			Message      string `json:"message"`
		} `json:"data"`
	}
	if err := json.Unmarshal(raw, &parsed); err != nil {
		return nil, err
	}

	res := &BilibiliQRPollResult{
		Code:    parsed.Data.Code,
		Message: parsed.Data.Message,
	}
	switch parsed.Data.Code {
	case 0:
		res.Status = "confirmed"
		cookieMap := make(map[string]string)
		for _, ck := range resp.Cookies() {
			if ck.Name != "" && ck.Value != "" {
				cookieMap[ck.Name] = ck.Value
			}
		}
		if parsed.Data.URL != "" {
			if u, err := url.Parse(parsed.Data.URL); err == nil {
				for _, key := range []string{"DedeUserID", "DedeUserID__ckMd5", "SESSDATA", "bili_jct", "sid"} {
					if v := u.Query().Get(key); v != "" {
						cookieMap[key] = v
					}
				}
			}
		}
		if len(cookieMap) == 0 {
			return nil, errors.New("login confirmed without cookies")
		}
		keys := make([]string, 0, len(cookieMap))
		for k := range cookieMap {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		parts := make([]string, 0, len(keys))
		for _, k := range keys {
			parts = append(parts, fmt.Sprintf("%s=%s", k, cookieMap[k]))
		}
		res.Cookie = strings.Join(parts, "; ")
	case 86101:
		res.Status = "waiting"
	case 86090:
		res.Status = "scanned"
	case 86038:
		res.Status = "expired"
	default:
		res.Status = "waiting"
	}
	return res, nil
}
