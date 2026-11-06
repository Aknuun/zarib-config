package web

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func floatPtr(value float64) *float64 {
	return &value
}

func journalServer(t *testing.T) *server {
	t.Helper()
	root := t.TempDir()
	return &server{
		jobsDir:    filepath.Join(root, "jobs"),
		jobsFile:   filepath.Join(root, "jobs.json"),
		blockedIPs: map[string]bool{},
		jobs:       map[string]*job{},
		rates:      map[string]map[string][]time.Time{},
		sem:        make(chan struct{}, maxActive),
	}
}

func writeJournal(t *testing.T, s *server, journal jobJournal) {
	t.Helper()
	data, err := json.Marshal(journal)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(s.jobsFile, data, 0o600); err != nil {
		t.Fatal(err)
	}
}

func readJournal(t *testing.T, s *server) jobJournal {
	t.Helper()
	data, err := os.ReadFile(s.jobsFile)
	if err != nil {
		t.Fatal(err)
	}
	var journal jobJournal
	if err := json.Unmarshal(data, &journal); err != nil {
		t.Fatal(err)
	}
	return journal
}

func TestSaveJobsPersistsOnlyPending(t *testing.T) {
	s := journalServer(t)
	created := 100.0
	s.mu.Lock()
	s.jobs["queued"] = &job{
		ID: "queued", Status: "queued", Created: created, Mode: "single", Pick: 2,
		SizeMB: 100, Core: "auto", IP: "1.2.3.4", SubKey: "key", SubURL: "https://example.com/sub",
		Wait: 5, Drift: 3, Claimed: floatPtr(1.5), Sync: true,
	}
	s.jobs["running"] = &job{ID: "running", Status: "running", Created: created, Mode: "batch", Picks: []int{1, 2}}
	s.jobs["done"] = &job{ID: "done", Status: "done", Created: created}
	s.saveJobsLocked()
	s.mu.Unlock()

	info, err := os.Stat(s.jobsFile)
	if err != nil {
		t.Fatal(err)
	}
	if perm := info.Mode().Perm(); perm != 0o600 {
		t.Fatalf("journal mode = %o, want 600", perm)
	}
	journal := readJournal(t, s)
	if journal.Version != 1 || len(journal.Jobs) != 2 {
		t.Fatalf("unexpected journal: %+v", journal)
	}
	var queued *jobRecord
	for index := range journal.Jobs {
		if journal.Jobs[index].ID == "queued" {
			queued = &journal.Jobs[index]
		}
	}
	if queued == nil || queued.Wait != 5 || queued.Drift != 3 || !queued.Sync ||
		queued.Claimed == nil || *queued.Claimed != 1.5 || queued.SubURL != "https://example.com/sub" {
		t.Fatalf("queued spec not preserved: %+v", queued)
	}
}

func TestLoadJobsRequeuesWithinLimits(t *testing.T) {
	s := journalServer(t)
	old := startJob
	started := make(chan string, 10)
	startJob = func(_ *server, id string) { started <- id }
	defer func() { startJob = old }()

	id := "abc123"
	dir := filepath.Join(s.jobsDir, id)
	if err := os.MkdirAll(dir, 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "sub.txt"), []byte("https://example.com/sub\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	writeJournal(t, s, jobJournal{Version: 1, Jobs: []jobRecord{
		{ID: id, Status: "queued", Created: 100, Mode: "single", Pick: 1, SizeMB: 100,
			Core: "auto", IP: "1.2.3.4", SubKey: "key", SubURL: "https://example.com/sub",
			Wait: 4, Drift: 2, Claimed: floatPtr(1.2), Sync: true},
		{ID: "duplicate", Status: "queued", Mode: "single", IP: "1.2.3.4", SubKey: "key"},
	}})

	s.loadJobs()

	item := s.jobs[id]
	if item == nil || item.Status != "queued" || item.Wait != 4 || item.Drift != 2 ||
		item.Claimed == nil || *item.Claimed != 1.2 || !item.Sync {
		t.Fatalf("queued job not restored: %+v", item)
	}
	if item.SubFile != filepath.Join(dir, "sub.txt") {
		t.Fatalf("unexpected sub file: %s", item.SubFile)
	}
	select {
	case got := <-started:
		if got != id {
			t.Fatalf("unexpected started job: %s", got)
		}
	case <-time.After(time.Second):
		t.Fatal("queued job was not restarted")
	}
	if len(started) != 0 {
		t.Fatal("duplicate subscription should not be started")
	}
	if s.jobs["duplicate"] != nil {
		t.Fatal("duplicate subscription should not be requeued")
	}
	if journal := readJournal(t, s); len(journal.Jobs) != 1 {
		t.Fatalf("journal should keep only the requeued job: %+v", journal.Jobs)
	}
}

