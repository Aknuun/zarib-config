package web

import (
	"bufio"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/Aknuun/zarib-config/internal/ratio"
)

const (
	maxActive   = 5
	maxQueued   = 25
	maxLogLines = 120
	jobTimeout  = 30 * time.Minute
	minSizeMB   = ratio.MinSizeMB
	maxSizeMB   = ratio.MaxSizeMB
)

type job struct {
	ID          string
	Status      string
	Stage       string
	Progress    float64
	Message     string
	Created     float64
	Started     *float64
	Finished    *float64
	Result      *ratio.Result
	Error       string
	ErrorCode   string
	Log         []string
	Cancel      bool
	TimedOut    bool
	Pick        int
	SizeMB      float64
	Core        string
	Tunnel      *bool
	Dir         string
	SubFile     string
	ResultFile  string
	ControlFile string
	Child       *ratio.Child
}

var publicURLFn = ratio.IsPublicURL

type server struct {
	staticDir       string
	jobsDir         string
	statsFile       string
	telegramToken   string
	telegramChatID  string
	telegramAPIBase string
	mu              sync.Mutex
	jobs            map[string]*job
	rates           map[string]map[string][]time.Time
	presence        map[string]float64
	live            liveStats
	geoMu           sync.Mutex
	geoCache        map[string]geoEntry
	sem             chan struct{}
}

type publicJob struct {
	ID        string        `json:"id"`
	Status    string        `json:"status"`
	Stage     string        `json:"stage"`
	Progress  float64       `json:"progress"`
	Message   string        `json:"message"`
	Elapsed   float64       `json:"elapsed"`
	Created   float64       `json:"created"`
	Result    *ratio.Result `json:"result"`
	Error     string        `json:"error"`
	ErrorCode string        `json:"error_code"`
	Log       []string      `json:"log"`
	Pick      int           `json:"pick"`
	SizeMB    float64       `json:"size_mb"`
	Core      string        `json:"core"`
}

type inspectIn struct {
	SubURL string `json:"sub_url"`
}

type jobIn struct {
	SubURL  string   `json:"sub_url"`
	Pick    *int     `json:"pick"`
	SizeMB  *float64 `json:"size_mb"`
	Claimed *float64 `json:"claimed"`
	Wait    *int     `json:"wait"`
	Drift   *int     `json:"drift"`
	Sync    *bool    `json:"sync"`
	Core    *string  `json:"core"`
	Tunnel  *bool    `json:"tunnel"`
}

func Run(addr, staticDir string) error {
	_ = os.Setenv("RATIO_SAFE_FETCH", "1")
	absoluteStatic, err := filepath.Abs(staticDir)
	if err != nil {
		return err
	}
	if info, err := os.Stat(filepath.Join(absoluteStatic, "index.html")); err != nil || info.IsDir() {
		return fmt.Errorf("static directory is invalid: %s", absoluteStatic)
	}
	cacheRoot := os.Getenv("RATIO_WEB_CACHE")
	if cacheRoot == "" {
		home, err := os.UserHomeDir()
		if err != nil || home == "" {
			home = "."
		}
		cacheRoot = filepath.Join(home, ".cache", "v2ray-ratio-web")
	}
	jobsDir := filepath.Join(cacheRoot, "jobs")
	if err := os.MkdirAll(jobsDir, 0o700); err != nil {
		return err
	}
	if entries, err := os.ReadDir(jobsDir); err == nil {
		for _, entry := range entries {
			_ = os.RemoveAll(filepath.Join(jobsDir, entry.Name()))
		}
	}
	s := &server{
		staticDir:       absoluteStatic,
		jobsDir:         jobsDir,
		statsFile:       filepath.Join(cacheRoot, "stats.json"),
		telegramToken:   strings.TrimSpace(os.Getenv("ZARIB_TELEGRAM_TOKEN")),
		telegramChatID:  strings.TrimSpace(os.Getenv("ZARIB_TELEGRAM_CHAT_ID")),
		telegramAPIBase: "https://api.telegram.org",
		jobs:            map[string]*job{},
		rates:           map[string]map[string][]time.Time{},
		presence:        map[string]float64{},
		geoCache:        map[string]geoEntry{},
		sem:             make(chan struct{}, maxActive),
	}
	s.loadStats()
	go s.pruneLoop()
	fmt.Printf("zarib-config web listening on %s\n", addr)
	httpServer := &http.Server{
		Addr:              addr,
		Handler:           s.securityHeaders(http.HandlerFunc(s.route)),
		ReadHeaderTimeout: 5 * time.Second,
	}
	return httpServer.ListenAndServe()
}

