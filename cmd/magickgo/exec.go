package main

import (
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

func runExec(args []string) error {
	var magickgoFlags []string
	var magickArgs []string
	dashDashIdx := -1
	for i, arg := range args {
		if arg == "--" {
			dashDashIdx = i
			break
		}
	}
	if dashDashIdx >= 0 {
		magickgoFlags = args[:dashDashIdx]
		magickArgs = args[dashDashIdx+1:]
	} else {
		magickArgs = args
	}

	var opts commonOptions
	fs := flag.NewFlagSet("exec", flag.ContinueOnError)
	addCommonFlags(fs, &opts)
	if len(magickgoFlags) > 0 {
		if err := fs.Parse(magickgoFlags); err != nil {
			return err
		}
	}

	ctx, err := initialize(opts)
	if err != nil {
		return err
	}
	defer ctx.Close()

	if len(magickArgs) == 0 {
		return fmt.Errorf("exec requires at least one argument for the bundled magick command, e.g. magickgo exec input.png -resize 50%% output.png or magickgo exec --policy permissive -- input.pdf output.png")
	}

	magickCmdPath := filepath.Join(ctx.bundle.Root, "bin", "magick")
	cmd := exec.Command(magickCmdPath, magickArgs...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}
