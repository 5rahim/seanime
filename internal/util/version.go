package util

import (
	"strconv"
	"strings"
)

// CompareVersion compares two versions and returns the difference between them.
//
//	 3: Current version is newer by major version.
//	 2: Current version is newer by minor version.
//	 1: Current version is newer by patch version.
//		-3: Current version is older by major version.
//		-2: Current version is older by minor version.
//		-1: Current version is older by patch version.
func CompareVersion(prevVersion string, currVersion string) (int, bool) {

	prevParts := strings.Split(prevVersion, ".")
	currParts := strings.Split(currVersion, ".")

	if len(prevParts) != 3 || len(currParts) != 3 {
		return 0, false
	}

	prevMajor, _ := strconv.Atoi(prevParts[0])
	prevMinor, _ := strconv.Atoi(prevParts[1])
	prevPatch, _ := strconv.Atoi(prevParts[2])

	latestMajor, _ := strconv.Atoi(currParts[0])
	latestMinor, _ := strconv.Atoi(currParts[1])
	latestPatch, _ := strconv.Atoi(currParts[2])

	if prevMajor > latestMajor {
		return -3, false
	}

	if prevMajor < latestMajor {
		return 3, true
	}

	if prevMinor > latestMinor {
		return -2, false
	}

	if prevMinor < latestMinor {
		return 2, true
	}

	if prevPatch > latestPatch {
		return -1, false
	}

	if prevPatch < latestPatch {
		return 1, true
	}

	return 0, false
}

func VersionIsOlderThan(version string, compare string) bool {
	diff, _ := CompareVersion(version, compare)
	return diff > 0
}