func (s *server) securityHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("X-Frame-Options", "DENY")
		w.Header().Set("Referrer-Policy", "no-referrer")
		w.Header().Set("Permissions-Policy", "geolocation=(), microphone=(), camera=()")
		w.Header().Set("Content-Security-Policy",
			"default-src 'self'; script-src 'self'; "+
				"style-src 'self' 'unsafe-inline' https://cdn.jsdelivr.net; "+
				"font-src 'self' https://cdn.jsdelivr.net; img-src 'self' data:; "+
				"connect-src 'self'; base-uri 'none'; form-action 'self'; frame-ancestors 'none'")
		if r.Header.Get("X-Forwarded-Proto") == "https" {
			w.Header().Set("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
		}
		if strings.HasPrefix(r.URL.Path, "/api/") {
			w.Header().Set("Cache-Control", "no-store")
		}
		next.ServeHTTP(w, r)
	})
}

func (s *server) route(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path
	switch {
	case path == "/healthz":
		if !allowMethods(w, r, http.MethodGet, http.MethodHead) {
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"ok": true, "version": ratio.Version})
	case path == "/api/inspect":
		if !allowMethods(w, r, http.MethodPost) {
			return
		}
		s.inspect(w, r)
	case path == "/api/jobs":
		if !allowMethods(w, r, http.MethodPost) {
			return
		}
		s.createJob(w, r)
	case path == "/api/presence":
		if !allowMethods(w, r, http.MethodPost) {
			return
		}
		s.presenceHandler(w, r)
	case path == "/api/geo":
		if !allowMethods(w, r, http.MethodGet) {
			return
		}
		s.geoCheck(w, r)
	case path == "/api/ping":
		if !allowMethods(w, r, http.MethodGet) {
			return
		}
		s.pingCheck(w, r)
	case path == "/api/report":
		if !allowMethods(w, r, http.MethodPost) {
			return
		}
		s.report(w, r)
	case strings.HasPrefix(path, "/api/jobs/"):
		s.jobRoute(w, r)
	case path == "/":
		if !allowMethods(w, r, http.MethodGet, http.MethodHead) {
			return
		}
		w.Header().Set("Cache-Control", "no-cache")
		http.ServeFile(w, r, filepath.Join(s.staticDir, "index.html"))
	case strings.HasPrefix(path, "/static/"):
		if !allowMethods(w, r, http.MethodGet, http.MethodHead) {
			return
		}
		s.serveStatic(w, r)
	default:
		if strings.HasPrefix(path, "/api/") {
			apiError(w, "not_found", "Not found.", http.StatusNotFound)
			return
		}
		http.NotFound(w, r)
	}
}

func allowMethods(w http.ResponseWriter, r *http.Request, methods ...string) bool {
	for _, method := range methods {
		if r.Method == method {
			return true
		}
	}
	w.Header().Set("Allow", strings.Join(methods, ", "))
	apiError(w, "not_found", "Not found.", http.StatusMethodNotAllowed)
	return false
}

func (s *server) serveStatic(w http.ResponseWriter, r *http.Request) {
	relative := strings.TrimPrefix(r.URL.Path, "/static/")
	cleaned := filepath.Clean(relative)
	if cleaned == "." || strings.HasPrefix(cleaned, "..") || filepath.IsAbs(cleaned) {
		http.NotFound(w, r)
		return
	}
	target := filepath.Join(s.staticDir, cleaned)
	if !strings.HasPrefix(target, s.staticDir+string(os.PathSeparator)) && target != s.staticDir {
		http.NotFound(w, r)
		return
	}
	info, err := os.Stat(target)
	if err != nil || info.IsDir() {
		http.NotFound(w, r)
		return
	}
	w.Header().Set("Cache-Control", "public, max-age=3600")
	http.ServeFile(w, r, target)
}

