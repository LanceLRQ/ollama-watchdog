//go:build !dev
// +build !dev

package server

import (
	"embed"
)

//go:embed web/*
var WebsiteAssetsEmbed embed.FS
