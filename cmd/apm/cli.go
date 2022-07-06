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
	Debug        bool
	UseGitConfig bool
	WorkDir      string
}

type InstallCmd struct {
	File string `kong:"help='Path to a file with requirements',name='file',short='f',optional,default='requirements.yml'"`
}

type AddCmd struct {
	Url      string            `kong:"help='Package URL, will be skipped when installation from file is set.',name='url',short='u',arg,placeholder='url',optional"`
	Path     string            `kong:"help='Path to .apkg in the remote repository',name='path',short='p',default='.',optional"`
	Mappings map[string]string `kong:"help='Package mappings, will mount a source file or directory within a destination directory. Example, <remote_file_or_dir>@<version>=./roles',name='mappings',short='m',default='\"*@master=.\"',optional"`
	File     string            `kong:"help='Path to a file with requirements',name='file',short='f',optional,default='requirements.yml'"`
	Save     bool              `kong:"help='Save added package to requirements',name='save',short='s',optional,default=false"`
	// NoLink bool
	// Boost     bool
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

func saveRequirements(filename string, req *parser.Requirements) error {
	file, err := os.Create(filename)
	if err != nil {
		logrus.Error(err)
		return err
	}
	defer file.Close()

	if err := parser.Save(file, req); err != nil {
		logrus.Error(err)
		return err
	}

	return nil
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
				URL: overrideUrl(pkg.Src, ctx.UseGitConfig),
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

	requirements, err := loadRequirements(cmd.File)
	if err != nil {
		logrus.Error(err)
		return err
	}

	packages := make([]*manager.Package, 0)
	url := overrideUrl(cmd.Url, ctx.UseGitConfig)
	for k, v := range cmd.Mappings {
		src, version := parseUrl(strings.Trim(k, " "))
		packages = append(packages, &manager.Package{
			URL:      url,
			Path:     cmd.Path,
			Version:  version,
			Mappings: []manager.Mapping{{Src: src, Dest: v}},
		})
		requirements.Add(parser.Package{
			Src:      url,
			Mappings: []parser.ReqiurementMapping{{SrcDest: parser.SrcDest{Src: src, Dest: v}, Version: version}},
		})
	}
	if err := m.Install(packages, &manager.InstallOptions{WorkDir: ctx.WorkDir}); err != nil {
		logrus.Error(err)
		return err
	}

	if cmd.Save {
		saveRequirements(cmd.File, requirements)
	}

	return nil
}

func (cmd *ListCmd) Run(ctx *Context) (err error) {
	var versions []string
	d := downloader.NewDownloader()
	url := overrideUrl(cmd.Url, ctx.UseGitConfig)
	versions, err = d.FetchVersion(url, nil)
	for _, v := range versions {
		fmt.Println(v)
	}

	return
}

var CLI struct {
	Debug        bool   `kong:"help='Enable debug mode.',name='debug'"`
	WorkDir      string `kong:"help='Working directory with .apm mount point. It is current directory by default',name='workdir',short='w',optional"`
	UseGitConfig bool   `kong:"help='Use gitconfig to override url',name='use-gitconfig',default=true,optional"`
	// User         string
	// AuthType     string
	Install InstallCmd `kong:"cmd,help='Install a package'"`
	List    ListCmd    `kong:"cmd,help='List remote versions'"`
	Add     AddCmd     `kong:"cmd,help='Add a package'"`
}
