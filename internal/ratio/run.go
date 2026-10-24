package ratio

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

var ErrNoConfigs = errors.New("no configs")

type Options struct {
	Sub          string
	SubFile      string
	Config       string
	SizeMB       *float64
	FreeSize     bool
	Wait         int
	PollTimeout  int
	DriftWindow  int
	Pick         int
	List         bool
	ChunkMB      float64
	Timeout      float64
	TestURL      string
	ProbeURL     string
	ProbeTimeout int
	ProbeOnly    bool
	ProbeRetries int
	ProbeWait    int
	Sync         bool
	ControlFile  string
	Core         string
	Claimed      *float64
	JSONOut      string
	KeepTemp     bool
}

func DefaultOptions() Options {
	return Options{
		Wait: DefaultWait, PollTimeout: DefaultPollTimeout, DriftWindow: DefaultDriftWindow,
		ChunkMB: DefaultChunkMB, Timeout: 180, TestURL: DefaultTestURL,
		ProbeURL: DefaultProbeURL, ProbeTimeout: DefaultProbeTimeout,
		ProbeRetries: DefaultProbeRetries, ProbeWait: DefaultProbeWait, Core: "auto",
	}
}

func ensureOutboundTag(outbounds []any, outbound map[string]any) string {
	tag := toString(outbound["tag"])
	if tag != "" {
		return tag
	}
	used := make(map[string]bool, len(outbounds))
	for _, item := range outbounds {
		if other := asObject(item); other != nil {
			used[toString(other["tag"])] = true
		}
	}
	tag = "ratio-test-proxy"
	for suffix := 2; used[tag]; suffix++ {
		tag = fmt.Sprintf("ratio-test-proxy-%d", suffix)
	}
	outbound["tag"] = tag
	return tag
}

func BuildClientConfig(input string) (map[string]any, map[string]any, string, string, string, error) {
	text := strings.TrimSpace(input)
	remark := ""
	if strings.HasPrefix(text, "{") {
		var config map[string]any
		decoder := json.NewDecoder(strings.NewReader(text))
		decoder.UseNumber()
		if err := decoder.Decode(&config); err != nil {
			return nil, nil, "", "", "", ratioErrorf("Invalid config JSON: %v", err)
		}
		if config == nil {
			return nil, nil, "", "", "", ratioErrorf("Config JSON must be an object.")
		}
		outbounds := asArray(config["outbounds"])
		if len(outbounds) == 0 {
			return nil, nil, "", "", "", ratioErrorf("The JSON config has no outbound; provide a client config, not a server config.")
		}
		var target string
		for _, item := range outbounds {
			outbound := asObject(item)
			if outbound == nil || DirectProtocols[toString(outbound["protocol"])] {
				continue
			}
			target = ensureOutboundTag(outbounds, outbound)
			break
		}
		if target == "" {
			return nil, nil, "", "", "", ratioErrorf("The JSON config has no proxy outbound; provide a client config with a proxy outbound.")
		}
		for _, item := range outbounds {
			outbound := asObject(item)
			if outbound == nil {
				continue
			}
			if value := toString(outbound["remarks"]); value != "" {
				remark = value
				break
			}
		}
		return config, nil, target, remark, text, nil
	}
	if !strings.Contains(text, "://") {
		return nil, nil, "", "", "", ratioErrorf("Config must be a link (vless/vmess/trojan/ss) or JSON.")
	}
	outbound, err := OutboundFromURI(text)
	if err != nil {
		return nil, nil, "", "", "", err
	}
	remark = configRemark(text, outbound)
	delete(outbound, "_remark")
	return nil, outbound, "", remark, text, nil
}

func PrepareConfig(base map[string]any, outbound map[string]any, targetTag string, port int) (map[string]any, error) {
	var config map[string]any
	if base == nil {
		config = map[string]any{
			"log":      map[string]any{"loglevel": "warning"},
			"inbounds": []any{},
			"outbounds": []any{
				outbound,
				map[string]any{"protocol": "freedom", "tag": "direct"},
			},
		}
		targetTag = toString(outbound["tag"])
	} else {
		payload, err := json.Marshal(base)
		if err != nil {
			return nil, ratioErrorf("Could not copy the JSON config: %v", err)
		}
		decoder := json.NewDecoder(strings.NewReader(string(payload)))
		decoder.UseNumber()
		if err := decoder.Decode(&config); err != nil {
			return nil, ratioErrorf("Could not copy the JSON config: %v", err)
		}
		config["log"] = map[string]any{"loglevel": "warning"}
		config["inbounds"] = []any{}
	}
	tag := "ratio-test-socks"
	inbounds := asArray(config["inbounds"])
	inbounds = append(inbounds, map[string]any{
		"tag": tag, "listen": "127.0.0.1", "port": port, "protocol": "socks",
		"settings": map[string]any{"auth": "noauth", "udp": false},
	})
	config["inbounds"] = inbounds
	if targetTag != "" {
		config["routing"] = map[string]any{
			"domainStrategy": "AsIs",
			"rules": []any{
				map[string]any{"type": "field", "inboundTag": []any{tag}, "outboundTag": targetTag},
			},
		}
	} else {
		delete(config, "routing")
	}
	return config, nil
}

