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
	Override     bool
	UseGitConfig bool
	Auth         AuthType
	Username     string
	Password     string
	OnlySwitch   bool
}

type Downloader struct {
	options *Options
}

func NewDownloader() *Downloader {
	return &Downloader{
		options: DefaultOptions(),
	}
}

func DefaultOptions() *Options {
	return &Options{
		Override:     true,
		UseGitConfig: true,
		Auth:         NoAuth,
		OnlySwitch:   false,
	}
}

func (options *Options) Validate() (err error) {
	return nil
}

func (d *Downloader) auth() (method transport.AuthMethod, err error) {
	switch d.options.Auth {
	case NoAuth:
	case SSHAgentAuth:
		method, err = ssh.NewSSHAgentAuth("")
	case BasicAuth:
		method = &http.BasicAuth{
			Username: d.options.Username,
			Password: d.options.Password,
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

func (d *Downloader) retrieveRemoteVersion(url string, options *Options) (versions []string, err error) {
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

func (d *Downloader) RewriteURLFromGitConfig(url string) string {
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

func (d *Downloader) prepareUrl(url string) (newUrl string, err error) {
	newUrl = url
	if !strings.HasPrefix(url, "http") && !strings.HasPrefix(url, "ssh") {
		newUrl = "https://" + url
	}
	if _, err = gourl.Parse(newUrl); err != nil {
		return
	}

	if d.options.UseGitConfig {
		newUrl = d.RewriteURLFromGitConfig(newUrl)
	}

	return
}

func (d *Downloader) prepare(url string, options *Options) (newUrl string, err error) {
	if options == nil {
		d.options = DefaultOptions()
	} else {
		d.options = options
	}
	if err = d.options.Validate(); err != nil {
		return
	}

	newUrl, err = d.prepareUrl(url)

	if strings.HasPrefix(newUrl, "ssh") {
		d.options.Auth = SSHAgentAuth
	}
	return
}

// Get a package from `url` with `version` to `dest` directory.
// If scheme is not specified for url will be used 'https'.
// Default version is 'master'.
func (d *Downloader) Get(url string, version string, dest string, options *Options) (newUrl string, err error) {

	if newUrl, err = d.prepare(url, options); err != nil {
		return
	}

	if !d.options.OnlySwitch {
		if err = d.clone(dest, newUrl); err != nil {
			return
		}
	}
	err = d.Switch(dest, version)

	return
}

func (d *Downloader) FetchVersion(url string, options *Options) (versions []string, err error) {
	if url, err = d.prepare(url, options); err != nil {
		return
	}

	versions, err = d.retrieveRemoteVersion(url, options)
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
