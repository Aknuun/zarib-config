package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/Aknuun/zarib-config/internal/ratio"
	"github.com/Aknuun/zarib-config/internal/web"
)

func main() {
	if len(os.Args) > 1 && os.Args[1] == "serve" {
		serve(os.Args[2:])
		return
	}
	os.Exit(runCLI(os.Args[1:]))
}

func serve(args []string) {
	fs := flag.NewFlagSet("serve", flag.ExitOnError)
	addr := fs.String("addr", "127.0.0.1:8000", "listen address")
	staticDir := fs.String("static-dir", "web/static", "static assets directory")
	_ = fs.Parse(args)
	if err := web.Run(*addr, *staticDir); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func runMenu() int {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	for {
		fmt.Println()
		fmt.Println(ratio.C("zarib-config "+ratio.Version, "cyan", "bold"))
		fmt.Println("1) Test a subscription")
		fmt.Println("2) List configs")
		fmt.Println("3) Check connectivity")
		fmt.Println("4) Exit")
		def := "1"
		raw, err := ratio.Ask("Choose", &def)
		if err != nil {
			return 0
		}
		switch strings.ToLower(strings.TrimSpace(raw)) {
		case "1":
			menuTest(ctx)
			if ctx.Err() != nil {
				return 0
			}
		case "2":
			menuList()
		case "3":
			menuProbe(ctx)
			if ctx.Err() != nil {
				return 0
			}
		case "4", "q", "quit", "exit":
			return 0
		default:
			fmt.Println("Invalid choice.")
		}
	}
}

func askSub() (string, error) {
	for {
		raw, err := ratio.Ask("Subscription URL", nil)
		if err != nil {
			return "", err
		}
		raw = strings.TrimSpace(raw)
		if raw != "" {
			return raw, nil
		}
		fmt.Println("Subscription URL is required.")
	}
}

func loadAndPrintConfigs(sub string) ([]ratio.Config, error) {
	body, _, err := ratio.Fetch(sub, 30*time.Second)
	if err != nil {
		return nil, err
	}
	configs := ratio.ExtractConfigs(body)
	if len(configs) == 0 {
		return nil, ratio.ErrNoConfigs
	}
	for index, config := range configs {
		remark := config.Remark
		if remark == "" {
			remark = "(unnamed)"
		}
		fmt.Printf("%d) %s | %s | %s\n", index+1, config.Protocol, config.Target, remark)
	}
	return configs, nil
}

func parsePicks(raw string, count int) ([]int, error) {
	raw = strings.ToLower(strings.TrimSpace(raw))
	if raw == "all" || raw == "*" {
		picks := make([]int, count)
		for index := range picks {
			picks[index] = index + 1
		}
		return picks, nil
	}
	for _, prefix := range []string{"first ", "count:", "count "} {
		if strings.HasPrefix(raw, prefix) {
			value, err := strconv.Atoi(ratio.NormalizeDigits(strings.TrimSpace(strings.TrimPrefix(raw, prefix))))
			if err != nil || value < 1 || value > count {
				return nil, fmt.Errorf("count must be between 1 and %d", count)
			}
			picks := make([]int, value)
			for index := range picks {
				picks[index] = index + 1
			}
			return picks, nil
		}
	}
	parts := strings.FieldsFunc(raw, func(r rune) bool {
		return r == ',' || r == ';' || r == ' ' || r == '\t'
	})
	if len(parts) == 0 {
		return nil, fmt.Errorf("enter at least one config number")
	}
	seen := map[int]bool{}
	picks := []int{}
	for _, part := range parts {
		value, err := strconv.Atoi(ratio.NormalizeDigits(part))
		if err != nil || value < 1 || value > count {
			return nil, fmt.Errorf("config number must be between 1 and %d", count)
		}
		if !seen[value] {
			seen[value] = true
			picks = append(picks, value)
		}
	}
	return picks, nil
}

func askSize() float64 {
	def := fmt.Sprintf("%g", ratio.CLIDefaultSizeMB)
	raw, err := ratio.Ask("Test size in MB", &def)
	if err != nil {
		return ratio.CLIDefaultSizeMB
	}
	value, err := ratio.ParseFloat(raw)
	if err != nil || value <= 0 {
		fmt.Printf("Invalid size; using %g MB.\n", ratio.CLIDefaultSizeMB)
		return ratio.CLIDefaultSizeMB
	}
	return value
}

func askCore() string {
	def := "auto"
	raw, err := ratio.Ask("Xray core (auto/latest/legacy/system)", &def)
	if err != nil {
		return "auto"
	}
	raw = strings.ToLower(strings.TrimSpace(raw))
	for _, choice := range ratio.CoreChoices {
		if raw == choice {
			return raw
		}
	}
	fmt.Println("Invalid core; using auto.")
	return "auto"
}

func menuList() {
	sub, err := askSub()
	if err != nil {
		return
	}
	if _, err := loadAndPrintConfigs(sub); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
	}
}

func menuProbe(ctx context.Context) {
	sub, err := askSub()
	if err != nil {
		return
	}
	configs, err := loadAndPrintConfigs(sub)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		return
	}
	def := "1"
	raw, err := ratio.Ask("Config number", &def)
	if err != nil {
		return
	}
	pick, err := strconv.Atoi(ratio.NormalizeDigits(strings.TrimSpace(raw)))
	if err != nil || pick < 1 || pick > len(configs) {
		fmt.Printf("Config number must be between 1 and %d.\n", len(configs))
		return
	}
	opts := ratio.DefaultOptions()
	opts.Sub = sub
	opts.Pick = pick
	opts.ProbeOnly = true
	opts.FreeSize = true
	opts.Core = askCore()
	if _, err := ratio.Run(ctx, opts); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
	}
}

