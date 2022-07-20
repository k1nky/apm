package parser

import (
	"io"

	"gopkg.in/yaml.v2"
)

type ReqiuredMapping struct {
	Src     string `yaml:"src"`
	Dest    string `yaml:"dest"`
	Version string `yaml:"version"`
}

type RequiredPackage struct {
	Url      string            `yaml:"src"`
	Mappings []ReqiuredMapping `yaml:"mappings"`
}

type Requirements struct {
	Packages []RequiredPackage `yaml:"packages"`
}

func (r *Requirements) Read(reader io.Reader) (err error) {
	temp := &Requirements{}

	err = yaml.NewDecoder(reader).Decode(temp)

	if temp != nil {
		r.Packages = temp.Packages
	}

	// TODO: validate
	return err
}

func (r *Requirements) Write(writer io.Writer) (err error) {
	err = yaml.NewEncoder(writer).Encode(r)
	return
}

func (r *Requirements) SearchByUrl(url string) int {
	for k, v := range r.Packages {
		if v.Url == url {
			return k
		}
	}
	return -1
}

func (r *Requirements) SearchByMapping(url string, m ReqiuredMapping) int {
	urlIndex := r.SearchByUrl(url)
	if urlIndex == -1 {
		return -1
	}
	for k, v := range r.Packages[urlIndex].Mappings {
		if v.Src == m.Src && v.Dest == m.Dest {
			return k
		}
	}
	return -1
}

func (r *Requirements) Add(p RequiredPackage) {
	urlIndex := r.SearchByUrl(p.Url)

	if urlIndex == -1 {
		r.Packages = append(r.Packages, p)
		return
	}
	for _, v := range p.Mappings {
		mappingIndex := r.SearchByMapping(p.Url, v)
		if mappingIndex == -1 {
			r.Packages[urlIndex].Mappings = append(r.Packages[urlIndex].Mappings, v)
		} else {
			r.Packages[urlIndex].Mappings[mappingIndex] = v
		}
	}
}
