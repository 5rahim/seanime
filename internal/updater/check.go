package updater

import (
	"errors"
	"fmt"
	"github.com/goccy/go-json"
	"io"
	"net/http"
	"runtime"
)

const (
	latestReleaseUrl = "https://seanime.rahim.app/api/release"
)

func (u *Updater) getReleaseName(version string) string {

	arch := runtime.GOARCH
	switch runtime.GOARCH {
	case "amd64":
		arch = "x86_64"
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

	return fmt.Sprintf("seanime-%s_%s_%s", version, oos, arch)
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

	return &res.Release, nil
}