func TestLoadJobsMarksRunningInterrupted(t *testing.T) {
	s := journalServer(t)
	old := startJob
	started := make(chan string, 10)
	startJob = func(_ *server, id string) { started <- id }
	defer func() { startJob = old }()

	id := "deadbeef"
	dir := filepath.Join(s.jobsDir, id)
	if err := os.MkdirAll(dir, 0o700); err != nil {
		t.Fatal(err)
	}
	startedAt := 50.0
	writeJournal(t, s, jobJournal{Version: 1, Jobs: []jobRecord{
		{ID: id, Status: "running", Created: 10, Started: &startedAt, Mode: "single",
			Pick: 3, SizeMB: 100, Core: "auto", IP: "1.2.3.4", SubKey: "key"},
	}})

	s.loadJobs()

	item := s.jobs[id]
	if item == nil || item.Status != "error" || item.ErrorCode != "interrupted" || item.Finished == nil {
		t.Fatalf("running job should be marked interrupted: %+v", item)
	}
	if len(started) != 0 {
		t.Fatalf("interrupted job must not restart: %v", started)
	}
	if _, err := os.Stat(dir); !os.IsNotExist(err) {
		t.Fatalf("interrupted job dir should be removed, stat err: %v", err)
	}
	if journal := readJournal(t, s); len(journal.Jobs) != 0 {
		t.Fatalf("journal should be empty after recovery: %+v", journal.Jobs)
	}
}

func TestLoadJobsSkipsBlockedAndMissingDirs(t *testing.T) {
	s := journalServer(t)
	s.blockedIPs["1.2.3.4"] = true
	old := startJob
	started := make(chan string, 10)
	startJob = func(_ *server, id string) { started <- id }
	defer func() { startJob = old }()

	blockedDir := filepath.Join(s.jobsDir, "blocked")
	if err := os.MkdirAll(blockedDir, 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(blockedDir, "sub.txt"), []byte("https://example.com/sub\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	writeJournal(t, s, jobJournal{Version: 1, Jobs: []jobRecord{
		{ID: "blocked", Status: "queued", Mode: "single", Pick: 1, IP: "1.2.3.4", SubKey: "a"},
		{ID: "missing", Status: "queued", Mode: "single", Pick: 1, IP: "5.6.7.8", SubKey: "b"},
	}})

	s.loadJobs()

	if len(started) != 0 || len(s.jobs) != 0 {
		t.Fatalf("blocked or missing jobs must not be recovered: started=%v jobs=%v", started, s.jobs)
	}
	if _, err := os.Stat(blockedDir); !os.IsNotExist(err) {
		t.Fatalf("blocked job dir should be removed, stat err: %v", err)
	}
}

func TestLoadJobsIgnoresCorruptJournal(t *testing.T) {
	s := journalServer(t)
	if err := os.WriteFile(s.jobsFile, []byte("{not json"), 0o600); err != nil {
		t.Fatal(err)
	}
	s.loadJobs()
	if len(s.jobs) != 0 {
		t.Fatalf("corrupt journal should be ignored: %v", s.jobs)
	}
}
