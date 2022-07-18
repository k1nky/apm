package manager

import (
	"crypto/md5"
	"fmt"
	"io/fs"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/k1nky/apm/internal/copy"
	"github.com/k1nky/apm/internal/downloader"
	"github.com/sirupsen/logrus"
)

type Manager struct {
	TmpDir  string
	Storage string
	WorkDir string
}
type InstallOptions struct {
	WorkDir         string
	DownloadOptions *downloader.Options
	OnceDownload    bool
	Force           bool
}
type Mapping struct {
	Src  string
	Dest string
}
type Package struct {
	URL         string
	Version     string
	Path        string
	Mappings    []Mapping
	storagePath string
}

const (
	DefaultStoragePath = "~/.apm"
	DefaultVersion     = "master"
	DefaultTmpPrefix   = "apm-"
)

func (opts *InstallOptions) Validate() error {
	if opts.DownloadOptions == nil {
		opts.DownloadOptions = downloader.DefaultOptions()
	}
	return nil
}

func DefaultInstallOptions() InstallOptions {
	return InstallOptions{
		DownloadOptions: downloader.DefaultOptions(),
		WorkDir:         ".",
		OnceDownload:    true,
		Force:           false,
	}
}

func (m *Manager) MakeStorage(dir string) (err error) {
	if len(dir) == 0 {
		dir = DefaultStoragePath
	}
	if strings.HasPrefix(dir, "~/") {
		home, _ := os.UserHomeDir()
		dir = filepath.Join(home, dir[2:])
	}

	m.Storage = dir
	if err := os.MkdirAll(m.Storage, copy.Mode0755); err != nil && !os.IsExist(err) {
		return err
	}

	return nil
}

func (p Package) Hash() string {
	return fmt.Sprintf("%x@%s", md5.Sum([]byte(p.URL+p.Path)), p.Version)
}

func (p *Package) Validate() error {
	if p.URL == "" {
		return fmt.Errorf("invalid package url")
	}
	if p.Version == "" {
		p.Version = DefaultVersion
	}
	if p.Path == "" {
		p.Path = "."
	}
	if len(p.Mappings) == 0 {
		p.Mappings = []Mapping{{
			Src:  "*",
			Dest: ".",
		}}
	}
	return nil
}

func (m *Manager) cleanup() {
	if m.TmpDir != "" {
		os.RemoveAll(m.TmpDir)
		m.TmpDir = ""
	}
}

func (m *Manager) download(p *Package, opts *downloader.Options) (err error) {

	d := downloader.NewDownloader()
	if m.TmpDir == "" {
		if m.TmpDir, err = ioutil.TempDir("", DefaultTmpPrefix); err != nil {
			return
		}
	}

	err = d.Get(p.URL, p.Version, m.TmpDir, opts)
	return
}

func (m *Manager) unpack(p *Package) (err error) {
	if err = os.MkdirAll(p.storagePath, copy.Mode0755); err != nil {
		return err
	}

	// copy to storage
	if err = copy.Copy(m.TmpDir, p.storagePath, &copy.CopyOptions{Override: true}); err != nil {
		return
	}
	err = os.RemoveAll(path.Join(p.storagePath, ".git"))
	return
}

func (m *Manager) setupWorkdir(p *Package) (err error) {
	if m.WorkDir == "" {
		if m.WorkDir, err = os.Getwd(); err != nil {
			return
		}
	}

	if err = os.MkdirAll(path.Join(m.WorkDir, ".apm"), copy.Mode0755); err != nil {
		if !os.IsExist(err) {
			return
		}
	}
	if err = os.Chdir(m.WorkDir); err != nil {
		return
	}
	err = makeLink(path.Join(".apm", p.Hash()), p.storagePath, true)
	return
}

func makeLink(name string, target string, override bool) (err error) {
	var info fs.FileInfo

	if info, err = os.Lstat(name); err == nil {
		if !override || info.Mode()&os.ModeSymlink == 0 {
			return fmt.Errorf("file %s already exists", name)
		}
		if err = os.Remove(name); err != nil {
			return
		}
		err = nil
	}
	err = os.Symlink(target, name)
	return
}

func (m *Manager) setupMappings(p *Package) (err error) {

	var (
		relpath  string
		basename string
		fs       []string
	)

	packagePath := path.Join(".apm", p.Hash())

	// TODO: validate mappings

	for _, m := range p.Mappings {
		m.Src = path.Join(p.Path, m.Src)
		if m.Src == "" || m.Src == "." {
			relpath, _ = filepath.Rel(path.Dir(m.Dest), packagePath)
			if err = os.MkdirAll(path.Dir(m.Dest), copy.Mode0755); err != nil {
				return
			}
			if err = makeLink(m.Dest, relpath, true); err != nil {
				return
			}
		} else if strings.Contains(m.Src, "*") {
			if fs, err = copy.ResolveGlob(packagePath, m.Src, nil); err != nil {
				return
			}
			for _, f := range fs {
				basename = path.Base(f)
				relpath, _ = filepath.Rel(m.Dest, path.Dir(f))
				if err = os.MkdirAll(m.Dest, copy.Mode0755); err != nil {
					return
				}
				if err = makeLink(path.Join(m.Dest, basename), path.Join(relpath, basename), true); err != nil {
					return
				}
			}
		} else {
			relpath, _ = filepath.Rel(path.Dir(m.Dest), path.Join(packagePath, m.Src))
			if err = os.MkdirAll(path.Dir(m.Dest), copy.Mode0755); err != nil {
				return
			}
			if err = makeLink(m.Dest, relpath, true); err != nil {
				return
			}
		}
	}

	return
}

func (m *Manager) Install(pkgs []*Package, opts *InstallOptions) (err error) {

	defer m.cleanup()

	opts.Validate()

	if err = m.MakeStorage(""); err != nil {
		return
	}

	for _, p := range pkgs {
		contextLogger := logrus.WithFields(logrus.Fields{
			"url":     p.URL,
			"version": p.Version,
			"path":    p.Path,
		})
		contextLogger.Info("installing a package")

		if err := p.Validate(); err != nil {
			return err
		}
		if opts == nil {
			opts = &InstallOptions{}
		}
		m.WorkDir = opts.WorkDir
		p.storagePath = path.Join(m.Storage, p.Hash())

		if opts.Force || !m.isPackageExist(p) {
			if err = m.download(p, opts.DownloadOptions); err != nil {
				return
			}
			if err = m.unpack(p); err != nil {
				return
			}
		}
		opts.DownloadOptions.OnlySwitch = opts.OnceDownload

		if err = m.setup(p); err != nil {
			return
		}

		contextLogger.Info("the package is installed")
	}

	return nil
}

func (m *Manager) isPackageExist(p *Package) bool {
	_, err := os.Stat(p.storagePath)
	return os.IsExist(err)
}

func (m *Manager) setup(p *Package) (err error) {

	if err = m.setupWorkdir(p); err != nil {
		return
	}

	if err = m.setupMappings(p); err != nil {
		return
	}

	return nil
}
