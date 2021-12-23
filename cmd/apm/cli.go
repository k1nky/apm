package main

import (
	"strings"

	"github.com/k1nky/apm/internal/manager"
	"github.com/sirupsen/logrus"
)

type Context struct {
	Debug   bool
	WorkDir string
}

type InstallCmd struct {
	Url  string `kong:"help='Package URL, will be skipped when installation from file is set. Default version is <master>',name='url',short='u',arg,placeholder='url[@version]'"`
	Path string `kong:"help='Path to .apkg in the remote repository',name='path',short='p',default='.',optional"`

	// From      string `kong:"help='Path to a file with requirements',name='from',short='f',optional"`
	// Update    bool
	// CheckMode bool
	// Boost     bool
	// Mappings  map[string]string
}

func parseUrl(s string) (url string, version string) {
	parts := strings.Split(s, "@")
	if len(parts) == 0 {
		return
	}
	url = parts[0]
	if len(parts) > 1 {
		version = parts[1]
	}
	if len(version) == 0 {
		version = "master"
	}
	return
}

func (cmd *InstallCmd) Run(ctx *Context) error {
	m := manager.Manager{}
	url, version := parseUrl(cmd.Url)
	if err := m.Install(&manager.Package{
		URL:     url,
		Version: version,
		Path:    cmd.Path,
	}, &manager.InstallOptions{WorkDir: ctx.WorkDir}); err != nil {
		logrus.Error(err)
		return err
	}
	return nil
}

var CLI struct {
	Debug   bool   `help:"Enable debug mode."`
	WorkDir string `kong:"help='Working directory with .apm mount point. It is current directory by default',name='workdir',short='w',optional"`
	// User         string
	// AuthType     string
	// UseGitConfig bool
	Install InstallCmd `kong:"cmd,help='Install a package'"`
}
