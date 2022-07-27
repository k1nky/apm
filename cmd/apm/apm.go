package main

import (
	apmlog "github.com/k1nky/apm/internal/log"

	"github.com/alecthomas/kong"
	"github.com/sirupsen/logrus"
)

var BuildVersion = "unknown"
var BuildTarget = "unknown"

func init() {
	logrus.SetFormatter(&apmlog.Formatter{
		LogFormat: "[%lvl%]: %time% - %msg% %fields%\n",
		OnlyTime:  true,
	})
}

func main() {
	ctx := kong.Parse(&CLI)
	// Call the Run() method of the selected parsed command.
	err := ctx.Run(&Context{
		Debug:        CLI.Debug,
		WorkDir:      expandPath(CLI.WorkDir),
		File:         CLI.File,
		UseGitConfig: CLI.UseGitConfig,
	})
	ctx.FatalIfErrorf(err)
}
