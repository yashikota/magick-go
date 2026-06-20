package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"os"
	"runtime"

	"github.com/yashikota/magick-go/internal/magick"
	"github.com/yashikota/magick-go/internal/runtimebundle"
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
	lib       *magick.Library
	configDir string
}

func main() {
	if err := run(os.Args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, "magickgo:", err)
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
		return runIdentify(args[1:])
	case "convert":
		return runConvert(args[1:])
	case "resize":
		return runResize(args[1:])
	case "help", "-h", "--help":
		usage(os.Stdout)
		return nil
	default:
		usage(os.Stderr)
		return fmt.Errorf("unknown command %q", args[0])
	}
}

func usage(out *os.File) {
	fmt.Fprintln(out, `Usage:
  magickgo doctor [--verbose] [--json]
  magickgo formats [--json]
  magickgo identify [options] input.png
  magickgo convert [options] input.heic output.webp
  magickgo resize [options] input.jpg output.webp --width 1200

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
	lib, err := magick.Load(bundle.Root)
	if err != nil {
		_ = os.RemoveAll(configDir)
		return nil, err
	}
	return &appContext{bundle: bundle, lib: lib, configDir: configDir}, nil
}

func (c *appContext) Close() {
	if c != nil && c.configDir != "" {
		_ = os.RemoveAll(c.configDir)
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
