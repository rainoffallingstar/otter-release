package main

import (
	"embed"

	"github.com/xdxtools/xdxtools-go/internal/assets"
)

//go:embed inst/Rscripts/* inst/snakefiles/* inst/rootless_rules/* inst/envs/*
var embeddedAssets embed.FS

func init() {
	assets.SetEmbeddedAssets(embeddedAssets)
}
