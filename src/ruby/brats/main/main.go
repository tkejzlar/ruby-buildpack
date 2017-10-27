package main

import (
	"archive/zip"
	"fmt"
	"io"
	"io/ioutil"
)

func ModifyZipfile(path string, func(path string, r io.Reader) (io.Reader, err)) (string, error) {
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
			_, err = io.Copy(writer, rc)
			if err != nil {
				return "", err
			}
		}
		rc.Close()
	}

	return newfile.Name(), nil
}

func CopyBuildpack(path string) (string, error) {
	return ModifyZipfile(path, func(path string, r io.Reader) (io.Reader, err) 
}

func main() {
	bp, err := CopyBuildpack("/home/pivotal/workspace/ruby-buildpack/ruby_buildpack-v1.7.4.zip")
	fmt.Println(bp, err)
}