func ChooseConfig(configs []Config, pick int) (Config, error) {
	if pick != 0 {
		if pick >= 1 && pick <= len(configs) {
			return configs[pick-1], nil
		}
		return Config{}, ratioErrorf("Config number must be between 1 and %d.", len(configs))
	}
	if len(configs) == 1 {
		fmt.Printf("Found one config in the subscription: %s\n", configs[0].Remark)
		return configs[0], nil
	}
	fmt.Println()
	fmt.Println("Available configs in the subscription:")
	for index, item := range configs {
		remark := item.Remark
		if remark == "" {
			remark = "(unnamed)"
		}
		fmt.Printf("  %d) %s | %s | %s\n", index+1, item.Protocol, item.Target, remark)
	}
	if !IsInteractive() {
		return Config{}, ratioErrorf("Multiple configs found; use --pick N to choose one.")
	}
	for {
		def := "1"
		raw, err := Ask("Config number", &def)
		if err != nil {
			return Config{}, err
		}
		index, err := strconv.Atoi(NormalizeDigits(raw))
		if err == nil && index >= 1 && index <= len(configs) {
			return configs[index-1], nil
		}
		fmt.Println("Invalid number.")
	}
}

func CurlSize(ctx context.Context, port int, targetURL string, maxTime int, speedLimit int) (int64, string) {
	args := []string{
		"-sS", "--socks5-hostname", fmt.Sprintf("127.0.0.1:%d", port),
		"--connect-timeout", "10", "--max-time", strconv.Itoa(maxTime),
	}
	if speedLimit > 0 {
		args = append(args, "--speed-limit", strconv.Itoa(speedLimit), "--speed-time", "30")
	}
	args = append(args, "-o", os.DevNull, "-w", "%{size_download}", targetURL)
	requestCtx, cancel := context.WithTimeout(ctx, time.Duration(maxTime+15)*time.Second)
	defer cancel()
	command := exec.CommandContext(requestCtx, "curl", args...)
	var stdout, stderr strings.Builder
	command.Stdout = &stdout
	command.Stderr = &stderr
	err := command.Run()
	if requestCtx.Err() != nil {
		return 0, "timeout"
	}
	size, _ := strconv.ParseFloat(strings.TrimSpace(stdout.String()), 64)
	if err != nil {
		if stderr.Len() > 0 {
			return int64(size), strings.TrimSpace(stderr.String())
		}
		return int64(size), err.Error()
	}
	return int64(size), strings.TrimSpace(stderr.String())
}

func CurlProbe(ctx context.Context, port int, targetURL string, maxTime int) (int, int64, string) {
	args := []string{
		"-sS", "--socks5-hostname", fmt.Sprintf("127.0.0.1:%d", port),
		"--connect-timeout", "10", "--max-time", strconv.Itoa(maxTime),
		"-o", os.DevNull, "-w", "%{http_code} %{size_download}", targetURL,
	}
	requestCtx, cancel := context.WithTimeout(ctx, time.Duration(maxTime+15)*time.Second)
	defer cancel()
	command := exec.CommandContext(requestCtx, "curl", args...)
	var stdout, stderr strings.Builder
	command.Stdout = &stdout
	command.Stderr = &stderr
	err := command.Run()
	if requestCtx.Err() != nil {
		return 0, 0, "timeout"
	}
	if err != nil && stdout.Len() == 0 {
		return 0, 0, strings.TrimSpace(stderr.String())
	}
	parts := strings.Fields(stdout.String())
	code := 0
	size := int64(0)
	if len(parts) > 0 {
		code, _ = strconv.Atoi(parts[0])
	}
	if len(parts) > 1 {
		value, _ := strconv.ParseFloat(parts[1], 64)
		size = int64(value)
	}
	return code, size, strings.TrimSpace(stderr.String())
}

func Probe(ctx context.Context, port int, urls []string, timeout int) (int64, error) {
	if timeout < 1 {
		timeout = 1
	}
	deadline := time.Now().Add(time.Duration(timeout) * time.Second)
	probeCtx, cancel := context.WithTimeout(ctx, time.Duration(timeout)*time.Second)
	defer cancel()
	for index, targetURL := range urls {
		remaining := time.Until(deadline)
		if remaining <= 0 {
			break
		}
		perURL := int(remaining.Seconds())
		switch left := len(urls) - index; {
		case left >= 3:
			perURL = 8
		case left == 2:
			perURL = 6
		}
		if seconds := int(remaining.Seconds()); perURL > seconds {
			perURL = seconds
		}
		if perURL < 1 {
			perURL = 1
		}
		code, size, curlErr := CurlProbe(probeCtx, port, targetURL, perURL)
		if code >= 200 && code < 400 {
			return size, nil
		}
		detail := curlErr
		if detail == "" {
			if code != 0 {
				detail = fmt.Sprintf("HTTP %d", code)
			} else {
				detail = "no response"
			}
		}
		fmt.Fprintf(os.Stderr, "  probe failed: %s (%s)\n", detail, targetURL)
	}
	return 0, probeFailuref("The config did not respond through the proxy to our test server within %d seconds. TCP ping may still work, but the node may block the test destination or be too slow. Try another config.", timeout)
}

var probeFn = Probe

func ProbeWithRetry(ctx context.Context, port int, urls []string, timeout, retries, waitSeconds int, interactive bool) error {
	attempts := retries + 1
	if attempts < 1 {
		attempts = 1
	}
	for attempt := 1; attempt <= attempts; attempt++ {
		_, err := probeFn(ctx, port, urls, timeout)
		if err == nil {
			return nil
		}
		if attempt >= attempts {
			return err
		}
		fmt.Println(C(fmt.Sprintf("Connection probe failed: %v", err), "yellow"))
		if interactive && !AskYesNo(fmt.Sprintf("Wait %ds and retry the connection?", waitSeconds), true) {
			return err
		}
		countdownFn(waitSeconds, "Retrying connection")
	}
	return nil
}

