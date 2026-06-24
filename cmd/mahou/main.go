package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"syscall"

	"github.com/yashikota/mahou/mahou"
	"github.com/yashikota/mahou/runtimebundle"
)

type commonOptions struct {
	quality         int
	strip           bool
	autoOrient      bool
	format          string
	json            bool
	verbose         bool
	unsafeEnablePDF bool
	policy          string
}

type appContext struct {
	bundle    *runtimebundle.Bundle
	lib       *mahou.Library
	configDir string
	closeOnce sync.Once
}

var (
	cleanupMu sync.Mutex
	cleanups  []func()
)

func registerCleanup(f func()) {
	cleanupMu.Lock()
	cleanups = append(cleanups, f)
	cleanupMu.Unlock()
}

func runCleanups() {
	cleanupMu.Lock()
	defer cleanupMu.Unlock()
	for _, f := range cleanups {
		f()
	}
	cleanups = nil
}

func main() {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-sigChan
		runCleanups()
		os.Exit(1)
	}()

	err := run(os.Args[1:])
	runCleanups()
	if err != nil {
		fmt.Fprintln(os.Stderr, "mahou:", err)
		os.Exit(1)
	}
}

func run(args []string) error {
	if len(args) == 0 {
		usage(os.Stderr)
		return errors.New("command is required")
	}
	switch args[0] {
	case "doctor":
		return runDoctor(args[1:])
	case "formats":
		return runFormats(args[1:])
	case "identify":
		if hasUnrecognizedFlags(args[1:], identifyAllowedFlags) {
			return runDirectMagick(args)
		}
		return runIdentify(args[1:])
	case "convert":
		if hasUnrecognizedFlags(args[1:], convertAllowedFlags) {
			return runDirectMagick(args)
		}
		return runConvert(args[1:])
	case "resize":
		return runResize(args[1:])
	case "exec":
		return runExec(args[1:])
	case "help", "-h", "--help":
		usage(os.Stdout)
		return nil
	default:
		// Fallback directly to the bundled ImageMagick command line!
		return runDirectMagick(args)
	}
}

func usage(out *os.File) {
	fmt.Fprintln(out, `Usage:
  mahou doctor [--verbose] [--json]
  mahou formats [--json]
  mahou identify [options] input.png
  mahou convert [options] input.heic output.webp
  mahou resize [options] input.jpg output.webp --width 1200
  mahou exec [options] [args...]

Options:
  --quality N
  --strip
  --auto-orient
  --format FORMAT
  --json
  --verbose
  --unsafe-enable-pdf
  --policy safe|permissive`)
}

func addCommonFlags(fs *flag.FlagSet, opts *commonOptions) {
	fs.IntVar(&opts.quality, "quality", 0, "output compression quality")
	fs.BoolVar(&opts.strip, "strip", false, "strip metadata")
	fs.BoolVar(&opts.autoOrient, "auto-orient", false, "apply EXIF orientation")
	fs.StringVar(&opts.format, "format", "", "force output format")
	fs.BoolVar(&opts.json, "json", false, "print JSON")
	fs.BoolVar(&opts.verbose, "verbose", false, "print verbose diagnostics")
	fs.BoolVar(&opts.unsafeEnablePDF, "unsafe-enable-pdf", false, "enable PDF/PS/EPS processing for this run")
	fs.StringVar(&opts.policy, "policy", "safe", "security policy: safe or permissive")
}

func parseFlags(fs *flag.FlagSet, args []string) error {
	return fs.Parse(normalizeFlagOrder(args))
}

func normalizeFlagOrder(args []string) []string {
	valueFlags := map[string]bool{
		"quality": true,
		"format":  true,
		"policy":  true,
		"width":   true,
	}
	var flags []string
	var positionals []string
	for i := 0; i < len(args); i++ {
		arg := args[i]
		if arg == "--" {
			positionals = append(positionals, args[i+1:]...)
			break
		}
		if len(arg) > 1 && arg[0] == '-' {
			flags = append(flags, arg)
			name := arg
			for len(name) > 0 && name[0] == '-' {
				name = name[1:]
			}
			if eq := indexByte(name, '='); eq >= 0 {
				name = name[:eq]
			}
			if valueFlags[name] && indexByte(arg, '=') < 0 && i+1 < len(args) {
				i++
				flags = append(flags, args[i])
			}
			continue
		}
		positionals = append(positionals, arg)
	}
	return append(flags, positionals...)
}

