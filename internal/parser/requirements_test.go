package parser

import (
	"bytes"
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

func TestWriteCorrectRequirements(t *testing.T) {
	req := Requirements{
		Packages: []RequiredPackage{
			{
				Url: "https://github.com/k1nky/ansible-simple-roles.git",
				Mappings: []ReqiuredMapping{
					{
						Version: "master",
						Src:     "motd",
						Dest:    "roles/motd",
					},
				},
			},
		},
	}
	writer := bytes.NewBufferString("")
	err := req.Write(writer)
	if err != nil {
		t.Error(err)
	}
	if len(writer.String()) == 0 {
		t.Error("saved requirements is empty")
	}
	t.Log(writer.String())
}
