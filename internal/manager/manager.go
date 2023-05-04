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

	"github.com/pterm/pterm"
)

type Manager struct {
	TmpDir  string
	Storage string
	WorkDir string
}
type InstallOptions struct {
	WorkDir         string
	DownloadOptions *downloader.Options
	Force           bool
}
type Package struct {
	URL     string
	Version string
	Src     string
	Dest    string
}

const (
	DefaultStoragePath = "~/.apm"
	DefaultVersion     = "master"
	DefaultTmpPrefix   = "apm-"
)

func DefaultInstallOptions() *InstallOptions {
	return &InstallOptions{
		DownloadOptions: downloader.DefaultOptions(),
		WorkDir:         ".",
		Force:           false,
	}
}

func (opts *InstallOptions) Validate() error {
	if opts.DownloadOptions == nil {
		opts.DownloadOptions = downloader.DefaultOptions()
	}
	return nil
}

func (p Package) Hash() string {
	return fmt.Sprintf("%x", md5.Sum([]byte(p.URL+p.Src+p.Version)))
}

func (p Package) String() (s string) {
	s = fmt.Sprintf("%s,%s,%s", p.URL, p.Version, p.Src)
	return
}

func (p *Package) Validate() error {
	if p.URL == "" {
		return fmt.Errorf("invalid package url")
	}
	if p.Version == "" {
		p.Version = DefaultVersion
	}
	if p.Src == "" {
		p.Src = "."
	}
	if p.Dest == "" {
		return fmt.Errorf("invalid package destination")
	}
	return nil
}

func PackageFromString(str string) *Package {
	chunks := strings.Split(str, ",")
	p := &Package{
		URL: chunks[0],
	}
	if len(chunks) > 1 {
		p.Version = chunks[1]
		if len(chunks) > 2 {
			p.Src = chunks[2]
		}
	}
	return p
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

// func (m *Manager) unpack(destDir string, src string) (err error) {
func (m *Manager) unpack(src string, dest string) (err error) {
	// copy to storage
	tmpSrc := path.Join(m.TmpDir, src)
	if info, err := os.Stat(tmpSrc); err != nil {
		return err
	} else {
		if info.Mode().IsRegular() {
			dest = path.Join(dest, src)
		}
	}

	if err = copy.Copy(tmpSrc, dest, &copy.CopyOptions{
		Override: true,
		Plain:    src == "" || src == ".",
	}); err != nil {
		return
	}

	return
}

func (m *Manager) SetupWorkdir(wd string) (err error) {
	m.WorkDir = wd
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

func (m *Manager) setup(pkg *Package) (err error) {

	var (
		relpath string
	)

	pkgHash := pkg.Hash()
	pkgStoragePath := path.Join(m.Storage, pkgHash)

	if err = m.unpack(pkg.Src, pkgStoragePath); err != nil {
		return
	}
	if err = makeLink(path.Join(".apm", pkgHash), path.Join(pkgStoragePath, pkg.Src), true); err != nil {
		return
	}

	pkgLocalPath := path.Join(".apm", pkgHash)

	relpath, _ = filepath.Rel(path.Dir(pkg.Dest), pkgLocalPath)
	if err = os.MkdirAll(path.Dir(pkg.Dest), copy.Mode0755); err != nil {
		return
	}
	if err = makeLink(pkg.Dest, relpath, true); err != nil {
		return
	}

	return
}

func (m *Manager) Install(pkgs []*Package, opts *InstallOptions) (err error) {

	defer m.cleanup()

	if opts == nil {
		opts = DefaultInstallOptions()
	} else {
		opts.Validate()
	}

	if err = m.MakeStorage(""); err != nil {
		return
	}

	if err = m.SetupWorkdir(opts.WorkDir); err != nil {
		return
	}

	progressBar, _ := pterm.DefaultProgressbar.WithTotal(len(pkgs)).WithTitle("Installing").Start()

	for _, p := range pkgs {
		m.cleanup()
		progressBar.UpdateTitle("Installing " + p.String())

		if err := p.Validate(); err != nil {
			pterm.Warning.Println(err)
			continue
		}

		if err = m.download(p, opts.DownloadOptions); err != nil {
			pterm.Warning.Println(err)
			continue
		}

		if err = m.setup(p); err != nil {
			pterm.Warning.Println(err)
			continue
		}
		pterm.Success.Printfln("Installing " + p.String())
		progressBar.Increment()
	}

	return nil
}
