package classification

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"
)

func TestPullOllamaModelStreamingProgress(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/pull" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		var reqBody map[string]any
		if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
			t.Fatalf("decode req: %v", err)
		}
		if reqBody["name"] != "qwen3.5:0.8b" {
			t.Fatalf("unexpected model name: %v", reqBody["name"])
		}
		if reqBody["stream"] != true {
			t.Fatalf("expected stream=true, got %v", reqBody["stream"])
		}

		fl, _ := w.(http.Flusher)
		lines := []string{
			`{"status":"pulling manifest"}`,
			`{"status":"pulling aaaaaaaaaaaa","digest":"sha256:aaaaaaaaaaaa1111","total":1000,"completed":500}`,
			`{"status":"pulling bbbbbbbbbbbb","digest":"sha256:bbbbbbbbbbbb2222","total":1000,"completed":0}`,
			`{"status":"pulling aaaaaaaaaaaa","digest":"sha256:aaaaaaaaaaaa1111","total":1000,"completed":1000}`,
			`{"status":"pulling bbbbbbbbbbbb","digest":"sha256:bbbbbbbbbbbb2222","total":1000,"completed":1000}`,
			`{"status":"verifying sha256 digest"}`,
			`{"status":"writing manifest"}`,
			`{"status":"success"}`,
		}
		for _, ln := range lines {
			_, _ = w.Write([]byte(ln + "\n"))
			if fl != nil {
				fl.Flush()
			}
			time.Sleep(15 * time.Millisecond)
		}
	}))
	defer srv.Close()

	cls := NewLLMClassifier()
	var mu sync.Mutex
	var updates []ProgressUpdate
	err := cls.PullOllamaModel(context.Background(), srv.URL, "qwen3.5:0.8b", func(u ProgressUpdate) {
		mu.Lock()
		updates = append(updates, u)
		mu.Unlock()
	})
	if err != nil {
		t.Fatalf("PullOllamaModel failed: %v", err)
	}

	mu.Lock()
	defer mu.Unlock()
	if len(updates) < 3 {
		t.Fatalf("expected multiple progress updates, got %d", len(updates))
	}
	var sawDownloading, sawVerifying, sawFinalizing, sawReady bool
	for _, u := range updates {
		switch u.Stage {
		case "downloading":
			sawDownloading = true
			if u.TotalBytes > 0 && (u.Percent < 0 || u.Percent > 99) {
				t.Fatalf("unexpected downloading percent %v", u.Percent)
			}
		case "verifying":
			sawVerifying = true
		case "finalizing":
			sawFinalizing = true
		case "ready":
			sawReady = true
			if u.Percent != 100 || u.CompletedBytes != 2000 || u.TotalBytes != 2000 {
				t.Fatalf("unexpected ready update: %+v", u)
			}
		}
	}
	if !sawDownloading || !sawVerifying || !sawFinalizing || !sawReady {
		t.Fatalf("missing stages: downloading=%v verifying=%v finalizing=%v ready=%v", sawDownloading, sawVerifying, sawFinalizing, sawReady)
	}
}

func TestPullOllamaModelStreamErrorAndCancel(t *testing.T) {
	errSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("{\"status\":\"pulling manifest\"}\n"))
		_, _ = w.Write([]byte("{\"error\":\"pull model manifest: file does not exist\"}\n"))
	}))
	defer errSrv.Close()

	cls := NewLLMClassifier()
	err := cls.PullOllamaModel(context.Background(), errSrv.URL, "missing:tag", nil)
	if err == nil || !strings.Contains(err.Error(), "file does not exist") {
		t.Fatalf("expected manifest error, got %v", err)
	}

	blockSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fl, _ := w.(http.Flusher)
		_, _ = w.Write([]byte("{\"status\":\"pulling manifest\"}\n"))
		if fl != nil {
			fl.Flush()
		}
		<-r.Context().Done()
	}))
	defer blockSrv.Close()

	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		time.Sleep(40 * time.Millisecond)
		cancel()
	}()
	err = cls.PullOllamaModel(ctx, blockSrv.URL, "slow:model", nil)
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("expected context.Canceled, got %v", err)
	}
}

func TestWorkerProgressStderrWriter(t *testing.T) {
	var got []ProgressUpdate
	w := &workerProgressStderrWriter{
		onProgress: func(u ProgressUpdate) {
			got = append(got, u)
		},
	}

	chunk := "initializing\nPROGRESS_JSON:{\"stage\":\"resolving\",\"current_item\":\"repo/m\",\"indeterminate\":true,\"percent\":-1}\rPROGRESS_JSON:{\"stage\":\"downloading\",\"completed_bytes\":512,\"total_bytes\":1024,\"percent\":47.5,\"indeterminate\":false}\nreal warning line\nPROGRESS_JSON:{\"stage\":\"ready\",\"percent\":100}\ntrailing stderr"
	if _, err := w.Write([]byte(chunk)); err != nil {
		t.Fatalf("write failed: %v", err)
	}

	if len(got) != 3 {
		t.Fatalf("expected 3 progress updates, got %d (%+v)", len(got), got)
	}
	if got[0].Stage != "resolving" || !got[0].Indeterminate {
		t.Fatalf("unexpected first update: %+v", got[0])
	}
	if got[1].Stage != "downloading" || got[1].CompletedBytes != 512 || got[1].TotalBytes != 1024 {
		t.Fatalf("unexpected second update: %+v", got[1])
	}
	if got[2].Stage != "ready" || got[2].Percent != 100 {
		t.Fatalf("unexpected third update: %+v", got[2])
	}

	errOut := w.ErrorOutput()
	if strings.Contains(errOut, "PROGRESS_JSON:") {
		t.Fatalf("error output leaked PROGRESS_JSON: %s", errOut)
	}
	if !strings.Contains(errOut, "initializing") || !strings.Contains(errOut, "real warning line") || !strings.Contains(errOut, "trailing stderr") {
		t.Fatalf("error output missing non-progress lines: %s", errOut)
	}
}
