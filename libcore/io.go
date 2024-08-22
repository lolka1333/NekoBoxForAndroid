// Start of Selection
package libcore

import (
	"archive/zip"
	"io"
	"os"
	"path/filepath"

	"github.com/ulikunitz/xz"
)

// Unxz распаковывает xz-архив в указанный путь
func Unxz(archivePath, destinationPath string) error {
	inputFile, err := os.Open(archivePath)
	if err != nil {
		return err
	}
	defer inputFile.Close()

	xzReader, err := xz.NewReader(inputFile)
	if err != nil {
		return err
	}

	outputFile, err := os.Create(destinationPath)
	if err != nil {
		return err
	}
	defer outputFile.Close()

	_, err = io.Copy(outputFile, xzReader)
	return err
}

// Unzip распаковывает zip-архив в указанный путь
func Unzip(archivePath, destinationPath string) error {
	zipReader, err := zip.OpenReader(archivePath)
	if err != nil {
		return err
	}
	defer zipReader.Close()

	if err := os.MkdirAll(destinationPath, os.ModePerm); err != nil {
		return err
	}

	for _, file := range zipReader.File {
		filePath := filepath.Join(destinationPath, file.Name)

		if file.FileInfo().IsDir() {
			if err := os.MkdirAll(filePath, os.ModePerm); err != nil {
				return err
			}
			continue
		}

		if err := extractFile(file, filePath); err != nil {
			return err
		}
	}

	return nil
}

// extractFile извлекает файл из zip-архива
func extractFile(file *zip.File, filePath string) error {
	newFile, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer newFile.Close()

	zipFile, err := file.Open()
	if err != nil {
		return err
	}
	defer zipFile.Close()

	if _, err := io.Copy(newFile, zipFile); err != nil {
		return err
	}

	return nil
}
