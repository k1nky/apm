package main

import (
	"fmt"

	"github.com/k1nky/apm/internal/downloader"
	"github.com/k1nky/apm/internal/manager"
	"github.com/k1nky/apm/internal/parser"
	"github.com/pterm/pterm"
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
	Install InstallCmd `cmd:"" help:"Install packages from file"`
	List    ListCmd    `cmd:"" help:"List remote versions"`
	Link    LinkCmd    `cmd:"" help:"Link resources"`
	Version VersionCmd `cmd:"" help:"Show current version" aliases:"v"`
}

type InstallCmd struct {
}

type VersionCmd struct {
}

type LinkCmd struct {
	Url  string `help:"Package URL, will be skipped when installation from file is set." name:"url" short:"u" arg:"" placeholder:"url" optional:""`
	Dest string `help:"Destitaion." name:"dest" short:"d" arg:"" placeholder:"dest" optional:""`
	Save bool   `help:"Save added package to requirements" name:"save" short:"s" optional:"" default:"false"`
	// TODO: NoLink bool
	// TODO: Force bool
}

type ListCmd struct {
	Url string `help:"Package URL" arg:"" placeholder:"url" required:""`
}

func (cmd *InstallCmd) Run(ctx *Context) error {
	m := manager.Manager{}

	requirements, err := loadRequirements(ctx.File)
	if err != nil {
		pterm.Error.Println(err)
		return err
	}

	packages := make([]*manager.Package, 0)
	opts := &manager.InstallOptions{
		WorkDir: ctx.WorkDir,
	}
	for _, pkg := range requirements.Packages {
		for _, mpg := range pkg.Mappings {
			packages = append(packages, &manager.Package{
				URL:     overrideUrl(pkg.Url, ctx.UseGitConfig),
				Src:     mpg.Src,
				Version: mpg.Version,
				Dest:    mpg.Dest,
			})
		}
		if err := m.Install(packages, opts); err != nil {
			pterm.Error.Println(err)
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
		pterm.Error.Println(err)
		return err
	}

	packages := make([]*manager.Package, 0)
	// url := overrideUrl(cmd.Url, ctx.UseGitConfig)
	opts := &manager.InstallOptions{
		WorkDir: ctx.WorkDir,
	}
	pkg := manager.PackageFromString(cmd.Url)
	pkg.Dest = cmd.Dest
	packages = append(packages, pkg)
	requirements.Add(parser.RequiredPackage{
		// use original url to prevent unexpected overriding
		Url: pkg.URL,
		Mappings: []parser.ReqiuredMapping{
			{
				Src:     pkg.Src,
				Dest:    pkg.Dest,
				Version: pkg.Version,
			},
		},
	})
	if err := m.Install(packages, opts); err != nil {
		pterm.Error.Println(err)
		return err
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
