package main

import (
	"fmt"
	"strings"

	"github.com/k1nky/apm/internal/downloader"
	"github.com/k1nky/apm/internal/manager"
	"github.com/sirupsen/logrus"
)

type Context struct {
	Debug   bool
	WorkDir string
}

type InstallCmd struct {
	Url      string            `kong:"help='Package URL, will be skipped when installation from file is set. Default version is <master>',name='url',short='u',arg,placeholder='url[@version]'"`
	Path     string            `kong:"help='Path to .apkg in the remote repository',name='path',short='p',default='.',optional"`
	Mappings map[string]string `kong:"help='Package files mappings, will mount a source file or directory within a destination directory. Default, all source files will be mounted winthin the working directory',name='mappings',short='m',default='\"*=.\"',optional"`

	// From      string `kong:"help='Path to a file with requirements',name='from',short='f',optional"`
	// Boost     bool
	// Save bool
}

type ListCmd struct {
	Url string `kong:"help='Package URL',name='url',short='u',arg,placeholder='url'"`
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
	p := &manager.Package{
		Mappings: make([]manager.Mapping, 0),
	}

	p.URL, p.Version = parseUrl(cmd.Url)
	p.Path = cmd.Path
	for k, v := range cmd.Mappings {
		p.Mappings = append(p.Mappings, manager.Mapping{Src: k, Dest: v})
	}

	if err := m.Install(p, &manager.InstallOptions{WorkDir: ctx.WorkDir}); err != nil {
		logrus.Error(err)
		return err
	}
	return nil
}

func (cmd *ListCmd) Run(ctx *Context) (err error) {
	var versions []string
	d := downloader.NewDownloader()
	versions, err = d.FetchVersion(cmd.Url, nil)
	for _, v := range versions {
		fmt.Println(v)
	}

	return
}

var CLI struct {
	Debug   bool   `help:"Enable debug mode."`
	WorkDir string `kong:"help='Working directory with .apm mount point. It is current directory by default',name='workdir',short='w',optional"`
	// User         string
	// AuthType     string
	// UseGitConfig bool
	Install InstallCmd `kong:"cmd,help='Install a package'"`
	List    ListCmd    `kong:"cmd,help='List remote versions'"`
}