func (s *server) inspect(w http.ResponseWriter, r *http.Request) {
	ip := clientIP(r)
	if !s.rateLimit("inspect", ip, 30) {
		apiError(w, "rate_limited", "Too many requests.", http.StatusTooManyRequests)
		return
	}
	var payload inspectIn
	if err := decodeJSON(w, r, &payload); err != nil {
		apiError(w, "invalid_input", "Invalid request.", http.StatusUnprocessableEntity)
		return
	}
	subURL := strings.TrimSpace(payload.SubURL)
	if !isValidURL(subURL) {
		apiError(w, "invalid_url", "The subscription URL is not valid.", http.StatusBadRequest)
		return
	}
	if !publicURLFn(subURL) {
		apiError(w, "blocked_url", "The subscription URL must be publicly reachable.", http.StatusBadRequest)
		return
	}
	body, headers, err := ratio.Fetch(subURL, 20*time.Second)
	if err != nil {
		apiError(w, "fetch_failed", err.Error(), http.StatusBadRequest)
		return
	}
	usage := ratio.ParseUsage(headers, body)
	configs := ratio.ExtractConfigs(body)
	if len(configs) == 0 {
		apiError(w, "no_configs", "No supported configs found in the subscription.", http.StatusBadRequest)
		return
	}
	items := make([]map[string]any, 0, minInt(len(configs), 100))
	for index, item := range configs {
		if index >= 100 {
			break
		}
		claimed, _ := ratio.DetectClaimed(item.Remark, nil)
		items = append(items, map[string]any{
			"index": index + 1, "protocol": item.Protocol, "target": item.Target,
			"remark": item.Remark, "claimed": claimed,
			"uri": ratio.ConfigURI(item),
		})
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"usage": usageView(usage), "configs": items, "count": len(configs),
		"insecure": strings.HasPrefix(strings.ToLower(subURL), "http://"),
	})
}

func (s *server) createJob(w http.ResponseWriter, r *http.Request) {
	ip := clientIP(r)
	if !s.rateLimit("jobs", ip, 60) {
		apiError(w, "rate_limited", "Too many tests. Try again later.", http.StatusTooManyRequests)
		return
	}
	var payload jobIn
	if err := decodeJSON(w, r, &payload); err != nil {
		apiError(w, "invalid_input", "Invalid request.", http.StatusUnprocessableEntity)
		return
	}
	subURL := strings.TrimSpace(payload.SubURL)
	if !isValidURL(subURL) {
		apiError(w, "invalid_url", "The subscription URL is not valid.", http.StatusBadRequest)
		return
	}
	if !publicURLFn(subURL) {
		apiError(w, "blocked_url", "The subscription URL must be publicly reachable.", http.StatusBadRequest)
		return
	}
	pick := 1
	if payload.Pick != nil {
		pick = *payload.Pick
	}
	sizeMB := ratio.DefaultSizeMB
	if payload.SizeMB != nil {
		sizeMB = *payload.SizeMB
	}
	wait := ratio.DefaultWait
	if payload.Wait != nil {
		wait = *payload.Wait
	}
	drift := 0
	if payload.Drift != nil {
		drift = *payload.Drift
	}
	syncMode := false
	if payload.Sync != nil {
		syncMode = *payload.Sync
	}
	core := "auto"
	if payload.Core != nil {
		core = *payload.Core
	}
	if sizeMB < minSizeMB || sizeMB > maxSizeMB {
		apiError(w, "invalid_input", "Invalid test size.", http.StatusBadRequest)
		return
	}
	if pick < 1 || pick > 500 {
		apiError(w, "invalid_input", "Invalid config number.", http.StatusBadRequest)
		return
	}
	if payload.Claimed != nil && (*payload.Claimed < 0.1 || *payload.Claimed > 100) {
		apiError(w, "invalid_input", "Invalid claimed ratio.", http.StatusBadRequest)
		return
	}
	if wait < 0 || wait > 60 || drift < 0 || drift > 30 {
		apiError(w, "invalid_input", "Invalid wait values.", http.StatusBadRequest)
		return
	}
	validCore := false
	for _, choice := range ratio.CoreChoices {
		if core == choice {
			validCore = true
			break
		}
	}
	if !validCore {
		apiError(w, "invalid_input", "Invalid Xray core.", http.StatusBadRequest)
		return
	}
	_, queued := s.activeCounts()
	if queued >= maxQueued {
		apiError(w, "busy", "The test queue is full.", http.StatusTooManyRequests)
		return
	}
	jobID, err := randomID()
	if err != nil {
		apiError(w, "unknown", "Could not create the test.", http.StatusInternalServerError)
		return
	}
	jobDir := filepath.Join(s.jobsDir, jobID)
	if err := os.MkdirAll(jobDir, 0o700); err != nil {
		apiError(w, "unknown", "Could not create the test.", http.StatusInternalServerError)
		return
	}
	subFile := filepath.Join(jobDir, "sub.txt")
	if err := os.WriteFile(subFile, []byte(subURL+"\n"), 0o600); err != nil {
		os.RemoveAll(jobDir)
		apiError(w, "unknown", "Could not create the test.", http.StatusInternalServerError)
		return
	}
	controlFile := filepath.Join(jobDir, "control")
	if err := os.WriteFile(controlFile, nil, 0o600); err != nil {
		os.RemoveAll(jobDir)
		apiError(w, "unknown", "Could not create the test.", http.StatusInternalServerError)
		return
	}
	now := float64(time.Now().UnixNano()) / 1e9
	item := &job{
		ID: jobID, Status: "queued", Stage: "queued", Progress: 0, Created: now,
		Pick: pick, SizeMB: sizeMB, Core: core, Tunnel: payload.Tunnel,
		Dir: jobDir, SubFile: subFile, ResultFile: filepath.Join(jobDir, "result.json"),
		ControlFile: controlFile,
	}
	s.mu.Lock()
	s.jobs[jobID] = item
	s.mu.Unlock()
	go s.runJob(jobID, wait, drift, payload.Claimed, syncMode)
	writeJSON(w, http.StatusOK, map[string]any{"id": jobID, "status": "queued"})
}

