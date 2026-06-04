package main

import "embed"

//go:embed all:assets/static all:assets/templates
var assets embed.FS
