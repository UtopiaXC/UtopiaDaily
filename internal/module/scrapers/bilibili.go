package scrapers

import (
	"context"
	"crypto/md5"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"html"
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

var wbiMixinKeyEncTab = []int{
	46, 47, 18, 2, 53, 8, 23, 32, 15, 50, 10, 31, 58, 3, 45, 35, 27, 43, 5, 49,
	33, 9, 42, 19, 29, 28, 14, 39, 12, 38, 41, 13, 37, 48, 7, 16, 24, 55, 40,
	61, 26, 17, 0, 1, 60, 51, 30, 4, 22, 25, 54, 21, 56, 59, 6, 63, 57, 62, 11,
	36, 20, 34, 44, 52,
}

type BilibiliScraper struct {
	client *http.Client
}

func NewBilibiliScraper() *BilibiliScraper {
	jar, _ := cookiejar.New(nil)
	return &BilibiliScraper{
		client: &http.Client{
			Timeout: 20 * time.Second,
			Jar:     jar,
		},
	}
}

func (b *BilibiliScraper) Meta() module.Metadata {
	return module.Metadata{
		ID:      "bilibili_ranking_and_update",
		Kind:    module.KindScraper,
		Source:  module.SourceBuiltIn,
		Version: "2.0.0",
		Author:  "UtopiaXC",
	}
}

func (b *BilibiliScraper) ConfigSchema() []module.ConfigFieldSpec {
	return []module.ConfigFieldSpec{
		{Key: "uploader_id", Type: "array", DefaultItems: []string{}, SortOrder: 1},
		{Key: "ranking_limit", Type: "int", DefaultValue: "20", SortOrder: 2},
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
			"module.bilibili_ranking_and_update.task.fetch_ranking.name":   "Fetch Ranking",
			"module.bilibili_ranking_and_update.task.fetch_ranking.desc":   "Collect Bilibili overall hot video ranking",
			"module.bilibili_ranking_and_update.task.fetch_update.name":    "Fetch Updates",
			"module.bilibili_ranking_and_update.task.fetch_update.desc":    "Collect latest videos from configured uploaders",
		},
	}
}

func (b *BilibiliScraper) Scrape(ctx context.Context, taskKey string, cfg module.ConfigAccessor) ([]module.ScrapedItem, error) {
	b.warmSession(ctx)
	switch taskKey {
	case "fetch_ranking":
		limit := cfg.GetInt("ranking_limit", 20)
		if limit <= 0 {
			limit = 20
		}
		return b.fetchRanking(ctx, limit)
	case "fetch_update":
		uids := cfg.GetStringSlice("uploader_id")
		return b.fetchUpdates(ctx, uids)
	default:
		return nil, fmt.Errorf("unknown task: %s", taskKey)
	}
}

