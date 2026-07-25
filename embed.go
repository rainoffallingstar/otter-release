package main

import (
	"embed"

	"github.com/rainoffallingstar/otter/internal/assets"
)

//go:embed inst/Rscripts/* inst/snakefiles/* inst/rules/* inst/rules_legacy/* inst/envs/* inst/data/*
var embeddedAssets embed.FS

func init() {
	assets.SetEmbeddedAssets(embeddedAssets)
}
