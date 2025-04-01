package store

import (
	"archive/zip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

func Patch() error {
	object := "fuge_portable.zip"

	if err := DownloadFromS3("builds/"+object, object); err != nil {
		return err
	}

	zipFile := filepath.Join(downloadDir, object)
	if err := Unzip(zipFile, downloadDir); err != nil {
		return err
	}

	if err := os.Remove(zipFile); err != nil {
		return err
	}

	return nil
}

func Unzip(src, dst string) error {
	archive, err := zip.OpenReader(src)
	if err != nil {
		panic(err)
	}
	defer archive.Close()

	for _, f := range archive.File {
		filePath := filepath.Join(dst, f.Name)

		if !strings.HasPrefix(filePath, filepath.Clean(dst)+string(os.PathSeparator)) {
			fmt.Println("invalid file path")
			return nil
		}

		if f.FileInfo().IsDir() {
			// download, upload folders skipped
			// os.MkdirAll(filePath, os.ModePerm)
			continue
		}

		fmt.Println("unzipping file ", filePath)
		if err := os.MkdirAll(filepath.Dir(filePath), os.ModePerm); err != nil {
			return err
		}

		dstFile, err := os.OpenFile(filePath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
		if err != nil {
			return err
		}

		fileInArchive, err := f.Open()
		if err != nil {
			return err
		}

		if _, err := io.Copy(dstFile, fileInArchive); err != nil {
			return err
		}

		dstFile.Close()
		fileInArchive.Close()

	}
	return nil
}
