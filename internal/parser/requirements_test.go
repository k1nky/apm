package parser

import (
	"strings"
	"testing"
)

func TestParseCorrectRequirements(t *testing.T) {
	correctRequirements := `
packages:
- src: https://github.com/k1nky/ansible-simple-roles.git
  mappings:
    - src: motd
      dest: roles/motd
      version: master
`
	reader := strings.NewReader(correctRequirements)
	req := &Requirements{}
	err := req.Read(reader)
	if err != nil {
		t.Error(err)
	}
	if req == nil || len(req.Packages) == 0 {
		t.Error("requirements is nil or empty")
	}
}
