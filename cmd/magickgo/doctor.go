package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sort"
	"strings"

	"github.com/yashikota/magick-go/internal/magick"
	"github.com/yashikota/magick-go/internal/runtimebundle"
)

type doctorReport struct {
	Target              string            `json:"target"`
	RuntimeRoot         string            `json:"runtime_root"`
	RuntimeHash         string            `json:"runtime_hash"`
	LibraryPath         string            `json:"library_path"`
	LoadOK              bool              `json:"load_ok"`
	Version             string            `json:"version"`
	QuantumDepth        string            `json:"quantum_depth"`
	HDRI                string            `json:"hdri"`
	Environment         map[string]string `json:"environment"`
	Configure           map[string]string `json:"configure"`
	FormatSupport       map[string]bool   `json:"format_support"`
	Formats             []string          `json:"formats,omitempty"`
	MissingLibraryNotes []string          `json:"missing_library_notes,omitempty"`
}

func runDoctor(args []string) error {
	var opts commonOptions
	fs := flag.NewFlagSet("doctor", flag.ContinueOnError)
	addCommonFlags(fs, &opts)
	if err := parseFlags(fs, args); err != nil {
		return err
	}
	ctx, err := initialize(opts)
	if err != nil {
		return err
	}
	defer ctx.Close()
	diag := magick.DiagnosticsInfo()
	formats := diag.Formats
	support := diag.Support
	if len(formats) == 0 {
		formats = cliFormats(ctx.bundle.Root)
		support = formatSupport(formats)
	}
	report := doctorReport{
		Target:              targetString(),
		RuntimeRoot:         ctx.bundle.Root,
		RuntimeHash:         ctx.bundle.Hash,
		LibraryPath:         ctx.lib.Path(),
		LoadOK:              true,
		Version:             diag.Version,
		QuantumDepth:        diag.QuantumDepth,
		HDRI:                diag.HDRI,
		Environment:         runtimebundle.Environment(ctx.bundle.Root, ctx.configDir),
		Configure:           diag.Configure,
		FormatSupport:       support,
		MissingLibraryNotes: missingLibraryNotes(ctx.bundle.Root),
	}
	if opts.verbose {
		report.Formats = formats
	}
	if opts.json {
		return printJSON(report)
	}
	printDoctor(report, opts.verbose)
	return nil
}

func formatSupport(formats []string) map[string]bool {
	set := make(map[string]struct{}, len(formats))
	for _, f := range formats {
		set[f] = struct{}{}
	}
	support := make(map[string]bool)
	for _, name := range []string{"JPEG", "PNG", "WEBP", "TIFF", "HEIC", "JXL", "SVG", "PDF"} {
		_, ok := set[name]
		support[name] = ok
	}
	return support
}

func cliFormats(root string) []string {
	cmd := exec.Command(filepath.Join(root, "bin", "magick"), "-list", "format")
	out, err := cmd.Output()
	if err != nil {
		return nil
	}
	set := make(map[string]struct{})
	scanner := bufio.NewScanner(strings.NewReader(string(out)))
	for scanner.Scan() {
		for _, field := range strings.Fields(scanner.Text()) {
			name := strings.TrimRight(field, "*:")
			if name == strings.ToUpper(name) {
				set[name] = struct{}{}
			}
		}
	}
	formats := make([]string, 0, len(set))
	for name := range set {
		formats = append(formats, name)
	}
	sort.Strings(formats)
	return formats
}

