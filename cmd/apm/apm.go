package main

import (
	apmlog "github.com/k1nky/apm/internal/log"

	"github.com/alecthomas/kong"
	"github.com/sirupsen/logrus"
)

func init() {
	logrus.SetFormatter(&apmlog.Formatter{
		LogFormat: "[%lvl%]: %time% - %msg% %fields%\n",
	})
}

func main() {
	ctx := kong.Parse(&CLI)
	// Call the Run() method of the selected parsed command.
	err := ctx.Run(&Context{
		Debug:   CLI.Debug,
		WorkDir: CLI.WorkDir})
	ctx.FatalIfErrorf(err)

}
