package testutils

import (
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
)

func DownloadTestCpiRelease(location string) (string, error) {
	if location == "" {
		location = "https://s3.amazonaws.com/bosh-dependencies/cpi-0%2Bdev.1.tgz"
	}

	fmt.Sprintf("Downloading test CPI release from %s", location)

	out, err := ioutil.TempFile(os.TempDir(), "cpi-release")
	defer out.Close()
	if err != nil {
		panic("Could not create temp file for test CPI release")
	}

	client := &http.Client{}
	request, err := http.NewRequest("GET", location, nil)
	if err != nil {
		return "", err
	}

	request.URL.Opaque = location

	resp, err := client.Do(request)
	defer resp.Body.Close()
	if err != nil {
		return "", err
	}

	_, err = io.Copy(out, resp.Body)
	if err != nil {
		return "", err
	}

	return out.Name(), nil
}
