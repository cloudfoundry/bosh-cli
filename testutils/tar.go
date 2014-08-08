package testutils

import (
	"archive/tar"
	"io"
	"os"

	boshsys "github.com/cloudfoundry/bosh-agent/system"
)

type TarFileContent struct {
	Name, Body string
}

func GenerateTarfile(fs boshsys.FileSystem, tarFileContents []TarFileContent, tarFilePath string) error {
	tarFile, err := os.OpenFile(tarFilePath, os.O_WRONLY|os.O_CREATE, os.ModePerm)
	if err != nil {
		return err
	}

	err = GenerateTar(tarFile, tarFileContents)
	if err != nil {
		return err
	}

	err = tarFile.Close()
	if err != nil {
		return err
	}

	return nil
}

func GenerateTar(writer io.Writer, tarFileContents []TarFileContent) error {
	tarWriter := tar.NewWriter(writer)

	for _, tarFileContent := range tarFileContents {
		hdr := &tar.Header{
			Name: tarFileContent.Name,
			Size: int64(len(tarFileContent.Body)),
			Mode: 0644,
		}

		err := tarWriter.WriteHeader(hdr)
		if err != nil {
			return err
		}

		_, err = tarWriter.Write([]byte(tarFileContent.Body))
		if err != nil {
			return err
		}
	}

	err := tarWriter.Close()
	if err != nil {
		return err
	}

	return nil
}
