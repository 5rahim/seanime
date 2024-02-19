package updater

import (
	"errors"
	"fmt"
	"github.com/goccy/go-json"
	"io"
	"net/http"
	"runtime"
	"strconv"
	"strings"
)

const (
	latestReleaseUrl = "https://seanime.rahim.app/api/release" // GitHub API host
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

	res.Release.Version = strings.TrimPrefix(res.Release.TagName, "v")

	return &res.Release, nil
}

// compareVersion compares current and latest version is returns true if the latest version is newer than the current version.
// It also returns the update type (patch, minor, major) if the latest version is newer than the current version.
func (u *Updater) compareVersion(currVersion string, tagName string) (string, bool) {
	tagName = strings.TrimPrefix(tagName, "v")

	currVParts := strings.Split(currVersion, ".")
	latestVParts := strings.Split(tagName, ".")

	if len(currVParts) != 3 || len(latestVParts) != 3 {
		return "", false
	}

	currMajor, _ := strconv.Atoi(currVParts[0])
	currMinor, _ := strconv.Atoi(currVParts[1])
	currPatch, _ := strconv.Atoi(currVParts[2])

	latestMajor, _ := strconv.Atoi(latestVParts[0])
	latestMinor, _ := strconv.Atoi(latestVParts[1])
	latestPatch, _ := strconv.Atoi(latestVParts[2])

	if currMajor > latestMajor {
		return "", false
	}

	if currMajor < latestMajor {
		return MajorRelease, true
	}

	if currMinor > latestMinor {
		return "", false
	}

	if currMinor < latestMinor {
		return MinorRelease, true
	}

	if currPatch > latestPatch {
		return "", false
	}

	if currPatch < latestPatch {
		return PatchRelease, true
	}

	return "", false
}