func printDoctor(r doctorReport, verbose bool) {
	fmt.Fprintln(os.Stdout, "target:", r.Target)
	fmt.Fprintln(os.Stdout, "runtime:", r.RuntimeRoot)
	fmt.Fprintln(os.Stdout, "runtime hash:", r.RuntimeHash)
	fmt.Fprintln(os.Stdout, "libMagickWand:", r.LibraryPath)
	fmt.Fprintln(os.Stdout, "load:", okText(r.LoadOK))
	fmt.Fprintln(os.Stdout, "version:", r.Version)
	fmt.Fprintln(os.Stdout, "quantum depth:", r.QuantumDepth)
	fmt.Fprintln(os.Stdout, "HDRI:", r.HDRI)
	fmt.Fprintln(os.Stdout, "configure path:", r.Environment["MAGICK_CONFIGURE_PATH"])
	fmt.Fprintln(os.Stdout, "coder module path:", r.Environment["MAGICK_CODER_MODULE_PATH"])
	fmt.Fprintln(os.Stdout, "delegates:", r.Configure["DELEGATES"])
	if len(r.MissingLibraryNotes) == 0 {
		fmt.Fprintln(os.Stdout, "missing libraries: none detected")
	} else {
		fmt.Fprintln(os.Stdout, "missing libraries:")
		for _, note := range r.MissingLibraryNotes {
			fmt.Fprintln(os.Stdout, " ", note)
		}
	}
	for _, name := range []string{"JPEG", "PNG", "WEBP", "TIFF", "HEIC", "JXL", "SVG", "PDF"} {
		fmt.Fprintf(os.Stdout, "%s: %s\n", name, okText(r.FormatSupport[name]))
	}
	if verbose {
		fmt.Fprintln(os.Stdout, "formats:")
		for _, f := range r.Formats {
			fmt.Fprintln(os.Stdout, " ", f)
		}
	}
}

func okText(ok bool) string {
	if ok {
		return "ok"
	}
	return "missing"
}

func missingLibraryNotes(root string) []string {
	files := diagnosticLibraryFiles(root)
	notes := make([]string, 0)
	for _, file := range files {
		switch runtime.GOOS {
		case "linux":
			out, err := exec.Command("ldd", file).CombinedOutput()
			if err != nil {
				continue
			}
			for _, line := range strings.Split(string(out), "\n") {
				if strings.Contains(line, "not found") {
					notes = append(notes, filepath.Base(file)+": "+strings.TrimSpace(line))
				}
			}
		case "darwin":
			out, err := exec.Command("otool", "-L", file).CombinedOutput()
			if err != nil {
				continue
			}
			for _, line := range strings.Split(string(out), "\n") {
				line = strings.TrimSpace(line)
				if strings.HasPrefix(line, "/opt/homebrew/") || strings.HasPrefix(line, "/usr/local/") {
					notes = append(notes, filepath.Base(file)+": absolute Homebrew dependency "+strings.Fields(line)[0])
				}
			}
		}
	}
	return notes
}

func diagnosticLibraryFiles(root string) []string {
	patterns := []string{
		filepath.Join(root, "bin", "magick"),
		filepath.Join(root, "bin", "gs"),
		filepath.Join(root, "lib", "libMagickWand*"),
		filepath.Join(root, "lib", "libMagickCore*"),
		filepath.Join(root, "lib", "ImageMagick", "modules-*", "coders", "*.so"),
		filepath.Join(root, "lib", "ImageMagick", "modules-*", "filters", "*.so"),
		filepath.Join(root, "lib", "ImageMagick", "modules-*", "coders", "*.dylib"),
		filepath.Join(root, "lib", "ImageMagick", "modules-*", "filters", "*.dylib"),
		filepath.Join(root, "lib", "ImageMagick-*", "modules-*", "coders", "*.so"),
		filepath.Join(root, "lib", "ImageMagick-*", "modules-*", "filters", "*.so"),
		filepath.Join(root, "lib", "ImageMagick-*", "modules-*", "coders", "*.dylib"),
		filepath.Join(root, "lib", "ImageMagick-*", "modules-*", "filters", "*.dylib"),
	}
	var files []string
	for _, pattern := range patterns {
		matches, err := filepath.Glob(pattern)
		if err == nil {
			files = append(files, matches...)
		}
	}
	return files
}