var countdownFn = Countdown

func Download(ctx context.Context, port int, template string, total, chunk int64, chunkTimeout int) int64 {
	done := int64(0)
	attempts := 0
	for done < total {
		if ctx.Err() != nil {
			break
		}
		size := chunk
		if remaining := total - done; remaining < size {
			size = remaining
		}
		targetURL := strings.ReplaceAll(template, "{bytes}", strconv.FormatInt(size, 10))
		targetURL = strings.ReplaceAll(targetURL, "{remaining}", strconv.FormatInt(total-done, 10))
		received, _ := CurlSize(ctx, port, targetURL, chunkTimeout, 10240)
		done += received
		attempts++
		fmt.Fprintf(os.Stderr, "%s\n", C(fmt.Sprintf("  received %s | total %s of %s",
			Human(received), Human(done), Human(total)), "dim"))
		if received <= 0 {
			if attempts >= 6 {
				break
			}
			time.Sleep(time.Duration(1+attempts) * time.Second)
		} else if done < total && received < size/2 && attempts >= 6 {
			break
		}
	}
	return done
}

type usageReading struct {
	At    time.Time
	Usage *Usage
}

func ReadUsageAfter(ctx context.Context, subURL string, baseline Usage, window int) (*Usage, []usageReading, time.Time, error) {
	return readUsageAfterWithLabel(ctx, subURL, baseline, window, "Checking subscription usage ...")
}

func readUsageAfterWithLabel(ctx context.Context, subURL string, baseline Usage, window int, label string) (*Usage, []usageReading, time.Time, error) {
	baseTotal := baseline.Used()
	if window < 1 {
		window = 1
	}
	deadline := time.Now().Add(time.Duration(window) * time.Second)
	var readings []usageReading
	for {
		if label != "" {
			fmt.Println(C(label, "dim"))
		}
		reading, err := ReadUsage(subURL, 1, 3*time.Second)
		if err == nil && reading != nil {
			readings = append(readings, usageReading{At: time.Now(), Usage: reading})
			if reading.Used() > baseTotal {
				break
			}
		}
		remaining := time.Until(deadline)
		if remaining <= 0 {
			break
		}
		if remaining > time.Second {
			remaining = time.Second
		}
		select {
		case <-ctx.Done():
			return nil, readings, time.Time{}, ctx.Err()
		case <-time.After(remaining):
		}
	}
	if len(readings) == 0 {
		return nil, nil, time.Time{}, ratioErrorf("Could not read usage after the test.")
	}
	best := readings[0]
	for _, reading := range readings[1:] {
		if reading.Usage.Used() > best.Usage.Used() {
			best = reading
		}
	}
	return best.Usage, readings, best.At, nil
}

var readUsageAfterFn = ReadUsageAfter

var usageWaitStages = []int{5, 15, 30, 30}

func WaitForUsage(ctx context.Context, subURL string, before Usage, wait, pollTimeout, retries int, interactive bool, expectedBytes int64) (*Usage, []usageReading, time.Time, error) {
	baseline := before.Used()
	var best *Usage
	var bestTime time.Time
	var readings []usageReading
	for index, window := range usageWaitStages {
		if window < 1 {
			window = 1
		}
		fmt.Println(C(fmt.Sprintf("Waiting for the panel to record usage (up to %ds) ...", window), "dim"))
		after, got, afterTime, err := readUsageAfterFn(ctx, subURL, before, window)
		if err == nil && after != nil && after.Used() > baseline {
			best, bestTime, readings = after, afterTime, got
			break
		}
		if ctx.Err() != nil {
			if err != nil {
				return nil, nil, time.Time{}, err
			}
			return nil, nil, time.Time{}, ctx.Err()
		}
		if err != nil {
			fmt.Println(C(fmt.Sprintf("Could not read usage: %v", err), "yellow"))
		}
		if index < len(usageWaitStages)-1 {
			fmt.Println(C("Usage reported by the subscription did not change yet.", "yellow"))
		}
	}
	if best == nil {
		return nil, nil, time.Time{}, ratioErrorf("Usage reported by the subscription did not change after 52 seconds. The panel may be delayed or may only show usage in its dashboard.")
	}
	if expectedBytes <= 0 || best.Used()-baseline >= expectedBytes {
		return best, readings, bestTime, nil
	}
	settle := time.Duration(wait) * time.Second
	if settle < 4*time.Second {
		settle = 4 * time.Second
	}
	if settle > 12*time.Second {
		settle = 12 * time.Second
	}
	fmt.Println(C(fmt.Sprintf("Usage changed by %s; waiting up to %s for the panel to finish recording ...", Human(best.Used()-baseline), settle), "dim"))
	stableSince := time.Now()
	deadline := time.Now().Add(settle + 8*time.Second)
	for {
		if best.Used()-baseline >= expectedBytes || time.Since(stableSince) >= settle || time.Now().After(deadline) {
			break
		}
		select {
		case <-ctx.Done():
			return best, readings, bestTime, ctx.Err()
		case <-time.After(time.Second):
		}
		reading, got, at, err := readUsageAfterFn(ctx, subURL, before, 1)
		if err == nil && reading != nil {
			readings = append(readings, got...)
			if reading.Used() > best.Used() {
				best, bestTime = reading, at
				stableSince = time.Now()
			}
		}
	}
	return best, readings, bestTime, nil
}

func readControlAction(path string) string {
	data, err := os.ReadFile(path)
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(data))
}

