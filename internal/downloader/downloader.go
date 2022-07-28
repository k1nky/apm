package downloader

import (
	"errors"
	"fmt"
	gourl "net/url"
	"os"
	"sort"
	"strings"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/config"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/transport"
	"github.com/go-git/go-git/v5/plumbing/transport/http"
	"github.com/go-git/go-git/v5/plumbing/transport/ssh"
	"github.com/go-git/go-git/v5/storage/memory"
	"github.com/sirupsen/logrus"
)

const (
	NoAuth = iota
	SSHAgentAuth
	BasicAuth
)

type AuthType int

type Options struct {
	// TODO: Override existing directory
	Override   bool
	Auth       AuthType
	Username   string
	Password   string
	OnlySwitch bool
}

type Downloader struct {
	Options *Options
}

func NewDownloader(opts *Options) *Downloader {
	d := &Downloader{}
	d.Options = opts
	if d.Options == nil {
		d.Options = DefaultOptions()
	}
	return d
}

func DefaultOptions() *Options {
	return &Options{
		Override:   true,
		Auth:       NoAuth,
		OnlySwitch: false,
	}
}

func (options *Options) Validate() (err error) {
	return nil
}

func (d *Downloader) auth() (method transport.AuthMethod, err error) {
	switch d.Options.Auth {
	case NoAuth:
	case SSHAgentAuth:
		method, err = ssh.NewSSHAgentAuth("")
	case BasicAuth:
		method = &http.BasicAuth{
			Username: d.Options.Username,
			Password: d.Options.Password,
		}
	default:
		err = errors.New("unsupported auth method")
	}

	return
}

func (d *Downloader) clone(dir string, url string) (err error) {
	cloneOptions := &git.CloneOptions{
		URL:          url,
		SingleBranch: false,
		Tags:         git.AllTags,
		RemoteName:   "origin",
	}
	if logrus.GetLevel() < logrus.ErrorLevel {
		cloneOptions.Progress = os.Stdout
	}
	if method, err := d.auth(); err != nil {
		return err
	} else if method != nil {
		cloneOptions.Auth = method
	}
	_, err = git.PlainClone(dir, false, cloneOptions)

	return
}

func (d *Downloader) retrieveRemoteVersion(url string) (versions []string, err error) {
	listOptions := &git.ListOptions{
		InsecureSkipTLS: true,
	}
	if method, err := d.auth(); err != nil {
		return versions, err
	} else if method != nil {
		listOptions.Auth = method
	}

	remrepo := git.NewRemote(memory.NewStorage(), &config.RemoteConfig{
		Name: "origin",
		URLs: []string{url},
	})

	refs, err := remrepo.List(listOptions)
	if err != nil {
		return versions, err
	}
	for _, ref := range refs {
		if ref.Name().IsTag() || ref.Name().IsBranch() {
			versions = append(versions, ref.Name().Short())
		}
	}

	return
}

func RewriteURLFromGitConfig(url string) string {
	if cfg, err := config.LoadConfig(config.GlobalScope); err != nil {
		logrus.Debug(err)
	} else {
		for _, section := range cfg.Raw.Sections {
			if section.Name == "url" {
				for _, subsection := range section.Subsections {
					for _, opt := range subsection.Options {
						if opt.Key == "insteadOf" && strings.HasPrefix(url, opt.Value) {
							return strings.ReplaceAll(url, opt.Value, subsection.Name)
						}
					}
				}
			}
		}
	}
	return url
}

func RewriteUrl(url string, useGitConfig bool) (newUrl string, err error) {
	newUrl = url
	if !strings.HasPrefix(url, "http") && !strings.HasPrefix(url, "ssh") {
		newUrl = "https://" + url
	}
	if _, err = gourl.ParseRequestURI(newUrl); err != nil {
		return
	}

	if useGitConfig {
		newUrl = RewriteURLFromGitConfig(newUrl)
	}

	return
}

func (d *Downloader) prepare(url string) (err error) {
	if d.Options == nil {
		d.Options = DefaultOptions()
	}
	if err = d.Options.Validate(); err != nil {
		return
	}

	if strings.HasPrefix(url, "ssh") {
		d.Options.Auth = SSHAgentAuth
	}
	return
}

// Get a package from `url` with `version` to `dest` directory.
// If scheme is not specified for url will be used 'https'.
// Default version is 'master'.
func (d *Downloader) Get(url string, version string, dest string) (err error) {

	if err = d.prepare(url); err != nil {
		return
	}

	if !d.Options.OnlySwitch {
		if err = d.clone(dest, url); err != nil {
			return
		}
	}
	err = d.Switch(dest, version)

	return
}

func (d *Downloader) FetchVersion(url string) (versions []string, err error) {
	if err = d.prepare(url); err != nil {
		return
	}

	versions, err = d.retrieveRemoteVersion(url)
	sort.Strings(versions)
	return
}

func (d *Downloader) Switch(dir string, version string) (err error) {
	var (
		wt   *git.Worktree
		hash *plumbing.Hash
		repo *git.Repository
	)

	revisions := []plumbing.Revision{plumbing.Revision(version), plumbing.Revision("origin/" + version)}

	if repo, err = git.PlainOpen(dir); err != nil {
		return
	}
	// Workaround: https://github.com/go-git/go-git/issues/148#issuecomment-989635832
	repo.ResolveRevision(plumbing.Revision("HEAD"))
	for _, rev := range revisions {
		if hash, err = repo.ResolveRevision(rev); err == nil {
			break
		}
		hash = nil
	}
	if hash == nil {
		return fmt.Errorf("version not found")
	}
	if wt, err = repo.Worktree(); err != nil {
		return
	}
	err = wt.Checkout(&git.CheckoutOptions{
		Hash: *hash,
	})
	return
}
