package web

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// jobJournal is the on-disk record of tests that were queued or running when
// the process stopped. Finished jobs are removed so the file stays small.
type jobJournal struct {
	Version int         `json:"version"`
	Jobs    []jobRecord `json:"jobs"`
}

type jobRecord struct {
	ID      string   `json:"id"`
	Status  string   `json:"status"`
	Created float64  `json:"created"`
	Started *float64 `json:"started,omitempty"`
	Mode    string   `json:"mode"`
	Pick    int      `json:"pick"`
	Picks   []int    `json:"picks,omitempty"`
	SizeMB  float64  `json:"size_mb"`
	Core    string   `json:"core"`
	IP      string   `json:"ip"`
	SubKey  string   `json:"sub_key"`
	SubURL  string   `json:"sub_url"`
	Wait    int      `json:"wait"`
	Drift   int      `json:"drift"`
	Claimed *float64 `json:"claimed,omitempty"`
	Sync    bool     `json:"sync"`
}

func (s *server) saveJobsLocked() {
	if s.jobsFile == "" {
		return
	}
	journal := jobJournal{Version: 1}
	for _, item := range s.jobs {
		if item.Status != "queued" && item.Status != "running" {
			continue
		}
		journal.Jobs = append(journal.Jobs, jobRecord{
			ID: item.ID, Status: item.Status, Created: item.Created, Started: item.Started,
			Mode: item.Mode, Pick: item.Pick, Picks: item.Picks, SizeMB: item.SizeMB,
			Core: item.Core, IP: item.IP, SubKey: item.SubKey, SubURL: item.SubURL,
			Wait: item.Wait, Drift: item.Drift, Claimed: item.Claimed, Sync: item.Sync,
		})
	}
	data, err := json.Marshal(journal)
	if err != nil {
		fmt.Printf("jobs: could not encode journal: %v\n", err)
		return
	}
	if err := writeFileAtomic(s.jobsFile, data); err != nil {
		fmt.Printf("jobs: could not save journal: %v\n", err)
	}
}

func writeFileAtomic(path string, data []byte) error {
	temp := path + ".tmp"
	file, err := os.OpenFile(temp, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0o600)
	if err != nil {
		return err
	}
	if _, err := file.Write(data); err != nil {
		file.Close()
		os.Remove(temp)
		return err
	}
	if err := file.Sync(); err != nil {
		file.Close()
		os.Remove(temp)
		return err
	}
	if err := file.Close(); err != nil {
		os.Remove(temp)
		return err
	}
	if err := os.Rename(temp, path); err != nil {
		os.Remove(temp)
		return err
	}
	if dir, err := os.Open(filepath.Dir(path)); err == nil {
		_ = dir.Sync()
		dir.Close()
	}
	return nil
}

// loadJobs restores the journal before the server starts accepting requests.
// Queued jobs are re-queued within the usual limits; jobs that were running
// when the process died are reported as interrupted instead of being re-run,
// because they may already have consumed subscription traffic.
func (s *server) loadJobs() {
	if s.jobsFile == "" {
		return
	}
	data, err := os.ReadFile(s.jobsFile)
	if err != nil {
		if !os.IsNotExist(err) {
			fmt.Printf("jobs: could not read journal: %v\n", err)
		}
		return
	}
	var journal jobJournal
	if err := json.Unmarshal(data, &journal); err != nil {
		fmt.Printf("jobs: could not parse journal: %v\n", err)
		return
	}
	now := nowSeconds()
	var start []string
	var removeDirs []string
	s.mu.Lock()
	seen := map[string]bool{}
	queued, perIP := 0, map[string]int{}
	for _, item := range s.jobs {
		if item.Status == "queued" {
			queued++
			perIP[item.IP]++
		}
	}
	for _, record := range journal.Jobs {
		if record.ID == "" || strings.ContainsAny(record.ID, `/\`) || record.ID == "." || record.ID == ".." {
			continue
		}
		if seen[record.SubKey] && record.SubKey != "" {
			continue
		}
		jobDir := filepath.Join(s.jobsDir, record.ID)
		if record.Status == "running" {
			seen[record.SubKey] = record.SubKey != ""
			started := record.Started
			finished := now
			s.jobs[record.ID] = &job{
				ID: record.ID, Status: "error", Stage: "error", Created: record.Created,
				Started: started, Finished: &finished, Mode: record.Mode, Pick: record.Pick,
				Picks: record.Picks, SizeMB: record.SizeMB, Core: record.Core, IP: record.IP,
				SubKey: record.SubKey, SubURL: record.SubURL,
				Error:     "The server restarted while this test was running. Please start it again.",
				ErrorCode: "interrupted",
			}
			removeDirs = append(removeDirs, jobDir)
			continue
		}
		if record.Status != "queued" {
			continue
		}
		subFile := filepath.Join(jobDir, "sub.txt")
		if _, err := os.Stat(subFile); err != nil {
			continue
		}
		if s.isBlockedIP(record.IP) {
			removeDirs = append(removeDirs, jobDir)
			continue
		}
		if queued >= maxQueued || perIP[record.IP] >= maxPerIP {
			removeDirs = append(removeDirs, jobDir)
			continue
		}
		seen[record.SubKey] = record.SubKey != ""
		queued++
		perIP[record.IP]++
		s.jobs[record.ID] = &job{
			ID: record.ID, Status: "queued", Stage: "queued", Created: record.Created,
			Mode: record.Mode, Pick: record.Pick, Picks: record.Picks, SizeMB: record.SizeMB,
			Core: record.Core, IP: record.IP, SubKey: record.SubKey, SubURL: record.SubURL,
			Wait: record.Wait, Drift: record.Drift, Claimed: record.Claimed, Sync: record.Sync,
			Dir: jobDir, SubFile: subFile, ResultFile: filepath.Join(jobDir, "result.json"),
			ControlFile: filepath.Join(jobDir, "control"),
		}
		start = append(start, record.ID)
	}
	s.saveJobsLocked()
	s.mu.Unlock()
	for _, dir := range removeDirs {
		_ = os.RemoveAll(dir)
	}
	for _, id := range start {
		go startJob(s, id)
	}
}

var startJob = func(s *server, id string) {
	s.runJob(id)
}

func (s *server) cleanupJobDirs() {
	entries, err := os.ReadDir(s.jobsDir)
	if err != nil {
		return
	}
	cutoff := time.Now().Add(-time.Hour)
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, entry := range entries {
		if _, keep := s.jobs[entry.Name()]; keep {
			continue
		}
		if info, err := entry.Info(); err == nil && info.ModTime().Before(cutoff) {
			_ = os.RemoveAll(filepath.Join(s.jobsDir, entry.Name()))
		}
	}
}