var (
	usageInitialWait  = 120 * time.Second
	usageExtendedWait = 5 * time.Minute
	usageExtendedPoll = 10 * time.Second
)

func WaitForUsageWithControl(ctx context.Context, subURL string, before Usage, expectedBytes int64, controlFile string) (*Usage, []usageReading, time.Time, error) {
	baseline := before.Used()
	start := time.Now()
	var best *Usage
	var bestTime time.Time
	var readings []usageReading
	poll := func() bool {
		fmt.Println(C("Checking subscription usage ...", "dim"))
		reading, err := ReadUsage(subURL, 1, 3*time.Second)
		if err != nil || reading == nil {
			return false
		}
		readings = append(readings, usageReading{At: time.Now(), Usage: reading})
		if reading.Used() > baseline {
			if best == nil || reading.Used() > best.Used() {
				best = reading
				bestTime = time.Now()
			}
			if expectedBytes <= 0 || best.Used()-baseline >= expectedBytes {
				return true
			}
		}
		return false
	}
	for {
		if poll() {
			return best, readings, bestTime, nil
		}
		elapsed := time.Since(start)
		if elapsed >= usageInitialWait {
			break
		}
		interval := 5 * time.Second
		if elapsed < 10*time.Second {
			interval = time.Second
		}
		if remaining := usageInitialWait - elapsed; interval > remaining {
			interval = remaining
		}
		select {
		case <-ctx.Done():
			return best, readings, bestTime, ctx.Err()
		case <-time.After(interval):
		}
	}
	if best != nil {
		return best, readings, bestTime, nil
	}
	if controlFile == "" {
		if !IsInteractive() {
			return nil, nil, time.Time{}, ratioErrorf("Usage reported by the subscription did not change after 120 seconds. The panel may be delayed or may only show usage in its dashboard.")
		}
		fmt.Println(C("Usage did not change after 120 seconds.", "yellow"))
		if !AskYesNo("Wait up to 5 more minutes for the panel to record usage?", true) {
			return nil, nil, time.Time{}, ratioErrorf("Usage reported by the subscription did not change after 120 seconds.")
		}
		goto extended
	}
	fmt.Println(C("Usage did not change after 120 seconds. Waiting for user confirmation ...", "yellow"))
	for {
		switch readControlAction(controlFile) {
		case "extend":
			goto extended
		case "cancel":
			return nil, nil, time.Time{}, ratioErrorf("Cancelled.")
		}
		select {
		case <-ctx.Done():
			return nil, nil, time.Time{}, ctx.Err()
		case <-time.After(time.Second):
		}
	}
extended:
	fmt.Println(C("Continuing to wait for usage (up to 5 minutes) ...", "cyan"))
	extendedStart := time.Now()
	for {
		if poll() {
			return best, readings, bestTime, nil
		}
		if time.Since(extendedStart) >= usageExtendedWait {
			break
		}
		select {
		case <-ctx.Done():
			return best, readings, bestTime, ctx.Err()
		case <-time.After(usageExtendedPoll):
		}
	}
	if best != nil {
		return best, readings, bestTime, nil
	}
	return nil, nil, time.Time{}, ratioErrorf("Usage did not change after 5 minutes. The panel may be delayed; try again later.")
}

type Result struct {
	Version                  string   `json:"version"`
	TestedAt                 string   `json:"tested_at"`
	Config                   *string  `json:"config"`
	Core                     string   `json:"core"`
	XrayVersion              string   `json:"xray_version"`
	DownloadedBytes          int64    `json:"downloaded_bytes"`
	UsageBefore              Usage    `json:"usage_before"`
	UsageAfter               Usage    `json:"usage_after"`
	DeltaUpload              int64    `json:"delta_upload"`
	DeltaDownload            int64    `json:"delta_download"`
	DeltaTotal               int64    `json:"delta_total"`
	Ratio                    float64  `json:"ratio"`
	RatioRaw                 float64  `json:"ratio_raw"`
	BackgroundBytesPerSecond float64  `json:"background_bytes_per_second"`
	BackgroundBytesInTest    int64    `json:"background_bytes_in_test"`
	ClaimedRatio             *float64 `json:"claimed_ratio"`
	ClaimedMatches           *bool    `json:"claimed_matches"`
	Verdict                  *string  `json:"verdict"`
	RemainingBytes           *int64   `json:"remaining_bytes"`
	TotalBytes               *int64   `json:"total_bytes"`
	UploadAfter              *int64   `json:"upload_after"`
	DownloadAfter            *int64   `json:"download_after"`
	DurationSeconds          float64  `json:"duration_seconds"`
}

type runResult struct {
	Before     Usage
	After      Usage
	Readings   []usageReading
	AfterTime  time.Time
	Downloaded int64
	Elapsed    float64
	Core       string
	XrayPath   string
	DriftRate  float64
	DriftUsed  int64
	ProbeOnly  bool
}

func findProxyOutbound(base map[string]any, targetTag string) map[string]any {
	if base == nil {
		return nil
	}
	for _, item := range asArray(base["outbounds"]) {
		outbound := asObject(item)
		if outbound == nil || DirectProtocols[toString(outbound["protocol"])] {
			continue
		}
		if targetTag == "" || toString(outbound["tag"]) == targetTag {
			return outbound
		}
	}
	return nil
}

