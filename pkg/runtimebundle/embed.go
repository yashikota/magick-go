package runtimebundle

import "embed"

// assets contains optional runtime-<target>.tar.zst files produced by CI.
//
//go:embed assets/*
var assets embed.FS
