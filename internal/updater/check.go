package updater

import (
	"errors"
	"fmt"
	"github.com/goccy/go-json"
	"io"
	"net/http"
	"runtime"
	"strings"
)

const (
	latestReleaseUrl = "https://seanime.rahim.app/api/release" // GitHub API host
)

func (u *Updater) GetReleaseName(version string) string {

	arch := runtime.GOARCH
	switch runtime.GOARCH {
	case "amd64":
		arch = "x86_64"
	case "arm64":
		arch = "arm64"
	case "386":
		return "i386"
	}
	oos := runtime.GOOS
	switch runtime.GOOS {
	case "linux":
		oos = "Linux"
	case "windows":
		oos = "Windows"
	case "darwin":
		oos = "MacOS"
	}

	ext := "tar.gz"
	if oos == "Windows" {
		ext = "zip"
	}

	return fmt.Sprintf("seanime-%s_%s_%s.%s", version, oos, arch, ext)
}

func (u *Updater) fetchLatestRelease() (*Release, error) {

	response, err := http.Get(latestReleaseUrl)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()

	// Check HTTP status code and errors
	statusCode := response.StatusCode

	if statusCode == 429 {
		return nil, errors.New("rate limited, try again later")
	}

	if !((statusCode >= 200) && (statusCode <= 299)) {
		err = fmt.Errorf("http error code: %d\n", statusCode)
		return nil, err
	}

	// Get byte response and http status code
	byteArr, readErr := io.ReadAll(response.Body)
	if readErr != nil {
		err = fmt.Errorf("error reading response: %s\n", readErr)
		return nil, err
	}

	// Unmarshal the byte response into a Release struct
	var res LatestReleaseResponse
	err = json.Unmarshal(byteArr, &res)
	if err != nil {
		return nil, err
	}

	res.Release.Version = strings.TrimPrefix(res.Release.TagName, "v")

	return &res.Release, nil
}
