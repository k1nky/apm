package parser

import (
	"io"
	"io/ioutil"

	"gopkg.in/yaml.v2"
)

func (r *Requirements) Read(reader io.Reader) (err error) {
	var (
		temp  *Requirements
		bytes []byte
	)

	if bytes, err = ioutil.ReadAll(reader); err != nil {
		return err
	}

	if err = yaml.Unmarshal(bytes, &temp); err != nil {
		return err
	}
	if temp != nil {
		r.Packages = temp.Packages
	}

	// TODO: validateConfig(cfg)
	return err
}

func (r *Requirements) Write(writer io.Writer) (err error) {
	var bs []byte
	if bs, err = yaml.Marshal(r); err != nil {
		return err
	}
	_, err = writer.Write(bs)
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