func runSingle(ctx context.Context, opts Options, subURL, configText string) (*runResult, error) {
	baseConfig, outbound, targetTag, _, scanText, err := BuildClientConfig(configText)
	if err != nil {
		return nil, err
	}
	_ = scanText
	proxyOutbound := outbound
	if baseConfig != nil {
		proxyOutbound = findProxyOutbound(baseConfig, targetTag)
	}
	if proxyOutbound != nil {
		if err := EnsureResolvable(OutboundHost(proxyOutbound)); err != nil {
			return nil, err
		}
	}
	if baseConfig != nil {
		fmt.Println(C("Using a test-safe copy of the JSON config (custom routing rules removed).", "dim"))
	}
	coreNames := CoreCandidates(opts.Core)
	var coreErrors []string
	var xrayPath, coreUsed string
	var child *Child
	var tmpdir, cfgPath, logPath string
	success := false
	defer func() {
		if child != nil {
			child.Stop()
		}
		if tmpdir != "" {
			if opts.KeepTemp && success {
				fmt.Println(C("Temporary files: "+tmpdir, "dim"))
			} else {
				os.RemoveAll(tmpdir)
			}
		}
	}()
	for coreIndex, coreName := range coreNames {
		xrayPath, err = EnsureXray(coreName)
		if err != nil {
			coreErrors = append(coreErrors, fmt.Sprintf("%s: %v", coreName, err))
			if len(coreNames) > 1 {
				fmt.Println(C(fmt.Sprintf("Xray core '%s' is unavailable: %v", coreName, err), "yellow"))
			}
			continue
		}
		port, err := FreePort()
		if err != nil {
			return nil, err
		}
		config, err := PrepareConfig(baseConfig, outbound, targetTag, port)
		if err != nil {
			return nil, err
		}
		tmpdir, err = os.MkdirTemp("", "ratio-test-")
		if err != nil {
			return nil, ratioErrorf("Could not create temporary files: %v", err)
		}
		cfgPath = filepath.Join(tmpdir, "config.json")
		logPath = filepath.Join(tmpdir, "xray.log")
		payload, err := json.MarshalIndent(config, "", "  ")
		if err != nil {
			return nil, ratioErrorf("Could not encode the Xray config: %v", err)
		}
		if err := os.WriteFile(cfgPath, payload, 0o600); err != nil {
			return nil, ratioErrorf("Could not write the Xray config: %v", err)
		}
		testCtx, cancel := context.WithTimeout(ctx, 40*time.Second)
		check := exec.CommandContext(testCtx, xrayPath, "run", "-test", "-c", cfgPath)
		output, checkErr := check.CombinedOutput()
		cancel()
		if checkErr != nil {
			message := string(output)
			if len(message) > 2000 {
				message = message[len(message)-2000:]
			}
			err = ratioErrorf("The config is not valid for Xray:\n%s", message)
			coreErrors = append(coreErrors, fmt.Sprintf("%s: %v", coreName, err))
			child = nil
			os.RemoveAll(tmpdir)
			tmpdir = ""
			if coreIndex < len(coreNames)-1 {
				fmt.Println(C(fmt.Sprintf("Xray core '%s' failed: %v", coreName, err), "yellow"))
				fmt.Println(C("Trying another Xray core ...", "cyan"))
				continue
			}
			return nil, err
		}
		logFile, err := os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o600)
		if err != nil {
			return nil, ratioErrorf("Could not open the Xray log: %v", err)
		}
		command := exec.Command(xrayPath, "run", "-c", cfgPath)
		command.Stdout = logFile
		command.Stderr = logFile
		child, err = StartChild(command)
		logFile.Close()
		if err != nil {
			return nil, ratioErrorf("Could not start Xray: %v", err)
		}
		if err := WaitForPort(port, child, logPath, 15*time.Second); err != nil {
			coreErrors = append(coreErrors, fmt.Sprintf("%s: %v", coreName, err))
			child.Stop()
			child = nil
			os.RemoveAll(tmpdir)
			tmpdir = ""
			if coreIndex < len(coreNames)-1 {
				fmt.Println(C(fmt.Sprintf("Xray core '%s' failed: %v", coreName, err), "yellow"))
				fmt.Println(C("Trying another Xray core ...", "cyan"))
				continue
			}
			return nil, err
		}
		fmt.Println(C(fmt.Sprintf("Xray (%s) started; probing connectivity through the config ...", coreName), "cyan"))
		probeURLs := []string{opts.ProbeURL}
		if opts.ProbeURL == DefaultProbeURL {
			probeURLs = ProbeURLs
		}
		if err := ProbeWithRetry(ctx, port, probeURLs, opts.ProbeTimeout, opts.ProbeRetries, opts.ProbeWait, IsInteractive()); err != nil {
			coreErrors = append(coreErrors, fmt.Sprintf("%s: %v", coreName, err))
			child.Stop()
			child = nil
			os.RemoveAll(tmpdir)
			tmpdir = ""
			if coreIndex < len(coreNames)-1 {
				fmt.Println(C(fmt.Sprintf("Xray core '%s' failed: %v", coreName, err), "yellow"))
				fmt.Println(C("Trying another Xray core ...", "cyan"))
				continue
			}
			return nil, err
		}
		coreUsed = coreName
		if opts.ProbeOnly {
			fmt.Println(C(fmt.Sprintf("Connection OK (%s).", coreUsed), "green"))
			success = true
			return &runResult{Core: coreUsed, XrayPath: xrayPath, ProbeOnly: true}, nil
		}
		fmt.Println(C(fmt.Sprintf("Connection OK (%s). Reading baseline usage ...", coreUsed), "green"))
		time.Sleep(2 * time.Second)
		before, err := ReadUsage(subURL, 3, 3*time.Second)
		if err != nil {
			return nil, err
		}
		beforeTime := time.Now()
		driftRate := 0.0
		reportBackground := func(driftBytes int64, driftWindow float64) {
			if driftBytes <= 0 {
				fmt.Println(C("No background usage detected.", "dim"))
				return
			}
			driftRate = float64(driftBytes) / driftWindow
			if driftBytes >= 64*1024 {
				fmt.Println(C(fmt.Sprintf("Warning: the subscription is consuming data while idle (%s in %.0fs, about %s/s).",
					Human(driftBytes), driftWindow, HumanFloat(driftRate)), "yellow"))
				fmt.Println(C("Results will be corrected using this background rate.", "dim"))
			} else {
				fmt.Println(C(fmt.Sprintf("Background usage is small (%s) and will be accounted for.", Human(driftBytes)), "dim"))
			}
		}
		if opts.Sync {
			fmt.Println(C("Measuring background usage across the panel's update cycle ...", "cyan"))
			first, _, firstTime, readErr := readUsageAfterWithLabel(ctx, subURL, *before, 75, "Measuring background usage (waiting for the panel update) ...")
			if readErr != nil {
				return nil, readErr
			}
			if first.Used() == before.Used() {
				fmt.Println(C("No background usage detected.", "dim"))
				before = first
				beforeTime = firstTime
			} else {
				second, _, secondTime, readErr := readUsageAfterWithLabel(ctx, subURL, *first, 75, "Measuring background usage (waiting for the next panel update) ...")
				if readErr != nil {
					return nil, readErr
				}
				driftWindow := secondTime.Sub(firstTime).Seconds()
				if driftWindow < 0.001 {
					driftWindow = 0.001
				}
				reportBackground(second.Used()-first.Used(), driftWindow)
				before = second
				beforeTime = secondTime
			}
		} else if opts.DriftWindow > 0 {
			driftStart := *before
			driftStartTime := beforeTime
			Countdown(opts.DriftWindow, "Measuring background usage")
			before, err = ReadUsage(subURL, 3, 3*time.Second)
			if err != nil {
				return nil, err
			}
			beforeTime = time.Now()
			driftWindow := beforeTime.Sub(driftStartTime).Seconds()
			if driftWindow < 0.001 {
				driftWindow = 0.001
			}
			reportBackground(before.Used()-driftStart.Used(), driftWindow)
		}
		totalBytes := int64(0)
		if opts.SizeMB != nil {
			totalBytes = int64(*opts.SizeMB * 1024 * 1024)
		} else {
			totalBytes = int64(DefaultSizeMB * 1024 * 1024)
		}
		chunkBytes := int64(opts.ChunkMB * 1024 * 1024)
		if chunkBytes < 256*1024 {
			chunkBytes = 256 * 1024
		}
		fmt.Println(C(fmt.Sprintf("Starting test download (%s) ...", Human(totalBytes)), "cyan"))
		started := time.Now()
		downloaded := Download(ctx, port, opts.TestURL, totalBytes, chunkBytes, int(opts.Timeout))
		elapsed := time.Since(started).Seconds()
		if downloaded <= 0 {
			return nil, ratioErrorf("No data was downloaded through the config.")
		}
		if downloaded < totalBytes {
			fmt.Println(C(fmt.Sprintf("Warning: only %s of %s was downloaded.", Human(downloaded), Human(totalBytes)), "yellow"))
		}
		fmt.Println(C(fmt.Sprintf("Download finished: %s in %.1fs.", Human(downloaded), elapsed), "green"))
		after, readings, afterTime, err := WaitForUsageWithControl(ctx, subURL, *before, downloaded, opts.ControlFile)
		if err != nil {
			return nil, err
		}
		driftUsed := int64(driftRate * maxFloat(0, afterTime.Sub(beforeTime).Seconds()))
		success = true
		return &runResult{
			Before: *before, After: *after, Readings: readings, AfterTime: afterTime,
			Downloaded: downloaded, Elapsed: elapsed, Core: coreUsed, XrayPath: xrayPath,
			DriftRate: driftRate, DriftUsed: driftUsed,
		}, nil
	}
	if len(coreErrors) > 0 {
		return nil, ratioErrorf("No Xray core could start the config. %s", strings.Join(coreErrors, " | "))
	}
	return nil, ratioErrorf("No Xray core could start the config.")
}

