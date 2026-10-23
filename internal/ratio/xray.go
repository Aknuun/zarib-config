package ratio

import (
	"archive/zip"
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"time"
)

func CoreCandidates(choice string) []string {
	choice = strings.ToLower(strings.TrimSpace(choice))
	if choice == "" {
		choice = "auto"
	}
	if choice == "latest" || choice == "legacy" || choice == "system" {
		return []string{choice}
	}
	candidates := []string{"latest", "legacy"}
	if _, err := exec.LookPath("xray"); err == nil {
		candidates = append(candidates, "system")
	}
	return candidates
}

func xrayAssetName() (string, error) {
	switch strings.ToLower(runtime.GOARCH) {
	case "amd64":
		return "Xray-linux-64.zip", nil
	case "arm64":
		return "Xray-linux-arm64-v8a.zip", nil
	case "arm":
		return "Xray-linux-arm32-v7a.zip", nil
	default:
		return "", ratioErrorf("Automatic Xray download is not supported on %s.", runtime.GOARCH)
	}
}

func xrayRuns(path string) bool {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	return exec.CommandContext(ctx, path, "-version").Run() == nil
}

func XrayVersion(path string) string {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	output, err := exec.CommandContext(ctx, path, "-version").CombinedOutput()
	if err != nil && len(output) == 0 {
		return ""
	}
	line := strings.SplitN(strings.TrimSpace(string(output)), "\n", 2)[0]
	return strings.TrimSpace(line)
}

func EnsureXray(core string) (string, error) {
	core = strings.ToLower(strings.TrimSpace(core))
	if core != "latest" && core != "legacy" && core != "system" {
		core = "latest"
	}
	if core == "system" {
		found, err := exec.LookPath("xray")
		if err != nil {
			return "", ratioErrorf("No system Xray binary was found in PATH.")
		}
		return found, nil
	}
	cache := CacheDir()
	target := filepath.Join(cache, "xray")
	if core == "legacy" {
		target = filepath.Join(cache, "xray-"+LegacyXrayVersion)
	}
	if info, err := os.Stat(target); err == nil && !info.IsDir() && xrayRuns(target) {
		return target, nil
	}
	if err := os.MkdirAll(cache, 0o755); err != nil {
		return "", ratioErrorf("Could not create the Xray cache directory: %v", err)
	}
	asset, err := xrayAssetName()
	if err != nil {
		return "", err
	}
	releaseURL := "https://github.com/XTLS/Xray-core/releases/latest/download/" + asset
	if core == "legacy" {
		releaseURL = fmt.Sprintf("https://github.com/XTLS/Xray-core/releases/download/%s/%s", LegacyXrayVersion, asset)
	}
	fmt.Printf("Xray (%s) not found; downloading from GitHub (%s) ...\n", core, asset)
	request, err := http.NewRequest(http.MethodGet, releaseURL, nil)
	if err != nil {
		return "", ratioErrorf("Could not download Xray (%s): %v", core, err)
	}
	request.Header.Set("User-Agent", "ratio-test")
	client := &http.Client{Timeout: 120 * time.Second}
	response, err := client.Do(request)
	if err != nil {
		return "", ratioErrorf("Could not download Xray (%s): %v", core, err)
	}
	defer response.Body.Close()
	if response.StatusCode < 200 || response.StatusCode >= 300 {
		return "", ratioErrorf("Could not download Xray (%s): HTTP %d", core, response.StatusCode)
	}
	archive, err := os.CreateTemp(cache, core+"-"+asset+".*.zip")
	if err != nil {
		return "", ratioErrorf("Could not create a temporary archive: %v", err)
	}
	archivePath := archive.Name()
	defer os.Remove(archivePath)
	if _, err := io.Copy(archive, response.Body); err != nil {
		archive.Close()
		return "", ratioErrorf("Could not download Xray (%s): %v", core, err)
	}
	if err := archive.Close(); err != nil {
		return "", ratioErrorf("Could not save the Xray archive: %v", err)
	}
	reader, err := zip.OpenReader(archivePath)
	if err != nil {
		return "", ratioErrorf("Could not open the Xray archive: %v", err)
	}
	defer reader.Close()
	var member *zip.File
	for _, file := range reader.File {
		if file.Name == "xray" || strings.HasSuffix(file.Name, "/xray") {
			member = file
			break
		}
	}
	if member == nil {
		return "", ratioErrorf("The xray binary was not found inside the archive.")
	}
	source, err := member.Open()
	if err != nil {
		return "", ratioErrorf("Could not read the Xray binary: %v", err)
	}
	defer source.Close()
	temporary, err := os.CreateTemp(cache, filepath.Base(target)+".*.tmp")
	if err != nil {
		return "", ratioErrorf("Could not create a temporary Xray binary: %v", err)
	}
	temporaryPath := temporary.Name()
	defer os.Remove(temporaryPath)
	if _, err := io.Copy(temporary, source); err != nil {
		temporary.Close()
		return "", ratioErrorf("Could not extract the Xray binary: %v", err)
	}
	if err := temporary.Close(); err != nil {
		return "", ratioErrorf("Could not save the Xray binary: %v", err)
	}
	if err := os.Chmod(temporaryPath, 0o755); err != nil {
		return "", ratioErrorf("Could not make the Xray binary executable: %v", err)
	}
	if err := os.Rename(temporaryPath, target); err != nil {
		return "", ratioErrorf("Could not install the Xray binary: %v", err)
	}
	if !xrayRuns(target) {
		return "", ratioErrorf("The downloaded Xray (%s) binary is not executable.", core)
	}
	return target, nil
}