func (s *server) runJob(jobID string, wait, drift int, claimed *float64, syncMode bool) {
	s.sem <- struct{}{}
	defer func() { <-s.sem }()
	s.mu.Lock()
	item := s.jobs[jobID]
	if item == nil {
		s.mu.Unlock()
		return
	}
	if item.Cancel {
		item.Status = "cancelled"
		item.Stage = "cancelled"
		finished := nowSeconds()
		item.Finished = &finished
		jobDir := item.Dir
		s.mu.Unlock()
		os.RemoveAll(jobDir)
		return
	}
	item.Status = "running"
	started := nowSeconds()
	item.Started = &started
	item.Stage = "reading"
	item.Progress = 3
	subFile, resultFile, jobDir, controlFile := item.SubFile, item.ResultFile, item.Dir, item.ControlFile
	pick, sizeMB, core := item.Pick, item.SizeMB, item.Core
	s.mu.Unlock()
	executable, err := os.Executable()
	if err != nil {
		s.finishError(jobID, "unknown", err.Error())
		os.RemoveAll(jobDir)
		return
	}
	args := []string{
		"--sub-file", subFile, "--pick", strconv.Itoa(pick),
		"--size", strconv.FormatFloat(sizeMB, 'g', -1, 64),
		"--chunk", strconv.FormatFloat(sizeMB, 'g', -1, 64),
		"--wait", strconv.Itoa(wait), "--drift-window", strconv.Itoa(drift), "--json-out", resultFile,
		"--control-file", controlFile, "--core", core, "--probe-retries", "0", "--probe-wait", "0",
	}
	if syncMode {
		args = append(args, "--sync")
	}
	if claimed != nil {
		args = append(args, "--claimed", strconv.FormatFloat(*claimed, 'g', -1, 64))
	}
	command := exec.Command(executable, args...)
	command.Env = withEnv(map[string]string{
		"RATIO_SAFE_FETCH": "1", "NO_COLOR": "1", "HOME": homeDir(),
		"XDG_CACHE_HOME": filepath.Join(homeDir(), ".cache"),
	})
	command.Stdin = nil
	stdout, err := command.StdoutPipe()
	if err != nil {
		s.finishError(jobID, "unknown", err.Error())
		os.RemoveAll(jobDir)
		return
	}
	command.Stderr = command.Stdout
	child, err := ratio.StartChildGroup(command)
	if err != nil {
		s.finishError(jobID, "unknown", err.Error())
		os.RemoveAll(jobDir)
		return
	}
	s.mu.Lock()
	if current := s.jobs[jobID]; current != nil {
		current.Child = child
	}
	s.mu.Unlock()
	timer := time.AfterFunc(jobTimeout, func() {
		s.mu.Lock()
		if current := s.jobs[jobID]; current != nil {
			current.TimedOut = true
		}
		s.mu.Unlock()
		child.Stop()
	})
	scanner := bufio.NewScanner(stdout)
	scanner.Buffer(make([]byte, 4096), 64*1024)
	for scanner.Scan() {
		line := scanner.Text()
		s.appendLog(jobID, line)
		s.progressFromLine(jobID, line)
	}
	waitErr := child.Wait()
	child.Stop()
	timer.Stop()
	s.mu.Lock()
	current := s.jobs[jobID]
	cancelled := current != nil && current.Cancel
	timedOut := current != nil && current.TimedOut
	logs := []string{}
	if current != nil {
		logs = append(logs, current.Log...)
	}
	s.mu.Unlock()
	exitCode := 0
	if waitErr != nil {
		exitCode = 1
		var exitErr *exec.ExitError
		if errors.As(waitErr, &exitErr) {
			exitCode = exitErr.ExitCode()
		}
	}
	if cancelled {
		s.mu.Lock()
		if current := s.jobs[jobID]; current != nil {
			finished := nowSeconds()
			current.Status = "cancelled"
			current.Stage = "cancelled"
			current.Progress = 100
			current.Finished = &finished
			current.Error = "Cancelled."
			current.ErrorCode = "cancelled"
			current.Child = nil
		}
		s.mu.Unlock()
	} else if exitCode == 0 {
		content, readErr := os.ReadFile(resultFile)
		if readErr != nil {
			s.finishError(jobID, "unknown", "Could not read the result: "+readErr.Error())
		} else {
			var result ratio.Result
			if err := json.Unmarshal(content, &result); err != nil {
				s.finishError(jobID, "unknown", "Could not read the result: "+err.Error())
			} else {
				var doneTunnel *bool
				s.mu.Lock()
				if current := s.jobs[jobID]; current != nil {
					finished := nowSeconds()
					current.Status = "done"
					current.Stage = "done"
					current.Progress = 100
					current.Message = "Done"
					current.Finished = &finished
					current.Result = &result
					current.Child = nil
					doneTunnel = current.Tunnel
				}
				s.mu.Unlock()
				s.recordResult(&result, doneTunnel)
			}
		}
	} else {
		s.finishError(jobID, classifyError(logs, timedOut, cancelled), technicalError(logs, exitCode))
	}
	os.RemoveAll(jobDir)
}

