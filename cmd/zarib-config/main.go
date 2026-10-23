package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"syscall"

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

func runCLI(args []string) int {
	fs := flag.NewFlagSet("zarib-config", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)
	defaults := ratio.DefaultOptions()
	var (
		sub          = fs.String("sub", "", "Subscription URL")
		subFile      = fs.String("sub-file", "", "Read the subscription URL from a file")
		config       = fs.String("config", "", "Config link or JSON file path")
		size         = fs.Float64("size", ratio.DefaultSizeMB, "Test size in MB")
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
	opts.JSONOut = *jsonOut
	opts.KeepTemp = *keepTemp
	if provided["size"] {
		value := *size
		opts.SizeMB = &value
	}
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
