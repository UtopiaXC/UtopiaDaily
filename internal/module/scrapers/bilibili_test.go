package scrapers

import (
	"context"
	"io"
	"net/http"
	"net/url"
	"strings"
	"testing"
)

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}

func jsonResponse(status int, body string, cookies ...string) *http.Response {
	h := make(http.Header)
	h.Set("Content-Type", "application/json")
	for _, c := range cookies {
		h.Add("Set-Cookie", c)
	}
	return &http.Response{
		StatusCode: status,
		Header:     h,
		Body:       io.NopCloser(strings.NewReader(body)),
	}
}

func TestParseCookiePairsFormats(t *testing.T) {
	cases := []struct {
		name string
		raw  string
		want map[string]string
		omit []string
	}{
		{
			name: "standard header with Cookie prefix and CRLF",
			raw:  "Cookie: SESSDATA=abc%2C123*xyz; bili_jct=csrf99; DedeUserID=42; Path=/; HttpOnly\r\n",
			want: map[string]string{
				"SESSDATA":   "abc%2C123*xyz",
				"bili_jct":   "csrf99",
				"DedeUserID": "42",
			},
			omit: []string{"Path", "HttpOnly"},
		},
		{
			name: "quoted full header and quoted values",
			raw:  `'cookie: SESSDATA="raw,comma+val"; bili_jct=csrf_val'`,
			want: map[string]string{
				"SESSDATA": "raw,comma+val",
				"bili_jct": "csrf_val",
			},
		},
		{
			name: "devtools application cookies table tab separated",
			raw:  "Name\tValue\tDomain\tPath\nSESSDATA\ttok%2C1\t.bilibili.com\t/\nbili_jct\tjct1\t.bilibili.com\t/\nDedeUserID\t10001\t.bilibili.com\t/",
			want: map[string]string{
				"SESSDATA":   "tok%2C1",
				"bili_jct":   "jct1",
				"DedeUserID": "10001",
			},
			omit: []string{"Name"},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			pairs := parseCookiePairs(tc.raw)
			got := make(map[string]string, len(pairs))
			for _, p := range pairs {
				got[p.Name] = p.Value
			}
			for k, wantVal := range tc.want {
				if got[k] != wantVal {
					t.Fatalf("key %s = %q, want %q", k, got[k], wantVal)
				}
			}
			for _, k := range tc.omit {
				if _, ok := got[k]; ok {
					t.Fatalf("unexpected cookie attribute parsed: %s", k)
				}
			}
		})
	}
}

func TestNewSessionClientRefreshesBrowserFingerprintsWhenSESSDATAPresent(t *testing.T) {
	scraper := NewBilibiliScraper()
	var seenNavCookies map[string]string
	var spiHit, ticketHit bool

	rt := roundTripFunc(func(req *http.Request) (*http.Response, error) {
		u := req.URL.String()
		switch {
		case u == "https://www.bilibili.com/":
			return jsonResponse(http.StatusOK, "ok", "b_nut=fresh_bnut; Domain=.bilibili.com; Path=/"), nil
		case strings.HasPrefix(u, "https://api.bilibili.com/x/frontend/finger/spi"):
			spiHit = true
			return jsonResponse(http.StatusOK, `{"code":0,"data":{"b_3":"fresh_b3","b_4":"fresh_b4"}}`), nil
		case strings.HasPrefix(u, "https://api.bilibili.com/bapis/bilibili.api.ticket.v1.Ticket/GenWebTicket"):
			ticketHit = true
			if req.URL.Query().Get("csrf") != "csrf_tok" {
				t.Fatalf("GenWebTicket csrf = %q, want csrf_tok", req.URL.Query().Get("csrf"))
			}
			return jsonResponse(http.StatusOK, `{"code":0,"data":{"ticket":"fresh_ticket"}}`), nil
		case strings.HasPrefix(u, "https://api.bilibili.com/x/web-interface/nav"):
			seenNavCookies = make(map[string]string)
			for _, ck := range req.Cookies() {
				seenNavCookies[ck.Name] = ck.Value
			}
			return jsonResponse(http.StatusOK, `{"code":0,"data":{"wbi_img":{"img_url":"https://i0.hdslb.com/bfs/wbi/7cd084941338484aae1ad9425b84077c.png","sub_url":"https://i0.hdslb.com/bfs/wbi/4932caff0ff746eab6f01bf08b70ac45.png"}}}`), nil
		default:
			return jsonResponse(http.StatusNotFound, `{}`), nil
		}
	})

	pastedFullCookie := "Cookie: DedeUserID=999; SESSDATA=sess%2C17000*abc; bili_jct=csrf_tok; buvid3=stale_browser_b3; buvid_fp=stale_fp; bili_ticket=stale_ticket; b_nut=stale_bnut"
	client := scraper.newSessionClientWithTransport(context.Background(), pastedFullCookie, rt)

	imgKey, subKey, err := scraper.fetchWBIKeys(context.Background(), client)
	if err != nil {
		t.Fatalf("fetchWBIKeys failed: %v", err)
	}
	if imgKey == "" || subKey == "" {
		t.Fatalf("expected non-empty WBI keys, got (%q, %q)", imgKey, subKey)
	}
	if !spiHit || !ticketHit {
		t.Fatalf("expected spi (%v) and ticket (%v) refresh endpoints to be called even when pasted cookie contained stale tokens", spiHit, ticketHit)
	}
	if seenNavCookies["SESSDATA"] != "sess%2C17000*abc" {
		t.Fatalf("nav SESSDATA = %q, want intact value", seenNavCookies["SESSDATA"])
	}
	if seenNavCookies["buvid3"] != "fresh_b3" {
		t.Fatalf("nav buvid3 = %q, want fresh_b3", seenNavCookies["buvid3"])
	}
	if seenNavCookies["bili_ticket"] != "fresh_ticket" {
		t.Fatalf("nav bili_ticket = %q, want fresh_ticket", seenNavCookies["bili_ticket"])
	}
	if seenNavCookies["b_nut"] != "fresh_bnut" {
		t.Fatalf("nav b_nut = %q, want fresh_bnut", seenNavCookies["b_nut"])
	}
	if _, leaked := seenNavCookies["buvid_fp"]; leaked {
		t.Fatalf("stale buvid_fp should have been stripped")
	}
}

func TestFetchWBIKeysRejectsEmptyImageURLs(t *testing.T) {
	scraper := NewBilibiliScraper()
	client := &http.Client{
		Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
			return jsonResponse(http.StatusPreconditionFailed, `{"code":-352,"message":"risk control"}`), nil
		}),
	}
	rootURL, _ := url.Parse("https://www.bilibili.com/")
	_ = rootURL

	_, _, err := scraper.fetchWBIKeys(context.Background(), client)
	if err == nil {
		t.Fatal("expected error when wbi_img URLs are empty")
	}
}