func (s *server) finishError(jobID, code, message string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	current := s.jobs[jobID]
	if current == nil {
		return
	}
	finished := nowSeconds()
	current.Status = "error"
	current.Stage = "error"
	current.Finished = &finished
	current.Error = message
	current.ErrorCode = code
	current.Child = nil
}

func (s *server) appendLog(jobID, line string) {
	line = ansiPattern.ReplaceAllString(line, "")
	if index := strings.LastIndex(line, "\r"); index >= 0 {
		line = line[index+1:]
	}
	line = strings.TrimRight(line, " \t\r\n")
	if line == "" {
		return
	}
	if len(line) > 500 {
		line = line[:500]
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	current := s.jobs[jobID]
	if current == nil {
		return
	}
	current.Log = append(current.Log, line)
	if len(current.Log) > maxLogLines {
		current.Log = current.Log[len(current.Log)-maxLogLines:]
	}
}

var ansiPattern = regexp.MustCompile(`\x1b\[[0-9;]*[A-Za-z]`)
var downloadProgressPattern = regexp.MustCompile(`total\s+([\d.]+)\s*MB\s+of\s+([\d.]+)\s*MB`)

func (s *server) progressFromLine(jobID, line string) {
	stripped := strings.TrimSpace(line)
	if stripped == "" {
		return
	}
	stage := ""
	progress := -1.0
	message := stripped
	switch {
	case strings.Contains(stripped, "Reading subscription info"):
		stage, progress = "reading", 5
	case strings.Contains(stripped, "Current subscription usage"):
		stage, progress = "reading", 10
	case strings.Contains(stripped, "Xray started") || (strings.Contains(stripped, "Xray (") && strings.Contains(stripped, "started")):
		stage, progress = "connecting", 15
	case strings.Contains(stripped, "Trying another Xray core"):
		stage, progress = "connecting", 15
	case strings.Contains(stripped, "Connection probe failed"):
		stage, progress = "connecting", 18
	case strings.Contains(stripped, "Retrying connection"):
		stage, progress = "connecting", 18
	case strings.Contains(stripped, "Connection OK"):
		stage, progress = "baseline", 25
	case strings.Contains(stripped, "Measuring background usage"):
		stage, progress = "drift", 30
	case strings.Contains(stripped, "Background usage"):
		stage, progress = "drift", 35
	case strings.Contains(stripped, "Starting test download"):
		stage, progress = "downloading", 35
	case strings.Contains(stripped, "Download finished"):
		stage, progress = "waiting", 85
	case strings.Contains(stripped, "Waiting for the panel"):
		stage, progress = "waiting", 85
	case strings.Contains(stripped, "Waiting for user confirmation"):
		stage, progress = "usage_pending", 90
	case strings.Contains(stripped, "Continuing to wait for usage"):
		stage, progress = "waiting_extended", 92
	case strings.Contains(stripped, "Checking subscription usage"):
		stage, progress = "waiting", 92
	case strings.Contains(stripped, "did not change yet"):
		stage, progress = "waiting", 90
	case strings.Contains(stripped, "COEFFICIENT TEST RESULT"):
		stage, progress = "done", 100
	}
	if match := downloadProgressPattern.FindStringSubmatch(stripped); len(match) == 3 {
		total, _ := strconv.ParseFloat(match[1], 64)
		target, _ := strconv.ParseFloat(match[2], 64)
		if target <= 0 {
			target = 1
		}
		stage = "downloading"
		progress = 35 + 50*minFloat(total/target, 1)
	}
	if stage == "" {
		return
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	current := s.jobs[jobID]
	if current == nil {
		return
	}
	if progress >= 0 && progress > current.Progress {
		current.Progress = progress
	}
	current.Stage = stage
	current.Message = message
}

func (s *server) jobRoute(w http.ResponseWriter, r *http.Request) {
	rest := strings.TrimPrefix(r.URL.Path, "/api/jobs/")
	parts := strings.Split(strings.Trim(rest, "/"), "/")
	if len(parts) == 0 || parts[0] == "" {
		apiError(w, "not_found", "Job not found.", http.StatusNotFound)
		return
	}
	jobID := parts[0]
	if len(parts) == 1 {
		if !allowMethods(w, r, http.MethodGet) {
			return
		}
		s.mu.Lock()
		current := s.jobs[jobID]
		if current == nil {
			s.mu.Unlock()
			apiError(w, "not_found", "Job not found.", http.StatusNotFound)
			return
		}
		response := publicJobFrom(current)
		s.mu.Unlock()
		writeJSON(w, http.StatusOK, response)
		return
	}
	if len(parts) == 2 && parts[1] == "cancel" {
		if !allowMethods(w, r, http.MethodPost) {
			return
		}
		s.cancelJob(w, jobID)
		return
	}
	if len(parts) == 2 && parts[1] == "extend" {
		if !allowMethods(w, r, http.MethodPost) {
			return
		}
		s.extendJob(w, jobID)
		return
	}
	apiError(w, "not_found", "Job not found.", http.StatusNotFound)
}

func (s *server) extendJob(w http.ResponseWriter, jobID string) {
	s.mu.Lock()
	current := s.jobs[jobID]
	if current == nil {
		s.mu.Unlock()
		apiError(w, "not_found", "Job not found.", http.StatusNotFound)
		return
	}
	controlFile := current.ControlFile
	status := current.Status
	stage := current.Stage
	s.mu.Unlock()
	if status != "running" || controlFile == "" {
		apiError(w, "invalid_state", "This test is not waiting for usage.", http.StatusConflict)
		return
	}
	if stage != "usage_pending" && stage != "waiting_extended" {
		apiError(w, "invalid_state", "This test is not waiting for usage.", http.StatusConflict)
		return
	}
	if err := os.WriteFile(controlFile, []byte("extend\n"), 0o600); err != nil {
		apiError(w, "unknown", "Could not extend the wait.", http.StatusInternalServerError)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"ok": true})
}

func (s *server) cancelJob(w http.ResponseWriter, jobID string) {
	s.mu.Lock()
	current := s.jobs[jobID]
	if current == nil {
		s.mu.Unlock()
		apiError(w, "not_found", "Job not found.", http.StatusNotFound)
		return
	}
	current.Cancel = true
	child := current.Child
	var queuedDir string
	if current.Status == "queued" {
		finished := nowSeconds()
		current.Status = "cancelled"
		current.Stage = "cancelled"
		current.Finished = &finished
		current.Error = "Cancelled."
		current.ErrorCode = "cancelled"
		queuedDir = current.Dir
	}
	s.mu.Unlock()
	if child != nil {
		child.Stop()
	}
	if queuedDir != "" {
		os.RemoveAll(queuedDir)
	}
	writeJSON(w, http.StatusOK, map[string]any{"ok": true})
}

func publicJobFrom(item *job) publicJob {
	now := nowSeconds()
	elapsed := 0.0
	if item.Finished != nil && item.Started != nil {
		elapsed = *item.Finished - *item.Started
	} else if item.Started != nil {
		elapsed = now - *item.Started
	}
	logs := item.Log
	if len(logs) > 80 {
		logs = logs[len(logs)-80:]
	}
	copied := append([]string(nil), logs...)
	return publicJob{
		ID: item.ID, Status: item.Status, Stage: item.Stage,
		Progress: round1(item.Progress), Message: item.Message,
		Elapsed: round1(elapsed), Created: item.Created, Result: item.Result,
		Error: item.Error, ErrorCode: item.ErrorCode, Log: copied,
		Pick: item.Pick, SizeMB: item.SizeMB, Core: item.Core,
	}
}

func (s *server) activeCounts() (int, int) {
	s.mu.Lock()
	defer s.mu.Unlock()
	running, queued := 0, 0
	for _, item := range s.jobs {
		switch item.Status {
		case "running":
			running++
		case "queued":
			queued++
		}
	}
	return running, queued
}

func (s *server) rateLimit(bucket, ip string, limit int) bool {
	now := time.Now()
	window := 3600 * time.Second
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.rates == nil {
		s.rates = map[string]map[string][]time.Time{}
	}
	entries := s.rates[bucket]
	if entries == nil {
		entries = map[string][]time.Time{}
		s.rates[bucket] = entries
	}
	queue := entries[ip]
	cutoff := now.Add(-window)
	start := 0
	for start < len(queue) && queue[start].Before(cutoff) {
		start++
	}
	queue = queue[start:]
	if len(queue) >= limit {
		entries[ip] = queue
		return false
	}
	entries[ip] = append(queue, now)
	return true
}

func (s *server) pruneLoop() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()
	for range ticker.C {
		now := nowSeconds()
		cutoff := now - 3600
		s.mu.Lock()
		for id, item := range s.jobs {
			if item.Finished != nil && *item.Finished < cutoff {
				delete(s.jobs, id)
			}
		}
		for id, seen := range s.presence {
			if now-seen > presenceWindow.Seconds() {
				delete(s.presence, id)
			}
		}
		rateCutoff := time.Now().Add(-time.Hour)
		for bucket, entries := range s.rates {
			for ip, queue := range entries {
				kept := queue[:0]
				for _, at := range queue {
					if at.After(rateCutoff) {
						kept = append(kept, at)
					}
				}
				if len(kept) == 0 {
					delete(entries, ip)
				} else {
					entries[ip] = kept
				}
			}
			if len(entries) == 0 {
				delete(s.rates, bucket)
			}
		}
		s.mu.Unlock()
		s.geoMu.Lock()
		for host, entry := range s.geoCache {
			if time.Now().After(entry.Expires) {
				delete(s.geoCache, host)
			}
		}
		s.geoMu.Unlock()
	}
}

