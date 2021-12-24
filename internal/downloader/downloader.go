package downloader

import (
	"errors"
	"fmt"
	"net/url"
	"strings"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/config"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/transport"
	"github.com/go-git/go-git/v5/plumbing/transport/http"
	"github.com/go-git/go-git/v5/plumbing/transport/ssh"
	"github.com/go-git/go-git/v5/storage/memory"
)

const (
	DownloaderNoAuth = iota
	DownloaderSSHAgentAuth
	DownloaderBasicAuth
)

type DownloaderAuthType int

type DownloaderOptions struct {
	// TODO: Override existing directory
	Override bool
	// TODO: Read .gitconfig to override URL
	UseGitConfig bool
	Auth         DownloaderAuthType
	Username     string
	Password     string
}

type Downloader struct {
	options *DownloaderOptions
}

func NewDownloader() *Downloader {
	return &Downloader{
		options: DefaultDownloaderOptions(),
	}
}

func DefaultDownloaderOptions() *DownloaderOptions {
	return &DownloaderOptions{
		Override:     true,
		UseGitConfig: true,
		Auth:         DownloaderNoAuth,
	}
}

func ValidateUrl(s string) (string, error) {
	if !strings.HasPrefix(s, "http") {
		s = "https://" + s
	}
	_, err := url.Parse(s)
	return s, err
}

func (options *DownloaderOptions) Validate() (err error) {
	return nil
}

func (d *Downloader) auth() (method transport.AuthMethod, err error) {
	switch d.options.Auth {
	case DownloaderNoAuth:
	case DownloaderSSHAgentAuth:
		method, err = ssh.NewSSHAgentAuth("")
	case DownloaderBasicAuth:
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
	if method, err := d.auth(); err != nil {
		return err
	} else if method != nil {
		cloneOptions.Auth = method
	}
	_, err = git.PlainClone(dir, false, cloneOptions)

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

func (d *Downloader) Get(url string, version string, dest string, options *DownloaderOptions) (err error) {

	if options == nil {
		options = DefaultDownloaderOptions()
	}
	if err = options.Validate(); err != nil {
		return
	}
	d.options = options
	if url, err = ValidateUrl(url); err != nil {
		return
	}

	if err = d.clone(dest, url); err != nil {
		return err
	}
	err = d.Switch(dest, version)

	return
}

func (d *Downloader) retrieveRemoteVersion(url string, options *DownloaderOptions) (versions []string, err error) {
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

func (d *Downloader) FetchVersion(url string, options *DownloaderOptions) (versions []string, err error) {
	if options == nil {
		options = DefaultDownloaderOptions()
	}
	if err = options.Validate(); err != nil {
		return
	}
	d.options = options
	if url, err = ValidateUrl(url); err != nil {
		return
	}

	versions, err = d.retrieveRemoteVersion(url, options)
	return
}
