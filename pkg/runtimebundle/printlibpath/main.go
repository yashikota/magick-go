package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/yashikota/magick-go/pkg/runtimebundle"
)

func main() {
	bundle, err := runtimebundle.Ensure()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	fmt.Print(filepath.Join(bundle.Root, "lib"))
}
