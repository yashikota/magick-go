package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/yashikota/mahou/mahou"
)

func runIdentify(args []string) error {
	var opts commonOptions
	fs := flag.NewFlagSet("identify", flag.ContinueOnError)
	addCommonFlags(fs, &opts)
	if err := parseFlags(fs, args); err != nil {
		return err
	}
	if fs.NArg() != 1 {
		return fmt.Errorf("identify requires input path")
	}
	ctx, err := initialize(opts)
	if err != nil {
		return err
	}
	defer ctx.Close()
	info, err := mahou.Identify(fs.Arg(0))
	if err != nil {
		return err
	}
	if opts.json {
		return printJSON(info)
	}
	fmt.Fprintf(os.Stdout, "%s %s %dx%d depth=%d\n", info.Path, info.Format, info.Width, info.Height, info.Depth)
	return nil
}