func usageView(usage *ratio.Usage) any {
	if usage == nil {
		return nil
	}
	upload := usage.Upload
	download := usage.Download
	used := upload + download
	var total any
	var remaining any
	if usage.Total != nil && *usage.Total != 0 {
		total = *usage.Total
		value := *usage.Total - used
		if value < 0 {
			value = 0
		}
		remaining = value
	}
	var expire any
	var daysLeft any
	if usage.Expire != nil && *usage.Expire != 0 {
		expire = *usage.Expire
		days := (*usage.Expire - time.Now().Unix()) / 86400
		if days < 0 {
			days = 0
		}
		daysLeft = days
	}
	return map[string]any{
		"upload": upload, "download": download, "used": used, "total": total,
		"remaining": remaining, "expire": expire, "days_left": daysLeft,
	}
}

func isValidURL(rawURL string) bool {
	parsed, err := url.Parse(rawURL)
	if err != nil {
		return false
	}
	return (parsed.Scheme == "http" || parsed.Scheme == "https") && parsed.Hostname() != ""
}

func clientIP(r *http.Request) string {
	if real := strings.TrimSpace(r.Header.Get("X-Real-IP")); real != "" {
		return real
	}
	if forwarded := r.Header.Get("X-Forwarded-For"); forwarded != "" {
		return strings.TrimSpace(strings.Split(forwarded, ",")[0])
	}
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err == nil {
		return host
	}
	return r.RemoteAddr
}