func maxFloat(a, b float64) float64 {
	if a > b {
		return a
	}
	return b
}

func Run(ctx context.Context, opts Options) (*Result, error) {
	subURL := opts.Sub
	if opts.SubFile != "" {
		content, err := os.ReadFile(opts.SubFile)
		if err != nil {
			return nil, ratioErrorf("Could not read the subscription URL file: %v", err)
		}
		subURL = strings.TrimSpace(string(content))
	}
	if subURL == "" {
		value, err := Ask("Subscription URL", nil)
		if err != nil {
			return nil, err
		}
		subURL = value
	}
	if subURL == "" {
		return nil, ratioErrorf("Subscription URL is required.")
	}
	if opts.List {
		body, _, err := Fetch(subURL, 30*time.Second)
		if err != nil {
			return nil, err
		}
		configs := ExtractConfigs(body)
		if len(configs) == 0 {
			fmt.Println("No supported configs found in the subscription.")
			return nil, ErrNoConfigs
		}
		for index, item := range configs {
			remark := item.Remark
			if remark == "" {
				remark = "(unnamed)"
			}
			fmt.Printf("%d) %s | %s | %s\n", index+1, item.Protocol, item.Target, remark)
		}
		return nil, nil
	}
	configArg := opts.Config
	var availableConfigs []Config
	if configArg == "" {
		fmt.Println("Reading configs from the subscription ...")
		configs := []Config{}
		body, _, fetchErr := Fetch(subURL, 30*time.Second)
		if fetchErr != nil {
			fmt.Printf("Could not read configs from the subscription: %v\n", fetchErr)
			if !IsInteractive() {
				return nil, fetchErr
			}
		} else {
			configs = ExtractConfigs(body)
		}
		if len(configs) > 0 {
			availableConfigs = configs
			selected, err := ChooseConfig(configs, opts.Pick)
			if err != nil {
				return nil, err
			}
			if selected.JSON != nil {
				payload, _ := json.Marshal(selected.JSON)
				configArg = string(payload)
			} else {
				configArg = selected.URI
			}
		} else {
			if !IsInteractive() {
				return nil, ratioErrorf("No supported configs found in the subscription.")
			}
			fmt.Println("No configs found in the subscription; enter one manually.")
			value, err := Ask("Config (JSON file path or vless:// / vmess:// / trojan:// / ss:// link)", nil)
			if err != nil {
				return nil, err
			}
			configArg = value
		}
	}
	if configArg == "" {
		return nil, ratioErrorf("A config is required.")
	}
	configText := configArg
	if !strings.HasPrefix(strings.TrimSpace(configArg), "{") {
		expanded := expandUser(configArg)
		if info, err := os.Stat(expanded); err == nil && !info.IsDir() {
			if content, err := os.ReadFile(expanded); err == nil {
				configText = string(content)
			}
		}
	}
	sizeMB := DefaultSizeMB
	if opts.SizeMB != nil {
		sizeMB = *opts.SizeMB
	} else if IsInteractive() {
		def := fmt.Sprintf("%g", DefaultSizeMB)
		raw, err := Ask("Test size in MB", &def)
		if err != nil {
			return nil, err
		}
		value, err := ParseFloat(raw)
		if err != nil {
			return nil, ratioErrorf("Invalid test size.")
		}
		sizeMB = value
	}
	if sizeMB <= 0 {
		return nil, ratioErrorf("Test size must be greater than zero.")
	}
	if !opts.FreeSize {
		if sizeMB < MinSizeMB {
			return nil, ratioErrorf("Minimum test size is %g MB.", MinSizeMB)
		}
		if sizeMB > MaxSizeMB {
			return nil, ratioErrorf("Maximum test size is %g MB.", MaxSizeMB)
		}
	}
	opts.SizeMB = &sizeMB
	wait := opts.Wait
	if wait < 0 {
		wait = 0
	}
	opts.Wait = wait
	if opts.PollTimeout < wait {
		opts.PollTimeout = wait
	}
	fmt.Println(C("Reading subscription info ...", "cyan"))
	initialUsage, err := ReadUsage(subURL, 3, 3*time.Second)
	if err != nil {
		return nil, err
	}
	totalBytes := int64(sizeMB * 1024 * 1024)
	var remaining *int64
	if initialUsage.Total != nil {
		value := *initialUsage.Total - initialUsage.Upload - initialUsage.Download
		remaining = &value
		if value > 0 && value < totalBytes {
			fmt.Println(C(fmt.Sprintf("Warning: remaining subscription volume (%s) is less than the test size.", Human(value)), "yellow"))
		}
	}
	usageLine := fmt.Sprintf("Current subscription usage: upload %s | download %s",
		Human(initialUsage.Upload), Human(initialUsage.Download))
	if remaining != nil {
		usageLine += " | remaining " + Human(*remaining)
	}
	fmt.Println(C(usageLine, "dim"))
	interactive := IsInteractive()
	var final *runResult
	var claimed *float64
	var remark string
	for {
		_, _, _, remarkText, scanText, err := BuildClientConfig(configText)
		if err != nil {
			return nil, err
		}
		remark = remarkText
		claimed, _ = DetectClaimed(scanText+" "+remark, opts.Claimed)
		result, err := runSingle(ctx, opts, subURL, configText)
		if err != nil {
			var probeErr *ProbeFailure
			if errors.As(err, &probeErr) && interactive && len(availableConfigs) > 0 {
				fmt.Println(C("The config did not connect.", "yellow"))
				def := ""
				raw, askErr := Ask("Enter another config number to try, or press Enter to stop", &def)
				if askErr == nil {
					index, _ := strconv.Atoi(NormalizeDigits(raw))
					if index >= 1 && index <= len(availableConfigs) {
						selected := availableConfigs[index-1]
						if selected.JSON != nil {
							payload, _ := json.Marshal(selected.JSON)
							configText = string(payload)
						} else {
							configText = selected.URI
						}
						continue
					}
				}
			}
			return nil, err
		}
		final = result
		break
	}
	if final.ProbeOnly {
		return nil, nil
	}
	deltaUp := final.After.Upload - final.Before.Upload
	deltaDown := final.After.Download - final.Before.Download
	used := deltaUp + deltaDown
	if used <= 0 {
		return nil, ratioErrorf("Usage reported by the subscription did not change. The panel may be delayed or may only show usage in its dashboard.")
	}
	rawRatio := float64(used) / float64(final.Downloaded)
	netUsed := used
	if final.DriftUsed > 0 {
		candidate := used - final.DriftUsed
		if candidate > 0 && candidate >= final.Downloaded/4 {
			netUsed = candidate
		} else {
			fmt.Println(C("Warning: the background estimate was too large to apply safely; showing the raw result.", "yellow"))
			final.DriftUsed = 0
		}
	}
	ratio := float64(netUsed) / float64(final.Downloaded)
	matches := false
	var verdict *string
	var verdictColor string
	coefColor := "green"
	if claimed != nil {
		matches = absFloat(ratio-*claimed) <= maxFloat(0.05, *claimed*0.03)
		value := ""
		if matches {
			value = "matches the claim"
			verdictColor = "green"
		} else if ratio > *claimed {
			value = "higher than the claim"
			verdictColor = "red"
			coefColor = "red"
		} else {
			value = "lower than the claim"
			verdictColor = "green"
		}
		verdict = &value
	} else {
		verdictColor = "dim"
		if ratio > 1.5 {
			coefColor = "red"
		} else if ratio > 1.1 {
			coefColor = "yellow"
		}
	}
	fmt.Println()
	fmt.Println(C(strings.Repeat("=", 54), "cyan"))
	fmt.Println(C("              COEFFICIENT TEST RESULT", "cyan", "bold"))
	fmt.Println(C(strings.Repeat("=", 54), "cyan"))
	if remark != "" {
		fmt.Println(C("Config: "+remark, "dim"))
	}
	fmt.Println()
	fmt.Println(C("  REAL COEFFICIENT", "bold"))
	fmt.Println()
	for _, line := range BigText(fmt.Sprintf("%.2fx", ratio)) {
		fmt.Println("    " + C(line, coefColor, "bold"))
	}
	fmt.Println()
	fmt.Printf("  For every %s downloaded, %s is deducted from your subscription.\n",
		C("1 MB", "bold"), C(fmt.Sprintf("%.2f MB", ratio), coefColor, "bold"))
	fmt.Println()
	kv := func(label, value string, color string) {
		fmt.Printf("  %s %s\n", C(fmt.Sprintf("%-24s", label), "dim"), C(value, color))
	}
	kv("Downloaded", Human(final.Downloaded), "bold")
	kv("Deducted", Human(netUsed), "bold")
	if final.DriftUsed > 0 {
		kv("Panel recorded", Human(used), "dim")
		kv("Background", Human(final.DriftUsed), "dim")
	}
	if claimed != nil {
		fmt.Printf("  %s %s  %s\n", C(fmt.Sprintf("%-24s", "Claimed"), "dim"),
			C(fmt.Sprintf("%.2fx", *claimed), "bold"), C(*verdict, verdictColor))
	} else {
		kv("Claimed", "not found", "dim")
	}
	var remainingAfter *int64
	if final.After.Total != nil {
		value := *final.After.Total - final.After.Upload - final.After.Download
		remainingAfter = &value
		if value > 0 {
			kv("Remaining", fmt.Sprintf("%s  (~%s real)", Human(value), HumanFloat(float64(value)/ratio)), "bold")
		}
	}
	if final.DriftUsed > 0 {
		fmt.Println(C(fmt.Sprintf("  background corrected by %s (raw %.3fx)", Human(final.DriftUsed), rawRatio), "dim"))
	}
	if ratio < 0.95 {
		fmt.Println(C("  Warning: coefficient is below 1; usage may be recorded late.", "yellow"))
	}
	fmt.Println(C(strings.Repeat("=", 54), "cyan"))
	result := &Result{
		Version: Version, TestedAt: time.Now().UTC().Format(time.RFC3339Nano),
		Core: final.Core, XrayVersion: XrayVersion(final.XrayPath),
		DownloadedBytes: final.Downloaded, UsageBefore: final.Before, UsageAfter: final.After,
		DeltaUpload: deltaUp, DeltaDownload: deltaDown, DeltaTotal: used,
		Ratio: roundTo(ratio, 4), RatioRaw: roundTo(rawRatio, 4),
		BackgroundBytesPerSecond: roundTo(final.DriftRate, 1),
		BackgroundBytesInTest:    final.DriftUsed, ClaimedRatio: claimed,
		RemainingBytes: remainingAfter, TotalBytes: final.After.Total,
		UploadAfter: &final.After.Upload, DownloadAfter: &final.After.Download,
		DurationSeconds: roundTo(final.Elapsed, 2),
	}
	if remark != "" {
		result.Config = &remark
	}
	if claimed != nil {
		result.ClaimedMatches = &matches
		result.Verdict = verdict
	}
	if opts.JSONOut != "" {
		payload, err := json.MarshalIndent(result, "", "  ")
		if err != nil {
			return nil, ratioErrorf("Could not encode the result: %v", err)
		}
		if err := os.WriteFile(opts.JSONOut, payload, 0o600); err != nil {
			return nil, ratioErrorf("Could not save the result: %v", err)
		}
		fmt.Println(C("JSON result saved to "+opts.JSONOut+".", "green"))
	}
	return result, nil
}

func absFloat(value float64) float64 {
	if value < 0 {
		return -value
	}
	return value
}

func roundTo(value float64, places int) float64 {
	scale := 1.0
	for i := 0; i < places; i++ {
		scale *= 10
	}
	return float64(int64(value*scale+0.5)) / scale
}

func expandUser(path string) string {
	if path == "~" {
		if home, err := os.UserHomeDir(); err == nil && home != "" {
			return home
		}
	}
	if strings.HasPrefix(path, "~/") {
		if home, err := os.UserHomeDir(); err == nil && home != "" {
			return filepath.Join(home, strings.TrimPrefix(path, "~/"))
		}
	}
	return path
}
