package service

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestClassificationServiceJobDedupeAndLifecycle(t *testing.T) {
	proceed := make(chan struct{})
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fl, _ := w.(http.Flusher)
		_, _ = w.Write([]byte("{\"status\":\"pulling aaaa\",\"digest\":\"sha256:aaaa\",\"total\":1000,\"completed\":250}\n"))
		if fl != nil {
			fl.Flush()
		}
		select {
		case <-proceed:
		case <-r.Context().Done():
			return
		}
		_, _ = w.Write([]byte("{\"status\":\"pulling aaaa\",\"digest\":\"sha256:aaaa\",\"total\":1000,\"completed\":1000}\n"))
		_, _ = w.Write([]byte("{\"status\":\"success\"}\n"))
		if fl != nil {
			fl.Flush()
		}
	}))
	defer srv.Close()

	svc := NewClassificationService(nil, nil, nil, nil, nil)
	defer svc.Stop()

	job1, err := svc.StartPullOllamaModelJob(srv.URL, "qwen3.5:0.8b")
	if err != nil {
		t.Fatalf("StartPullOllamaModelJob failed: %v", err)
	}
	if job1.Status != ClassJobRunning {
		t.Fatalf("expected initial status running, got %s", job1.Status)
	}

	// Second call while active must deduplicate to the same job ID
	jobDup, err := svc.StartPullOllamaModelJob(srv.URL, "qwen3.5:0.8b")
	if err != nil {
		t.Fatalf("duplicate StartPullOllamaModelJob failed: %v", err)
	}
	if jobDup.ID != job1.ID {
		t.Fatalf("expected deduped job ID %s, got %s", job1.ID, jobDup.ID)
	}

	// Wait until streaming progress has been recorded
	deadline := time.Now().Add(3 * time.Second)
	var observedProgress bool
	for time.Now().Before(deadline) {
		cur, getErr := svc.GetJob(job1.ID)
		if getErr != nil {
			t.Fatalf("GetJob failed: %v", getErr)
		}
		if cur.TotalBytes == 1000 && cur.CompletedBytes == 250 && !cur.Indeterminate {
			observedProgress = true
			break
		}
		time.Sleep(15 * time.Millisecond)
	}
	if !observedProgress {
		cur, _ := svc.GetJob(job1.ID)
		t.Fatalf("did not observe intermediate progress, got %+v", cur)
	}

	// Dismiss while running must fail
	if err := svc.DismissJob(job1.ID); err == nil {
		t.Fatalf("expected error dismissing running job")
	}

	close(proceed)

	deadline = time.Now().Add(3 * time.Second)
	var completedJob *ClassificationJob
	for time.Now().Before(deadline) {
		cur, _ := svc.GetJob(job1.ID)
		if cur != nil && cur.Status == ClassJobCompleted {
			completedJob = cur
			break
		}
		time.Sleep(15 * time.Millisecond)
	}
	if completedJob == nil {
		cur, _ := svc.GetJob(job1.ID)
		t.Fatalf("job did not reach completed status: %+v", cur)
	}
	if completedJob.Percent != 100 || completedJob.Stage != "ready" {
		t.Fatalf("unexpected completed state: %+v", completedJob)
	}

	if err := svc.DismissJob(job1.ID); err != nil {
		t.Fatalf("DismissJob failed: %v", err)
	}
	if list := svc.ListJobs(); len(list) != 0 {
		t.Fatalf("expected 0 jobs after dismiss, got %d", len(list))
	}
}

func TestClassificationServiceCancelAndStop(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fl, _ := w.(http.Flusher)
		_, _ = w.Write([]byte("{\"status\":\"pulling manifest\"}\n"))
		if fl != nil {
			fl.Flush()
		}
		<-r.Context().Done()
	}))
	defer srv.Close()

	svc := NewClassificationService(nil, nil, nil, nil, nil)
	job, err := svc.StartPullOllamaModelJob(srv.URL, "blocking:model")
	if err != nil {
		t.Fatalf("StartPullOllamaModelJob failed: %v", err)
	}

	cancelled, err := svc.CancelJob(job.ID)
	if err != nil {
		t.Fatalf("CancelJob failed: %v", err)
	}
	if cancelled.Status != ClassJobCancelled {
		t.Fatalf("expected cancelled status, got %s", cancelled.Status)
	}

	// Start another blocking job and verify Stop() terminates and waits cleanly
	job2, err := svc.StartPullOllamaModelJob(srv.URL, "another:model")
	if err != nil {
		t.Fatalf("second StartPullOllamaModelJob failed: %v", err)
	}
	svc.Stop()

	afterStop, err := svc.GetJob(job2.ID)
	if err != nil {
		t.Fatalf("GetJob after Stop failed: %v", err)
	}
	if afterStop.Status != ClassJobCancelled {
		t.Fatalf("expected job2 cancelled on Stop(), got %s", afterStop.Status)
	}
}