func decodeJSON(w http.ResponseWriter, r *http.Request, target any) error {
	r.Body = http.MaxBytesReader(w, r.Body, 16*1024)
	decoder := json.NewDecoder(r.Body)
	return decoder.Decode(target)
}

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}

func apiError(w http.ResponseWriter, code, message string, status int) {
	writeJSON(w, status, map[string]any{"code": code, "error": message})
}

func classifyError(logs []string, timedOut, cancelled bool) string {
	if cancelled {
		return "cancelled"
	}
	if timedOut {
		return "timeout"
	}
	text := strings.ToLower(strings.Join(logs, "\n"))
	switch {
	case strings.Contains(text, "did not respond through the proxy"):
		return "probe_failed"
	case strings.Contains(text, "cannot resolve host"):
		return "resolve_failed"
	case strings.Contains(text, "not valid for xray") || strings.Contains(text, "failed to start") || strings.Contains(text, "no xray core"):
		return "xray_failed"
	case strings.Contains(text, "could not read the subscription") || strings.Contains(text, "subscription request failed"):
		return "fetch_failed"
	case strings.Contains(text, "no supported configs") || strings.Contains(text, "no configs found"):
		return "no_configs"
	case strings.Contains(text, "config number") || strings.Contains(text, "invalid config"):
		return "pick_out_of_range"
	case strings.Contains(text, "did not change") || strings.Contains(text, "no subscription-userinfo"):
		return "usage_unchanged"
	case strings.Contains(text, "timed out") || strings.Contains(text, "timeout"):
		return "timeout"
	default:
		return "unknown"
	}
}

