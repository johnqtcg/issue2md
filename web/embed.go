package webassets

import "embed"

// FS contains web templates and static assets embedded at build time.
//
//go:embed templates/index.html static/* swaggerui/*
var FS embed.FS
