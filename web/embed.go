package web

import "embed"

// PublicFS embeds every file under public/ (site CSS/JS, images and the static
// admin app) so a compiled binary serves them without the source tree.
//
//go:embed all:public
var PublicFS embed.FS
