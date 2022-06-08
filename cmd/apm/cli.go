package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/k1nky/apm/internal/downloader"
	"github.com/k1nky/apm/internal/manager"
	"github.com/k1nky/apm/internal/parser"
	"github.com/sirupsen/logrus"
)

type Context struct {
	Debug   bool
	WorkDir string
}

type InstallCmd struct {
	File string `kong:"help='Path to a file with requirements',name='file',short='f',optional,default='requirements.yml'"`
}

type AddCmd struct {
	Url      string            `kong:"help='Package URL, will be skipped when installation from file is set.',name='url',short='u',arg,placeholder='url',optional"`
	Path     string            `kong:"help='Path to .apkg in the remote repository',name='path',short='p',default='.',optional"`
	Mappings map[string]string `kong:"help='Package mappings, will mount a source file or directory within a destination directory. Example, <remote_file_or_dir>@<version>=./roles',name='mappings',short='m',default='\"*@master=.\"',optional"`
	File     string            `kong:"help='Path to a file with requirements',name='file',short='f',optional,default='requirements.yml'"`
	// NoLink bool
	// Boost     bool
	// Save bool
}

type ListCmd struct {
	Url string `kong:"help='Package URL',arg,placeholder='url',required"`
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

func loadRequirements(filename string) (req *parser.Requirements, err error) {
	file, err := os.Open(filename)
	if os.IsNotExist(err) {
		return &parser.Requirements{}, nil
	} else if err != nil {
		return nil, err
	}
	defer file.Close()

	req, err = parser.Load(file)
	return req, err
}

func (cmd *InstallCmd) Run(ctx *Context) error {
	m := manager.Manager{}

	requirements, err := loadRequirements(cmd.File)
	if err != nil {
		logrus.Error(err)
		return err
	}

	// install from legacy file
	packages := make([]*manager.Package, 0)
	for _, pkg := range requirements.Packages {
		for _, mpg := range pkg.Mappings {
			packages = append(packages, &manager.Package{
				URL: pkg.Src,
				// Path:     cmd.Path,
				Path:     mpg.Src,
				Version:  mpg.Version,
				Mappings: []manager.Mapping{{Src: "", Dest: mpg.Dest}},
			})
		}
		if err := m.Install(packages, &manager.InstallOptions{
			WorkDir:     ctx.WorkDir,
			OnceDowload: true,
		}); err != nil {
			logrus.Error(err)
			return err
		}
		packages = packages[:0]
	}
	return nil
}

func (cmd *AddCmd) Run(ctx *Context) error {
	m := manager.Manager{}

	// requirements, err := loadRequirements(cmd.File)
	// if err != nil {
	// 	logrus.Error(err)
	// 	return err
	// }

	packages := make([]*manager.Package, 0)
	for k, v := range cmd.Mappings {
		src, version := parseUrl(k)
		packages = append(packages, &manager.Package{
			URL:      cmd.Url,
			Path:     cmd.Path,
			Version:  version,
			Mappings: []manager.Mapping{{Src: src, Dest: v}},
		})
	}
	if err := m.Install(packages, &manager.InstallOptions{WorkDir: ctx.WorkDir}); err != nil {
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
	Debug   bool   `kong:"help='Enable debug mode.',name='debug'"`
	WorkDir string `kong:"help='Working directory with .apm mount point. It is current directory by default',name='workdir',short='w',optional"`
	// User         string
	// AuthType     string
	// UseGitConfig bool
	Install InstallCmd `kong:"cmd,help='Install a package'"`
	List    ListCmd    `kong:"cmd,help='List remote versions'"`
	Add     AddCmd     `kong:"cmd,help='Add a package'"`
}
