package downloader

import (
	"errors"
	"fmt"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/transport"
	"github.com/go-git/go-git/v5/plumbing/transport/http"
	"github.com/go-git/go-git/v5/plumbing/transport/ssh"
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
		options: &DownloaderOptions{
			Override:     true,
			UseGitConfig: true,
			Auth:         DownloaderNoAuth,
		},
	}
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
		// revisions []plumbing.Revision = []plumbing.Revision{version, fmt.Sprintf("origin/{}", version)}
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

func (d *Downloader) Get(url string, version string, dest string, options *DownloaderOptions) error {

	if err := options.Validate(); err != nil {
		return err
	}
	d.options = options

	if err := d.clone(dest, url); err != nil {
		return err
	}
	if err := d.Switch(dest, version); err != nil {
		return err
	}

	return nil
}