func FreePort() (int, error) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return 0, ratioErrorf("Could not allocate a local port: %v", err)
	}
	defer listener.Close()
	return listener.Addr().(*net.TCPAddr).Port, nil
}

type Child struct {
	Cmd     *exec.Cmd
	Done    chan struct{}
	mu      sync.Mutex
	waitErr error
	exited  bool
}

func StartChild(cmd *exec.Cmd) (*Child, error) {
	prepareChild(cmd)
	if err := cmd.Start(); err != nil {
		return nil, err
	}
	child := &Child{Cmd: cmd, Done: make(chan struct{})}
	go func() {
		err := cmd.Wait()
		child.mu.Lock()
		child.waitErr = err
		child.exited = true
		child.mu.Unlock()
		close(child.Done)
	}()
	return child, nil
}

func (child *Child) Exited() bool {
	if child == nil {
		return true
	}
	child.mu.Lock()
	defer child.mu.Unlock()
	return child.exited
}

func (child *Child) Wait() error {
	if child == nil {
		return nil
	}
	<-child.Done
	child.mu.Lock()
	defer child.mu.Unlock()
	return child.waitErr
}

func (child *Child) Stop() {
	if child == nil || child.Cmd == nil || child.Cmd.Process == nil || child.Exited() {
		return
	}
	pid := child.Cmd.Process.Pid
	_ = signalChild(pid, false)
	select {
	case <-child.Done:
		return
	case <-time.After(5 * time.Second):
	}
	_ = signalChild(pid, true)
	select {
	case <-child.Done:
	case <-time.After(5 * time.Second):
	}
}

func WaitForPort(port int, child *Child, logPath string, timeout time.Duration) error {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		select {
		case <-child.Done:
			content, _ := os.ReadFile(logPath)
			log := string(content)
			if len(log) > 2000 {
				log = log[len(log)-2000:]
			}
			return ratioErrorf("Xray exited before startup:\n%s", log)
		default:
		}
		connection, err := net.DialTimeout("tcp", fmt.Sprintf("127.0.0.1:%d", port), 300*time.Millisecond)
		if err == nil {
			connection.Close()
			return nil
		}
		time.Sleep(100 * time.Millisecond)
	}
	return ratioErrorf("Xray did not open the SOCKS port in time.")
}
