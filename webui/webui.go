package webui

import (
	"embed"
	"io/fs"
)

//go:embed dist/*
var assets embed.FS

func Assets() fs.FS {
	fsys, err := fs.Sub(assets, "dist")
	if err != nil {
		panic(err)
	}
	return fsys
}
