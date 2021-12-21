package manager

import (
	"crypto/md5"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/k1nky/apm/internal/copy"
	"github.com/k1nky/apm/internal/downloader"
)

type Manager struct {
	TmpDir  string
	Storage string
	WorkDir string
}
type InstallOptions struct {
	DownloadOptions *downloader.DownloaderOptions
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

func (m *Manager) MakeStorage(dir string) (err error) {
	if len(dir) == 0 {
		dir = DefaultStoragePath
	}
	if strings.HasPrefix(dir, "~/") {
		home, _ := os.UserHomeDir()
		dir = filepath.Join(home, dir[2:])
	}

	m.Storage = dir
	if err := os.MkdirAll(m.Storage, copy.DefaultDestinationMode); err != nil && !os.IsExist(err) {
		return err
	}

	return nil
}

func (p Package) Hash() string {
	return fmt.Sprintf("%x", md5.Sum([]byte(p.URL+p.Version+p.Path)))
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

func (m Manager) cleanup() {
	if m.TmpDir != "" {
		os.RemoveAll(m.TmpDir)
	}
}

func (m *Manager) download(p *Package, opts *downloader.DownloaderOptions) (err error) {

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
	if err = m.MakeStorage(""); err != nil {
		return
	}

	hash := p.Hash()
	p.storagePath = path.Join(m.Storage, hash)
	if err = os.MkdirAll(p.storagePath, copy.DefaultDestinationMode); err != nil {
		return err
	}

	// copy to storage
	err = copy.Copy(m.TmpDir, []string{p.Path}, p.storagePath, nil)
	return
}

func (m *Manager) setupWorkdir(p *Package) (err error) {
	if m.WorkDir == "" {
		if m.WorkDir, err = os.Getwd(); err != nil {
			return
		}
	}

	if err = os.MkdirAll(path.Join(m.WorkDir, ".apm"), copy.DefaultDestinationMode); err != nil {
		return
	}
	if err = os.Chdir(m.WorkDir); err != nil {
		return
	}
	if err = os.Symlink(p.storagePath, path.Join(".apm", p.Hash())); err != nil {
		return
	}
	return
}

func makeLink(root string, name string, target string) (err error) {
	if err = os.MkdirAll(root, copy.DefaultDestinationMode); err != nil {
		return
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
		if m.Src == "" || m.Src == "." {
			relpath, _ = filepath.Rel(path.Dir(m.Dest), packagePath)
			if err = makeLink(path.Dir(m.Dest), m.Dest, relpath); err != nil {
				return
			}
		} else {
			if fs, err = copy.ResolveGlob(packagePath, m.Src, nil); err != nil {
				return
			}
			for _, f := range fs {
				basename = path.Base(f)
				relpath, _ = filepath.Rel(m.Dest, path.Dir(f))
				if err = makeLink(m.Dest, path.Join(m.Dest, basename), path.Join(relpath, basename)); err != nil {
					return
				}
			}
		}
	}

	return
}

func (m Manager) Install(p *Package, opts *InstallOptions) (err error) {

	if err := p.Validate(); err != nil {
		return err
	}
	if err = m.download(p, opts.DownloadOptions); err != nil {
		return
	}
	defer m.cleanup()

	if err = m.unpack(p); err != nil {
		return
	}

	if err = m.setupWorkdir(p); err != nil {
		return
	}

	if err = m.setupMappings(p); err != nil {
		return
	}

	return nil
}