func (b *BilibiliScraper) TestConfig(ctx context.Context, cfg module.ConfigAccessor) module.TestResult {
	b.warmSession(ctx)
	req, err := b.newRequest(ctx, "https://api.bilibili.com/x/web-interface/ranking/v2?rid=0&type=all")
	if err != nil {
		return module.TestResult{Success: false, MessageKey: "module.test.failed", Detail: err.Error()}
	}
	resp, err := b.client.Do(req)
	if err != nil {
		return module.TestResult{Success: false, MessageKey: "module.test.failed", Detail: err.Error()}
	}
	defer func() { _ = resp.Body.Close() }()
	var payload struct {
		Code int `json:"code"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil || payload.Code != 0 {
		return module.TestResult{Success: false, MessageKey: "module.test.failed", Detail: fmt.Sprintf("HTTP %d code %d", resp.StatusCode, payload.Code)}
	}
	return module.TestResult{Success: true, MessageKey: "module.test.passed"}
}

func (b *BilibiliScraper) warmSession(ctx context.Context) {
	req, err := b.newRequest(ctx, "https://www.bilibili.com/")
	if err != nil {
		return
	}
	resp, err := b.client.Do(req)
	if err == nil {
		_ = resp.Body.Close()
	}
}

func (b *BilibiliScraper) newRequest(ctx context.Context, targetURL string) (*http.Request, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, targetURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36")
	req.Header.Set("Referer", "https://www.bilibili.com/")
	return req, nil
}

func (b *BilibiliScraper) fetchRanking(ctx context.Context, limit int) ([]module.ScrapedItem, error) {
	req, err := b.newRequest(ctx, "https://api.bilibili.com/x/web-interface/ranking/v2?rid=0&type=all")
	if err != nil {
		return nil, err
	}
	resp, err := b.client.Do(req)
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
		return nil, fmt.Errorf("bilibili ranking api error: %s", parsed.Message)
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
			media = append(media, module.ScrapedMedia{Type: "image", URL: v.Pic})
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

func (b *BilibiliScraper) fetchWBIKeys(ctx context.Context) (imgKey, subKey string, err error) {
	req, err := b.newRequest(ctx, "https://api.bilibili.com/x/web-interface/nav")
	if err != nil {
		return "", "", err
	}
	resp, err := b.client.Do(req)
	if err != nil {
		return "", "", err
	}
	defer func() { _ = resp.Body.Close() }()

	var parsed struct {
		Data struct {
			WbiImg struct {
				ImgURL string `json:"img_url"`
				SubURL string `json:"sub_url"`
			} `json:"wbi_img"`
		} `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&parsed); err != nil {
		return "", "", err
	}
	imgKey = strings.TrimSuffix(path.Base(parsed.Data.WbiImg.ImgURL), path.Ext(parsed.Data.WbiImg.ImgURL))
	subKey = strings.TrimSuffix(path.Base(parsed.Data.WbiImg.SubURL), path.Ext(parsed.Data.WbiImg.SubURL))
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

func (b *BilibiliScraper) fetchUpdates(ctx context.Context, uids []string) ([]module.ScrapedItem, error) {
	if len(uids) == 0 {
		return nil, nil
	}
	imgKey, subKey, err := b.fetchWBIKeys(ctx)
	if err != nil {
		return nil, err
	}

	var all []module.ScrapedItem
	for _, rawUID := range uids {
		uid := strings.TrimSpace(rawUID)
		if uid == "" {
			continue
		}
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
		signed := signWBI(q, imgKey, subKey)

		req, err := b.newRequest(ctx, "https://api.bilibili.com/x/space/wbi/arc/search?"+signed.Encode())
		if err != nil {
			continue
		}
		resp, err := b.client.Do(req)
		if err != nil {
			continue
		}
		var parsed struct {
			Code int `json:"code"`
			Data struct {
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
		_ = json.NewDecoder(resp.Body).Decode(&parsed)
		_ = resp.Body.Close()
		if parsed.Code != 0 {
			continue
		}

		for _, v := range parsed.Data.List.Vlist {
			videoURL := "https://www.bilibili.com/video/" + v.Bvid
			bodyHTML := fmt.Sprintf(
				`<div class="bilibili-video"><div style="position:relative;padding-bottom:56.25%%;height:0;overflow:hidden;border-radius:8px;margin-bottom:10px;"><iframe src="https://player.bilibili.com/player.html?bvid=%s&autoplay=0" scrolling="no" frameborder="0" allowfullscreen="true" style="position:absolute;top:0;left:0;width:100%%;height:100%%;"></iframe></div><p>%s</p></div>`,
				html.EscapeString(v.Bvid), html.EscapeString(v.Description),
			)
			sum := sha256.Sum256([]byte("bilibili_update|" + v.Bvid))
			var media []module.ScrapedMedia
			if v.Pic != "" {
				media = append(media, module.ScrapedMedia{Type: "image", URL: v.Pic})
			}
			all = append(all, module.ScrapedItem{
				Title:       v.Title,
				Summary:     buildSummary(v.Description),
				Content:     bodyHTML,
				ContentType: "html",
				RenderType:  "html",
				SourceName:  v.Author,
				Author:      v.Author,
				ExternalURL: videoURL,
				PublishedAt: time.Unix(v.Created, 0).UTC(),
				Fingerprint: hex.EncodeToString(sum[:]),
				Media:       media,
			})
		}
		time.Sleep(500 * time.Millisecond)
	}
	return all, nil
}