func indexByte(s string, b byte) int {
	for i := 0; i < len(s); i++ {
		if s[i] == b {
			return i
		}
	}
	return -1
}

func initialize(opts commonOptions) (*appContext, error) {
	if opts.policy != "safe" && opts.policy != "permissive" {
		return nil, fmt.Errorf("--policy must be safe or permissive")
	}
	if opts.policy == "permissive" {
		opts.unsafeEnablePDF = true
	}
	bundle, err := runtimebundle.Ensure()
	if err != nil {
		return nil, err
	}
	configDir, err := runtimebundle.ApplyPolicy(opts.unsafeEnablePDF)
	if err != nil {
		return nil, err
	}
	runtimebundle.ConfigureEnvironment(bundle.Root, configDir)
	lib, err := mahou.Load(bundle.Root)
	if err != nil {
		_ = os.RemoveAll(configDir)
		return nil, err
	}
	ctx := &appContext{bundle: bundle, lib: lib, configDir: configDir}
	registerCleanup(ctx.Close)
	return ctx, nil
}

func (c *appContext) Close() {
	if c != nil {
		c.closeOnce.Do(func() {
			if c.configDir != "" {
				_ = os.RemoveAll(c.configDir)
			}
		})
	}
}

func printJSON(v any) error {
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	return enc.Encode(v)
}

func targetString() string {
	return runtime.GOOS + "/" + runtime.GOARCH
}

var convertAllowedFlags = map[string]bool{
	"quality":           true,
	"strip":             true,
	"auto-orient":       true,
	"format":            true,
	"policy":            true,
	"unsafe-enable-pdf": true,
	"json":              true,
	"verbose":           true,
}

var identifyAllowedFlags = map[string]bool{
	"policy":            true,
	"unsafe-enable-pdf": true,
	"json":              true,
	"verbose":           true,
}

func hasUnrecognizedFlags(args []string, allowed map[string]bool) bool {
	for i := 0; i < len(args); i++ {
		arg := args[i]
		if strings.HasPrefix(arg, "-") {
			name := strings.TrimLeft(arg, "-")
			if eq := strings.IndexByte(name, '='); eq >= 0 {
				name = name[:eq]
			}
			if !allowed[name] {
				return true
			}
			if name == "quality" || name == "format" || name == "policy" || name == "width" {
				if i+1 < len(args) && !strings.HasPrefix(args[i+1], "-") {
					i++
				}
			}
		}
	}
	return false
}

func runDirectMagick(args []string) error {
	opts, cleanArgs, err := extractCommonOptions(args)
	if err != nil {
		return err
	}
	ctx, err := initialize(opts)
	if err != nil {
		return err
	}
	defer ctx.Close()

	magickCmdPath := filepath.Join(ctx.bundle.Root, "bin", "magick")
	cmd := exec.Command(magickCmdPath, cleanArgs...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func extractCommonOptions(args []string) (commonOptions, []string, error) {
	opts := commonOptions{
		policy: "safe",
	}
	var cleanArgs []string
	for i := 0; i < len(args); i++ {
		arg := args[i]
		if arg == "--policy" && i+1 < len(args) {
			opts.policy = args[i+1]
			i++
			continue
		} else if strings.HasPrefix(arg, "--policy=") {
			opts.policy = strings.TrimPrefix(arg, "--policy=")
			continue
		} else if arg == "--unsafe-enable-pdf" {
			opts.unsafeEnablePDF = true
			continue
		}
		cleanArgs = append(cleanArgs, arg)
	}
	return opts, cleanArgs, nil
}
