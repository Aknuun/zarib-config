package ratio

import (
	"bufio"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

const (
	Version             = "1.2.7"
	UA                  = "v2rayNG/1.9.16"
	MinSizeMB           = 2.0
	MaxSizeMB           = 10.0
	DefaultSizeMB       = 5.0
	DefaultWait         = 2
	DefaultPollTimeout  = 10
	DefaultDriftWindow  = 5
	DefaultChunkMB      = 5.0
	DefaultTestURL      = "https://speed.cloudflare.com/__down?bytes={bytes}"
	DefaultProbeURL     = "https://speed.cloudflare.com/__down?bytes=4096"
	DefaultProbeTimeout = 20
	DefaultProbeRetries = 0
	DefaultProbeWait    = 10
	LegacyXrayVersion   = "v1.8.24"
	MaxSubBytes         = 4 * 1024 * 1024
)

var CoreChoices = []string{"auto", "latest", "legacy", "system"}

var ProbeURLs = []string{
	DefaultProbeURL,
	"https://www.google.com/generate_204",
	"http://cp.cloudflare.com/generate_204",
}

var ProxyProtocols = map[string]bool{
	"vless": true, "vmess": true, "trojan": true, "shadowsocks": true,
	"wireguard": true, "socks": true, "http": true,
}

var DirectProtocols = map[string]bool{"freedom": true, "blackhole": true, "dns": true}

type RatioError struct {
	Message string
}

func (e *RatioError) Error() string { return e.Message }

type ProbeFailure struct {
	Message string
}

func (e *ProbeFailure) Error() string { return e.Message }

func ratioErrorf(format string, args ...any) error {
	return &RatioError{Message: fmt.Sprintf(format, args...)}
}

func probeFailuref(format string, args ...any) error {
	return &ProbeFailure{Message: fmt.Sprintf(format, args...)}
}

func NormalizeDigits(value string) string {
	replacer := strings.NewReplacer(
		"۰", "0", "۱", "1", "۲", "2", "۳", "3", "۴", "4",
		"۵", "5", "۶", "6", "۷", "7", "۸", "8", "۹", "9",
		"٠", "0", "١", "1", "٢", "2", "٣", "3", "٤", "4",
		"٥", "5", "٦", "6", "٧", "7", "٨", "8", "٩", "9",
	)
	return replacer.Replace(value)
}

func Human(n int64) string {
	return HumanFloat(float64(n))
}

func HumanFloat(n float64) string {
	units := []string{"B", "KB", "MB", "GB", "TB", "PB"}
	idx := 0
	for math.Abs(n) >= 1024 && idx < len(units)-1 {
		n /= 1024
		idx++
	}
	if idx == 0 {
		return fmt.Sprintf("%.0f B", n)
	}
	return fmt.Sprintf("%.2f %s", n, units[idx])
}

var colorEnabled = func() bool {
	info, err := os.Stdout.Stat()
	if err != nil {
		return false
	}
	return info.Mode()&os.ModeCharDevice != 0 && os.Getenv("NO_COLOR") == ""
}()

var colorCodes = map[string]int{
	"red": 31, "green": 32, "yellow": 33, "blue": 34, "magenta": 35,
	"cyan": 36, "bold": 1, "dim": 2,
}

func C(text string, names ...string) string {
	if !colorEnabled {
		return text
	}
	var prefix strings.Builder
	for _, name := range names {
		if code, ok := colorCodes[name]; ok {
			fmt.Fprintf(&prefix, "\033[%dm", code)
		}
	}
	if prefix.Len() == 0 {
		return text
	}
	return prefix.String() + text + "\033[0m"
}

var bigFont = map[rune][]string{
	'0': {"###", "# #", "# #", "# #", "###"},
	'1': {"  #", "  #", "  #", "  #", "  #"},
	'2': {"###", "  #", "###", "#  ", "###"},
	'3': {"###", "  #", "###", "  #", "###"},
	'4': {"# #", "# #", "###", "  #", "  #"},
	'5': {"###", "#  ", "###", "  #", "###"},
	'6': {"###", "#  ", "###", "# #", "###"},
	'7': {"###", "  #", "  #", "  #", "  #"},
	'8': {"###", "# #", "###", "# #", "###"},
	'9': {"###", "# #", "###", "  #", "###"},
	'.': {"   ", "   ", "   ", "   ", " # "},
	'x': {"   ", "# #", " # ", "# #", "   "},
}

func BigText(text string) []string {
	var glyphs [][]string
	for _, ch := range text {
		if glyph, ok := bigFont[ch]; ok {
			glyphs = append(glyphs, glyph)
		}
	}
	if len(glyphs) == 0 {
		return nil
	}
	lines := make([]string, 5)
	for row := 0; row < 5; row++ {
		var cells []string
		for _, glyph := range glyphs {
			var line strings.Builder
			for _, cell := range glyph[row] {
				if cell == '#' {
					line.WriteString("██")
				} else {
					line.WriteString("  ")
				}
			}
			cells = append(cells, line.String())
		}
		lines[row] = strings.Join(cells, "  ")
	}
	return lines
}

func Countdown(seconds int, message string) {
	if seconds <= 0 {
		return
	}
	if !colorEnabled {
		fmt.Printf("%s (%ds) ...\n", message, seconds)
		time.Sleep(time.Duration(seconds) * time.Second)
		return
	}
	frames := "|/-\\"
	const width = 20
	for remaining := seconds; remaining > 0; remaining-- {
		filled := int(float64(seconds-remaining+1) / float64(seconds) * width)
		bar := strings.Repeat("█", filled) + strings.Repeat("░", width-filled)
		fmt.Printf("\r\033[K%s  %s  %s %s", C(message, "cyan"), C(bar, "green"),
			C(fmt.Sprintf("%2ds", remaining), "bold"), C(string(frames[remaining%4]), "cyan"))
		time.Sleep(time.Second)
	}
	fmt.Printf("\r\033[K%s\n", C(message+" ... done", "green"))
}

var stdinReader = bufio.NewReader(os.Stdin)

func Ask(prompt string, defaultValue *string) (string, error) {
	suffix := ""
	if defaultValue != nil && *defaultValue != "" {
		suffix = fmt.Sprintf(" [%s]", *defaultValue)
	}
	fmt.Printf("%s%s: ", prompt, suffix)
	value, err := stdinReader.ReadString('\n')
	if err != nil && value == "" {
		return "", ratioErrorf("Input closed.")
	}
	value = strings.TrimSpace(value)
	if value == "" && defaultValue != nil {
		return *defaultValue, nil
	}
	return value, nil
}

func ParseBool(value string) bool {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "1", "true", "yes", "on":
		return true
	default:
		return false
	}
}

func IsInteractive() bool {
	info, err := os.Stdin.Stat()
	if err != nil || info.Mode()&os.ModeCharDevice == 0 {
		return false
	}
	nullInfo, err := os.Stat(os.DevNull)
	if err == nil && os.SameFile(info, nullInfo) {
		return false
	}
	return true
}

func AskYesNo(prompt string, defaultValue bool) bool {
	if !IsInteractive() {
		return defaultValue
	}
	def := "n"
	if defaultValue {
		def = "y"
	}
	raw, err := Ask(prompt+" [y/n]", &def)
	if err != nil {
		return defaultValue
	}
	return ParseBool(raw)
}

func CacheDir() string {
	if value := os.Getenv("RATIO_TEST_CACHE"); value != "" {
		return value
	}
	home, err := os.UserHomeDir()
	if err != nil || home == "" {
		home = "."
	}
	return filepath.Join(home, ".cache", "ratio-test")
}

func ParseFloat(value string) (float64, error) {
	return strconv.ParseFloat(NormalizeDigits(strings.TrimSpace(value)), 64)
}
