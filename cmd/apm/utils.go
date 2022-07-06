package main

import (
	"github.com/k1nky/apm/internal/downloader"
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
