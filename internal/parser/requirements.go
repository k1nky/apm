package parser

import (
	"io"
	"io/ioutil"

	"gopkg.in/yaml.v2"
)

func Load(r io.Reader) (*Requirements, error) {
	var (
		requirements *Requirements
		err          error
		bytes        []byte
	)

	if bytes, err = ioutil.ReadAll(r); err != nil {
		return nil, err
	}

	if err = yaml.Unmarshal(bytes, &requirements); err != nil {
		return nil, err
	}
	if requirements == nil {
		requirements = &Requirements{}
	}

	// TODO: validateConfig(cfg)
	return requirements, err
}

func Save(w io.Writer, r *Requirements) (err error) {
	var bs []byte
	if bs, err = yaml.Marshal(r); err != nil {
		return err
	}
	_, err = w.Write(bs)
	return
}

func (r *Requirements) SearchByUrl(url string) int {
	for k, v := range r.Packages {
		if v.Src == url {
			return k
		}
	}
	return -1
}

func (r *Requirements) SearchByMapping(url string, m Mapping) int {
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

func (r *Requirements) Add(p Package) {
	urlIndex := r.SearchByUrl(p.Src)

	if urlIndex == -1 {
		r.Packages = append(r.Packages, p)
		return
	}
	for _, v := range p.Mappings {
		mappingIndex := r.SearchByMapping(p.Src, v)
		if mappingIndex == -1 {
			r.Packages[urlIndex].Mappings = append(r.Packages[urlIndex].Mappings, v)
		} else {
			r.Packages[urlIndex].Mappings[mappingIndex] = v
		}
	}
}
