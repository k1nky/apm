package main

import (
	"testing"
)

func TestParseUrl(t *testing.T) {
	what := []string{"github.com", "github.com@dev", "github.com@"}
	want := [][]string{
		{"github.com", "master"},
		{"github.com", "dev"},
		{"github.com", "master"},
	}
	for k, v := range what {
		t.Run(v, func(t *testing.T) {
			url, version := parseUrl(v)
			if url != want[k][0] || version != want[k][1] {
				t.Error("expect ", want[k])
			}
		})
	}
}
