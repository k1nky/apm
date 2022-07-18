package main

import (
	"os"
	"strings"

	"github.com/k1nky/apm/internal/downloader"
	"github.com/k1nky/apm/internal/parser"
	"github.com/sirupsen/logrus"
)

func overrideUrl(url string, useGitConfig bool) string {
	newUrl, err := downloader.RewriteUrl(url, useGitConfig)
	if err != nil {
		logrus.Fatal(err)
		return ""
	}
	logrus.Debugf("override url %s to %s", url, newUrl)
	return newUrl
}

func parseUrl(s string) (url string, version string) {
	parts := strings.Split(s, "@")
	if len(parts) == 0 {
		return
	}
	url = parts[0]
	if len(parts) > 1 {
		version = parts[1]
	}
	if len(version) == 0 {
		version = "master"
	}
	return
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
