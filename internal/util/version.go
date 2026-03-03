package util

import (
	"fmt"
	"net/url"
	"strconv"
	"strings"

	"github.com/Masterminds/semver/v3"
)

// IsValidBasicSemver
// e.g. "1.2.3" but not "1.2.3-beta" or "1.2"
func IsValidBasicSemver(version string) bool {
	parts := strings.Split(version, ".")
	if len(parts) != 3 {
		return false
	}

	for _, part := range parts {
		if _, err := strconv.Atoi(part); err != nil {
			return false
		}
	}

	return true
}

// CompareVersion compares two versions and returns the difference between them.
//
//	 3: Current version is newer by major version.
//	 2: Current version is newer by minor version.
//	 1: Current version is newer by patch version.
//		-3: Current version is older by major version.
//		-2: Current version is older by minor version.
//		-1: Current version is older by patch version.
func CompareVersion(current string, b string) (comp int, shouldUpdate bool) {

	currV, err := semver.NewVersion(current)
	if err != nil {
		return 0, false
	}
	otherV, err := semver.NewVersion(b)
	if err != nil {
		return 0, false
	}

	comp = currV.Compare(otherV)
	if comp == 0 {
		return 0, false
	}

	if currV.GreaterThan(otherV) {
		shouldUpdate = false

		if currV.Major() > otherV.Major() {
			comp *= 3
		} else if currV.Minor() > otherV.Minor() {
			comp *= 2
		} else if currV.Patch() > otherV.Patch() {
			comp *= 1
		}
	} else if currV.LessThan(otherV) {
		shouldUpdate = true

		if currV.Major() < otherV.Major() {
			comp *= 3
		} else if currV.Minor() < otherV.Minor() {
			comp *= 2
		} else if currV.Patch() < otherV.Patch() {
			comp *= 1
		}
	}

	return comp, shouldUpdate
}

func VersionIsOlderThan(version string, compare string) bool {
	comp, shouldUpdate := CompareVersion(version, compare)
	// shouldUpdate is false means the current version is newer
	return comp < 0 && shouldUpdate
}

var allowedGitHubOwners = []string{"5rahim"}

// validateReleaseUrl checks that the URL points to a GitHub release asset
// from an allowed owner.
func ValidateReleaseUrl(rawURL string) error {
	parsed, err := url.Parse(rawURL)
	if err != nil {
		return fmt.Errorf("malformed URL")
	}

	if parsed.Scheme != "https" {
		return fmt.Errorf("only HTTPS URLs are allowed")
	}

	switch parsed.Host {
	case "github.com":
		// e.g. https://github.com/5rahim/seanime/releases/download/v1.0.0/file.zip
		parts := strings.Split(strings.TrimPrefix(parsed.Path, "/"), "/")
		if len(parts) < 6 || parts[2] != "releases" || parts[3] != "download" {
			return fmt.Errorf("URL must point to a GitHub release asset")
		}
		owner := parts[0]
		for _, allowed := range allowedGitHubOwners {
			if strings.EqualFold(owner, allowed) {
				return nil
			}
		}
		return fmt.Errorf("repository owner %q is not allowed", owner)

	case "seanime.app":
		return nil

	default:
		return fmt.Errorf("host %q is not allowed", parsed.Host)
	}
}
