// Package web embeds the compiled frontend.
//
// The frontend lives in src/frontend and its Vite build writes into
// src/backend/web/dist (see vite.config.ts). //go:embed cannot reference
// parent directories, so the build output has to land inside this package.
// Run `make frontend` (or `npm run build` in src/frontend) before building Go.
package web

import (
	"embed"
	"io/fs"
)

//go:embed dist
var dist embed.FS

// Dist returns the frontend build output rooted at the dist directory.
func Dist() fs.FS {
	sub, err := fs.Sub(dist, "dist")
	if err != nil {
		// fs.Sub only fails for invalid paths; "dist" is a constant.
		panic(err)
	}
	return sub
}