func menuTest(ctx context.Context) {
	sub, err := askSub()
	if err != nil {
		return
	}
	configs, err := loadAndPrintConfigs(sub)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		return
	}
	def := "1"
	raw, err := ratio.Ask("Configs (e.g. 1,2,3 / all / first 3)", &def)
	if err != nil {
		return
	}
	picks, err := parsePicks(raw, len(configs))
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		return
	}
	size := askSize()
	core := askCore()
	background := ratio.AskYesNo("Measure background usage? (slower, more accurate)", false)
	if !ratio.AskYesNo(fmt.Sprintf("Start %d test(s)? Make sure no other device is using this subscription.", len(picks)), true) {
		return
	}
	for _, pick := range picks {
		if ctx.Err() != nil {
			return
		}
		opts := ratio.DefaultOptions()
		opts.Sub = sub
		opts.Pick = pick
		opts.SizeMB = &size
		opts.FreeSize = true
		opts.Core = core
		opts.Sync = background
		opts.DriftWindow = 0
		opts.Wait = 2
		if _, err := ratio.Run(ctx, opts); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			if ctx.Err() != nil {
				return
			}
		}
	}
}

func runCLI(args []string) int {
	fs := flag.NewFlagSet("zarib-config", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)
	defaults := ratio.DefaultOptions()
	var (
		sub          = fs.String("sub", "", "Subscription URL")
		subFile      = fs.String("sub-file", "", "Read the subscription URL from a file")
		config       = fs.String("config", "", "Config link or JSON file path")
		size         = fs.Float64("size", ratio.CLIDefaultSizeMB, "Test size in MB (default 100)")
		wait         = fs.Int("wait", defaults.Wait, "Seconds to wait before reading usage")
		pollTimeout  = fs.Int("poll-timeout", defaults.PollTimeout, "Maximum seconds to wait for usage")
		driftWindow  = fs.Int("drift-window", defaults.DriftWindow, "Seconds to sample background usage")
		pick         = fs.Int("pick", 0, "Config number inside the subscription")
		list         = fs.Bool("list", false, "Only list configs and exit")
		chunk        = fs.Float64("chunk", defaults.ChunkMB, "Download request size in MB")
		timeout      = fs.Float64("timeout", defaults.Timeout, "Maximum seconds per download request")
		testURL      = fs.String("test-url", defaults.TestURL, "Test download URL")
		probeURL     = fs.String("probe-url", defaults.ProbeURL, "Probe URL")
		probeTimeout = fs.Int("probe-timeout", defaults.ProbeTimeout, "Seconds per probe URL")
		probeOnly    = fs.Bool("probe-only", false, "Only check connectivity")
		probeRetries = fs.Int("probe-retries", defaults.ProbeRetries, "Extra connection attempts")
		probeWait    = fs.Int("probe-wait", defaults.ProbeWait, "Seconds between probe attempts")
		syncMode     = fs.Bool("sync", false, "Align the test with the panel's usage update cycle")
		controlFile  = fs.String("control-file", "", "Internal: file used to extend the usage wait")
		core         = fs.String("core", defaults.Core, "Xray core: auto, latest, legacy, system")
		claimed      = fs.Float64("claimed", 0, "Claimed coefficient")
		jsonOut      = fs.String("json-out", "", "Save the result to a JSON file")
		keepTemp     = fs.Bool("keep-temp", false, "Keep temporary files")
		version      = fs.Bool("version", false, "Print the version")
	)
	fs.Usage = func() {
		fmt.Fprintln(fs.Output(), "Measure the real traffic coefficient of a V2Ray config by comparing actual downloads with subscription usage.")
		fmt.Fprintln(fs.Output(), "\nUsage: zarib-config [options]")
		fs.PrintDefaults()
	}
	if err := fs.Parse(args); err != nil {
		return 2
	}
	if *version {
		fmt.Printf("zarib-config %s\n", ratio.Version)
		return 0
	}
	if _, err := exec.LookPath("curl"); err != nil {
		fmt.Fprintln(os.Stderr, "curl is not installed.")
		return 2
	}
	if len(args) == 0 && ratio.IsInteractive() {
		return runMenu()
	}
	provided := map[string]bool{}
	fs.Visit(func(item *flag.Flag) { provided[item.Name] = true })
	opts := defaults
	opts.Sub = *sub
	opts.SubFile = *subFile
	opts.Config = *config
	opts.Wait = *wait
	opts.PollTimeout = *pollTimeout
	opts.DriftWindow = *driftWindow
	opts.List = *list
	opts.ChunkMB = *chunk
	opts.Timeout = *timeout
	opts.TestURL = *testURL
	opts.ProbeURL = *probeURL
	opts.ProbeTimeout = *probeTimeout
	opts.ProbeOnly = *probeOnly
	opts.ProbeRetries = *probeRetries
	opts.ProbeWait = *probeWait
	opts.Sync = *syncMode
	opts.ControlFile = *controlFile
	opts.Core = *core
	opts.FreeSize = true
	opts.JSONOut = *jsonOut
	opts.KeepTemp = *keepTemp
	sizeValue := *size
	opts.SizeMB = &sizeValue
	if provided["pick"] {
		opts.Pick = *pick
		if opts.Pick == 0 {
			opts.Pick = -1
		}
	}
	if provided["claimed"] {
		value := *claimed
		opts.Claimed = &value
	}
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	_, err := ratio.Run(ctx, opts)
	if err == nil {
		return 0
	}
	if errors.Is(err, ratio.ErrNoConfigs) {
		return 2
	}
	if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
		fmt.Fprintln(os.Stderr, ratio.C("\nCancelled.", "yellow"))
		return 130
	}
	fmt.Fprintf(os.Stderr, "%s %v\n", ratio.C("Error:", "red", "bold"), err)
	return 2
}
