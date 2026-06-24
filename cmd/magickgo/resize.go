package main

import (
	"flag"
	"fmt"

	"github.com/yashikota/magick-go/pkg/magick"
)

func runResize(args []string) error {
	var opts commonOptions
	var width uint64
	fs := flag.NewFlagSet("resize", flag.ContinueOnError)
	addCommonFlags(fs, &opts)
	fs.Uint64Var(&width, "width", 0, "target width")
	if err := parseFlags(fs, args); err != nil {
		return err
	}
	if fs.NArg() != 2 {
		return fmt.Errorf("resize requires input and output paths")
	}
	if width == 0 {
		return fmt.Errorf("resize requires --width")
	}
	ctx, err := initialize(opts)
	if err != nil {
		return err
	}
	defer ctx.Close()
	return magick.Resize(fs.Arg(0), fs.Arg(1), width, convertOptions(opts))
}
