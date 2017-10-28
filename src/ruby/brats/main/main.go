package main

import (
	"archive/zip"
	"bytes"
	"fmt"
	"io"
	"io/ioutil"

	yaml "gopkg.in/yaml.v2"
)

func ModifyZipfile(path string, cb func(path string, r io.Reader) (io.Reader, error)) (string, error) {
	r, err := zip.OpenReader(path)
	if err != nil {
		return "", err
	}
	defer r.Close()

	newfile, err := ioutil.TempFile("", "buildpack.")
	if err != nil {
		return "", err
	}
	defer newfile.Close()

	zipWriter := zip.NewWriter(newfile)
	defer zipWriter.Close()

	for _, f := range r.File {
		rc, err := f.Open()
		if err != nil {
			return "", err
		}

		if f.FileInfo().IsDir() {
			// Nothing
		} else {
			header, err := zip.FileInfoHeader(f.FileInfo())
			if err != nil {
				return "", err
			}
			header.Method = zip.Deflate
			header.Name = f.Name

			writer, err := zipWriter.CreateHeader(header)
			if err != nil {
				return "", err
			}

			rc, err := cb(f.Name, rc)
			if err != nil {
				return "", err
			}

			_, err = io.Copy(writer, rc)
			if err != nil {
				return "", err
			}
		}
		rc.Close()
	}

	return newfile.Name(), nil
}

func CopyBuildpack(path string, cb func(*Manifest)) (string, error) {
	return ModifyZipfile(path, func(path string, r io.Reader) (io.Reader, error) {
		if path == "manifest.yml" {
			if r, err := ChangeManifest(r, cb); err != nil {
				return nil, err
			} else {
				return r, nil
			}
		}
		return r, nil
	})
}

type Manifest struct {
	Language        string `yaml:"language"`
	DefaultVersions []*struct {
		Name    string `yaml:"name"`
		Version string `yaml:"version"`
	} `yaml:"default_versions"`
	PrePackage                 string `yaml:"pre_package"`
	DependencyDeprecationDates []*struct {
		VersionLine string `yaml:"version_line"`
		Name        string `yaml:"name"`
		Date        string `yaml:"date"`
		Link        string `yaml:"link"`
	} `yaml:"dependency_deprecation_dates"`
	Dependencies []*struct {
		Name     string   `yaml:"name"`
		Version  string   `yaml:"version"`
		URI      string   `yaml:"uri"`
		Md5      string   `yaml:"md5,omitempty"`
		Sha256   string   `yaml:"sha256,omitempty"`
		CfStacks []string `yaml:"cf_stacks"`
	} `yaml:"dependencies"`
	IncludeFiles []string `yaml:"include_files"`
}

func ChangeManifest(r io.Reader, cb func(*Manifest)) (io.Reader, error) {
	data, err := ioutil.ReadAll(r)
	if err != nil {
		return nil, err
	}

	obj := Manifest{}
	if err := yaml.Unmarshal(data, &obj); err != nil {
		return nil, err
	}

	cb(&obj)

	if data, err := yaml.Marshal(&obj); err != nil {
		return nil, err
	} else {
		return bytes.NewReader(data), nil
	}
}

func main() {
	bp, err := CopyBuildpack("/Users/dgodd/workspace/ruby-buildpack/ruby_buildpack-v1.7.4.zip", func(obj *Manifest) {
		for _, date := range obj.DependencyDeprecationDates {
			date.Date = "2008-12-01"
		}
	})
	fmt.Println(bp, err)

	// f, err := os.Open("/Users/dgodd/workspace/ruby-buildpack/manifest.yml")
	// if err != nil {
	// 	panic(err)
	// }
	// defer f.Close()

	// m, e1 := ChangeManifest(f, func(obj *Manifest) {
	// 	for _, date := range obj.DependencyDeprecationDates {
	// 		date.Date = "2008-12-01"
	// 	}
	// })
	// m2, e2 := ioutil.ReadAll(m)
	// fmt.Println(string(m2), e1, e2)
}
