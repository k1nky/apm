package main

import (
	"github.com/alecthomas/kong"
)

var BuildVersion = "unknown"
var BuildTarget = "unknown"

func main() {
	ctx := kong.Parse(&CLI)
	err := ctx.Run(&Context{
		Debug:        CLI.Debug,
		WorkDir:      expandPath(CLI.WorkDir),
		File:         CLI.File,
		UseGitConfig: CLI.UseGitConfig,
	})
	ctx.FatalIfErrorf(err)
}
