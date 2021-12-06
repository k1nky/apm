package downloader

import (
	"errors"

	"github.com/sirupsen/logrus"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/config"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/transport"
	"github.com/go-git/go-git/v5/plumbing/transport/http"
	"github.com/go-git/go-git/v5/plumbing/transport/ssh"
	"github.com/go-git/go-git/v5/storage/memory"
)

type DownloaderGlobs []string

const (
	DownloaderNoAuth = iota
	DownloaderSSHAgentAuth
	DownloaderBasicAuth
)

type DownloaderAuthType int

type DownloaderOptions struct {
	Override     bool
	UseGitConfig bool
	Auth         DownloaderAuthType
	Username     string
	Password     string
	Globs        DownloaderGlobs
}

type Downloader struct {
	repo    *git.Repository
	options *DownloaderOptions
}

func (options *DownloaderOptions) Validate() (err error) {
	if len(options.Globs) == 0 {
		options.Globs = append(options.Globs, "*")
	}
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

func (d *Downloader) clone(dir string, url string, ref *plumbing.Reference) (err error) {

	var (
		wt   *git.Worktree
		hash *plumbing.Hash
	)

	cloneOptions := &git.CloneOptions{
		URL:          url,
		SingleBranch: false,
		Tags:         git.AllTags,
	}
	if ref.Name().IsBranch() {
		cloneOptions.ReferenceName = ref.Name()
	}
	if method, err := d.auth(); err != nil {
		return err
	} else if method != nil {
		cloneOptions.Auth = method
	}
	if d.repo, err = git.PlainClone(dir, false, cloneOptions); err != nil {
		return
	}

	if hash, err = d.repo.ResolveRevision(plumbing.Revision(ref.Name())); err != nil {
		return
	}
	if wt, err = d.repo.Worktree(); err != nil {
		return
	}
	err = wt.Checkout(&git.CheckoutOptions{
		Hash: *hash,
	})

	return
}

func (d *Downloader) pulse(url string, version string) (*plumbing.Reference, error) {
	listOptions := &git.ListOptions{}
	remRepo := git.NewRemote(memory.NewStorage(), &config.RemoteConfig{
		Name: "origin",
		URLs: []string{url},
	})
	if method, err := d.auth(); err != nil {
		return nil, err
	} else if method != nil {
		listOptions.Auth = method
	}
	refs, err := remRepo.List(listOptions)
	if err != nil {
		return nil, err
	}

	for _, ref := range refs {
		if ref.Name().Short() == version {
			return ref, nil
		}
	}

	return plumbing.NewHashReference(plumbing.ReferenceName(version), plumbing.NewHash(version)), nil
}

func (d *Downloader) Get(url string, version string, dest string, options *DownloaderOptions) error {

	if err := options.Validate(); err != nil {
		return err
	}
	d.options = options

	logrus.Debugf("check pulse of %s", url)
	ref, err := d.pulse(url, version)
	if err != nil {
		return err
	}

	if err := d.clone(dest, url, ref); err != nil {
		return err
	}

	return nil
}
