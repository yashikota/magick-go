package main

import (
	"flag"
	"fmt"

	"github.com/yashikota/magick-go/internal/magick"
)

func runConvert(args []string) error {
	var opts commonOptions
	fs := flag.NewFlagSet("convert", flag.ContinueOnError)
	addCommonFlags(fs, &opts)
	if err := parseFlags(fs, args); err != nil {
		return err
	}
	if fs.NArg() != 2 {
		return fmt.Errorf("convert requires input and output paths")
	}
	ctx, err := initialize(opts)
	if err != nil {
		return err
	}
	defer ctx.Close()
	return magick.Convert(fs.Arg(0), fs.Arg(1), convertOptions(opts))
}

func convertOptions(opts commonOptions) magick.ConvertOptions {
	return magick.ConvertOptions{
		Quality:    opts.quality,
		Strip:      opts.strip,
		AutoOrient: opts.autoOrient,
		Format:     opts.format,
	}
}