func technicalError(logs []string, exitCode int) string {
	var picked []string
	for _, line := range logs {
		if strings.HasPrefix(line, "Error:") || strings.Contains(line, "probe failed") {
			picked = append(picked, line)
		}
	}
	if len(picked) > 6 {
		picked = picked[len(picked)-6:]
	}
	if len(picked) > 0 {
		return strings.Join(picked, "\n")
	}
	var tail []string
	start := len(logs) - 8
	if start < 0 {
		start = 0
	}
	for _, line := range logs[start:] {
		if strings.TrimSpace(line) != "" {
			tail = append(tail, line)
		}
	}
	if len(tail) > 0 {
		return strings.Join(tail, "\n")
	}
	return fmt.Sprintf("Test process exited with code %d.", exitCode)
}

func randomID() (string, error) {
	buffer := make([]byte, 16)
	if _, err := rand.Read(buffer); err != nil {
		return "", err
	}
	return hex.EncodeToString(buffer), nil
}

func withEnv(overrides map[string]string) []string {
	var result []string
	for _, entry := range os.Environ() {
		key := strings.SplitN(entry, "=", 2)[0]
		if _, ok := overrides[key]; !ok {
			result = append(result, entry)
		}
	}
	for key, value := range overrides {
		result = append(result, key+"="+value)
	}
	return result
}

func homeDir() string {
	home, err := os.UserHomeDir()
	if err != nil || home == "" {
		return "."
	}
	return home
}

func nowSeconds() float64 {
	return float64(time.Now().UnixNano()) / 1e9
}

func round1(value float64) float64 {
	return float64(int64(value*10+0.5)) / 10
}

func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func minFloat(a, b float64) float64 {
	if a < b {
		return a
	}
	return b
}
