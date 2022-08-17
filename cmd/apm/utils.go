package main

import (
	"os"
	"os/user"
	"path/filepath"
	"strings"

	"github.com/k1nky/apm/internal/downloader"
	"github.com/k1nky/apm/internal/parser"
	"github.com/sirupsen/logrus"
)

func overrideURL(url string, useGitConfig bool) string {
	newUrl, err := downloader.RewriteURL(url, useGitConfig)
	if err != nil {
		logrus.Fatal(err)
		return ""
	}
	logrus.Debugf("override url %s to %s", url, newUrl)
	return newUrl
}

func loadRequirements(filename string) (req *parser.Requirements, err error) {
	file, err := os.Open(filename)
	if os.IsNotExist(err) {
		return &parser.Requirements{}, nil
	} else if err != nil {
		return nil, err
	}
	defer file.Close()

	req = &parser.Requirements{}
	err = req.Read(file)

	return req, err
}

func saveRequirements(filename string, req *parser.Requirements) error {
	file, err := os.Create(filename)
	if err != nil {
		logrus.Error(err)
		return err
	}
	defer file.Close()

	if req == nil {
		logrus.Warn("nothing to save")
		return nil
	}

	if err := req.Write(file); err != nil {
		logrus.Error(err)
		return err
	}

	return nil
}

func expandPath(p string) string {
	if strings.HasPrefix(p, "~") {
		usr, _ := user.Current()
		return filepath.Join(usr.HomeDir, p[2:])
	}
	return p
}
