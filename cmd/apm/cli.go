package main

import (
	"fmt"
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
	File         string
}

var CLI struct {
	Debug        bool   `help:"Enable debug mode." name:"debug"`
	WorkDir      string `help:"Working directory with .apm mount point. It is current directory by default" name:"workdir" short:"w" optional:""`
	UseGitConfig bool   `help:"Use gitconfig to override url" name:"gitconfig" default:"true" optional:"" negatable:""`
	File         string `help:"Path to a file with requirements" name:"file" short:"f" optional:"" default:"requirements.yml"`
	// TODO: User         string
	// TODO: AuthType     string
	Add     AddCmd     `cmd:"" help:"Add a package"`
	Install InstallCmd `cmd:"" help:"Install packages from file"`
	List    ListCmd    `cmd:"" help:"List remote versions"`
	Link    LinkCmd    `cmd:"" help:"Link resources"`
	Version VersionCmd `cmd:"" help:"Show current version" short:"v"`
}

type InstallCmd struct {
}

type VersionCmd struct {
}

type LinkCmd struct {
	Url      string            `help:"Package URL, will be skipped when installation from file is set." name:"url" short:"u" arg:"" placeholder:"url" optional:""`
	Version  string            `help:"Version" name:"version" short:"v" default:"master" optional:""`
	Mappings map[string]string `help:"Package mappings, will mount a source file or directory within a destination directory. Example, <remote_file_or_dir>=./roles" name:"map" short:"m" default:"*=.\"" optional:""`
	Save     bool              `help:"Save added package to requirements" name:"save" short:"s" optional:"" default:"false"`
	// TODO: NoLink bool
	// TODO: Boost     bool
}

type AddCmd struct {
	Url     string   `help:"Package URL, will be skipped when installation from file is set." name:"url" short:"u" arg:"" placeholder:"url" optional:""`
	Version string   `help:"Version" name:"version" short:"v" default:"master" optional:""`
	Paths   []string `help:"Path to .apkg in the remote repository" name:"path" short:"p" default:"." optional:""`
	Save    bool     `help:"Save added package to requirements" name:"save" short:"s" optional:"" default:"false"`
}

type ListCmd struct {
	Url string `help:"Package URL" arg:"" placeholder:"url" required:""`
}

func (cmd *InstallCmd) Run(ctx *Context) error {
	m := manager.Manager{}

	requirements, err := loadRequirements(ctx.File)
	if err != nil {
		logrus.Error(err)
		return err
	}

	packages := make([]*manager.Package, 0)
	for _, pkg := range requirements.Packages {
		for _, mpg := range pkg.Mappings {
			packages = append(packages, &manager.Package{
				URL: overrideUrl(pkg.Url, ctx.UseGitConfig),
				// Path:     cmd.Path,
				Path:     mpg.Src,
				Version:  mpg.Version,
				Mappings: []manager.Mapping{{Src: "", Dest: mpg.Dest}},
			})
		}
		if err := m.Install(packages, &manager.InstallOptions{
			WorkDir:      ctx.WorkDir,
			OnceDownload: true,
		}); err != nil {
			logrus.Error(err)
			return err
		}
		packages = packages[:0]
	}
	return nil
}

func (cmd *LinkCmd) Run(ctx *Context) error {

	m := manager.Manager{}

	requirements, err := loadRequirements(ctx.File)
	if err != nil {
		logrus.Error(err)
		return err
	}

	packages := make([]*manager.Package, 0)
	url := overrideUrl(cmd.Url, ctx.UseGitConfig)
	for k, v := range cmd.Mappings {
		src := strings.Trim(k, " ")
		packages = append(packages, &manager.Package{
			URL:      url,
			Path:     ".",
			Version:  cmd.Version,
			Mappings: []manager.Mapping{{Src: src, Dest: v}},
		})
		requirements.Add(parser.RequiredPackage{
			// use original url to prevent unexpected overriding
			Url: cmd.Url,
			Mappings: []parser.ReqiuredMapping{
				{
					Src:     src,
					Dest:    v,
					Version: cmd.Version,
				},
			},
		})
	}
	// TODO: setup InstallOptions
	if err := m.Install(packages, &manager.InstallOptions{WorkDir: ctx.WorkDir}); err != nil {
		logrus.Error(err)
		return err
	}

	if cmd.Save {
		saveRequirements(ctx.File, requirements)
	}

	return nil
}

func (cmd *AddCmd) Run(ctx *Context) error {

	m := manager.Manager{}

	requirements, err := loadRequirements(ctx.File)
	if err != nil {
		logrus.Error(err)
		return err
	}

	url := overrideUrl(cmd.Url, ctx.UseGitConfig)
	for _, path := range cmd.Paths {
		pkg := &manager.Package{
			URL:     url,
			Path:    path,
			Version: cmd.Version,
		}
		if err := m.InstallFromApkg(pkg, &manager.InstallOptions{WorkDir: ctx.WorkDir}); err != nil {
			logrus.Error(err)
			return err
		}
		for _, mpg := range pkg.Mappings {
			requirements.Add(parser.RequiredPackage{
				Url: cmd.Url,
				Mappings: []parser.ReqiuredMapping{{
					Src:     path,
					Dest:    mpg.Dest,
					Version: cmd.Version,
				}},
			})
		}
	}

	if cmd.Save {
		saveRequirements(ctx.File, requirements)
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

func (cmd *VersionCmd) Run(ctx *Context) (err error) {
	fmt.Printf("%s %s\n", BuildTarget, BuildVersion)
	return
}
