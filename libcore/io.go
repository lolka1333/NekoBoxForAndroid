package libcore

import (
	"archive/zip"
	"io"
	"os"
	"path/filepath"

	E "github.com/sagernet/sing/common/exceptions"
	"github.com/ulikunitz/xz"
)

func Unxz(archive string, path string) error {
	i, err := os.Open(archive)
	if err != nil {
		return E.New("error opening archive: " + err.Error())
	}
	defer i.Close()

	r, err := xz.NewReader(i)
	if err != nil {
		return E.New("error reading archive: " + err.Error())
	}

	o, err := os.Create(path)
	if err != nil {
		return E.New("error creating file: " + err.Error())
	}
	defer o.Close()

	_, err = io.Copy(o, r)
	if err != nil {
		return E.New("error copying data: " + err.Error())
	}

	return nil
}

func Unzip(archive string, path string) error {
	r, err := zip.OpenReader(archive)
	if err != nil {
		return E.New("error opening archive: " + err.Error())
	}
	defer r.Close()

	err = os.MkdirAll(path, os.ModePerm)
	if err != nil {
		return E.New("error creating directory: " + err.Error())
	}

	for _, file := range r.File {
		filePath := filepath.Join(path, file.Name)

		if file.FileInfo().IsDir() {
			err = os.MkdirAll(filePath, os.ModePerm)
			if err != nil {
				return E.New("error creating directory: " + err.Error())
			}
			continue
		}

		newFile, err := os.Create(filePath)
		if err != nil {
			return E.New("error creating file: " + err.Error())
		}
		defer newFile.Close()

		zipFile, err := file.Open()
		if err != nil {
			return E.New("error opening file in archive: " + err.Error())
		}
		defer zipFile.Close()

		_, err = io.Copy(newFile, zipFile)
		if err != nil {
			return E.New("error copying data: " + err.Error())
		}
	}

	return nil
}
