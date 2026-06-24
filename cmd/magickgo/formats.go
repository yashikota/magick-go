package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/yashikota/magick-go/pkg/magick"
)

func runFormats(args []string) error {
	var opts commonOptions
	fs := flag.NewFlagSet("formats", flag.ContinueOnError)
	addCommonFlags(fs, &opts)
	if err := parseFlags(fs, args); err != nil {
		return err
	}
	ctx, err := initialize(opts)
	if err != nil {
		return err
	}
	defer ctx.Close()
	formats := magick.Formats()
	if opts.json {
		return printJSON(formats)
	}
	for _, f := range formats {
		fmt.Fprintln(os.Stdout, f)
	}
	return nil
}
