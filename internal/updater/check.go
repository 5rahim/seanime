package updater

import (
	"errors"
	"fmt"
	"io"
	"runtime"
	"strings"

	"github.com/goccy/go-json"
)

// We fetch the latest release from the website first, if it fails we fallback to GitHub API
// This allows updates even if Seanime is removed from GitHub
var (
	websiteUrl        = "https://seanime.app/api/release"
	fallbackGithubUrl = "https://api.github.com/repos/5rahim/seanime/releases/latest"
)

type (
	GitHubResponse struct {
		Url             string `json:"url"`
		AssetsUrl       string `json:"assets_url"`
		UploadUrl       string `json:"upload_url"`
		HtmlUrl         string `json:"html_url"`
		ID              int64  `json:"id"`
		NodeID          string `json:"node_id"`
		TagName         string `json:"tag_name"`
		TargetCommitish string `json:"target_commitish"`
		Name            string `json:"name"`
		Draft           bool   `json:"draft"`
		Prerelease      bool   `json:"prerelease"`
		CreatedAt       string `json:"created_at"`
		PublishedAt     string `json:"published_at"`
		Assets          []struct {
			Url                string `json:"url"`
			ID                 int64  `json:"id"`
			NodeID             string `json:"node_id"`
			Name               string `json:"name"`
			Label              string `json:"label"`
			ContentType        string `json:"content_type"`
			State              string `json:"state"`
			Size               int64  `json:"size"`
			DownloadCount      int64  `json:"download_count"`
			CreatedAt          string `json:"created_at"`
			UpdatedAt          string `json:"updated_at"`
			BrowserDownloadURL string `json:"browser_download_url"`
		} `json:"assets"`
		TarballURL string `json:"tarball_url"`
		ZipballURL string `json:"zipball_url"`
		Body       string `json:"body"`
	}

	DocsResponse struct {
		Release Release `json:"release"`
	}

	Release struct {
		Url         string         `json:"url"`
		HtmlUrl     string         `json:"html_url"`
		NodeId      string         `json:"node_id"`
		TagName     string         `json:"tag_name"`
		Name        string         `json:"name"`
		Body        string         `json:"body"`
		PublishedAt string         `json:"published_at"`
		Released    bool           `json:"released"`
		Version     string         `json:"version"`
		Assets      []ReleaseAsset `json:"assets"`
	}
	ReleaseAsset struct {
		Url                string `json:"url"`
		Id                 int64  `json:"id"`
		NodeId             string `json:"node_id"`
		Name               string `json:"name"`
		ContentType        string `json:"content_type"`
		Uploaded           bool   `json:"uploaded"`
		Size               int64  `json:"size"`
		BrowserDownloadUrl string `json:"browser_download_url"`
	}
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
	var release *Release
	docsRelease, err := u.fetchLatestReleaseFromDocs()
	if err != nil {
		ghRelease, err := u.fetchLatestReleaseFromGitHub()
		if err != nil {
			return nil, err
		}
		release = ghRelease
	} else {
		release = docsRelease
	}

	return release, nil
}

func (u *Updater) fetchLatestReleaseFromGitHub() (*Release, error) {

	response, err := u.client.Get(fallbackGithubUrl)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()

	byteArr, readErr := io.ReadAll(response.Body)
	if readErr != nil {
		return nil, fmt.Errorf("error reading response: %w\n", readErr)
	}

	var res GitHubResponse
	err = json.Unmarshal(byteArr, &res)
	if err != nil {
		return nil, err
	}

	release := &Release{
		Url:         res.Url,
		HtmlUrl:     res.HtmlUrl,
		NodeId:      res.NodeID,
		TagName:     res.TagName,
		Name:        res.Name,
		Body:        res.Body,
		PublishedAt: res.PublishedAt,
		Released:    !res.Prerelease && !res.Draft,
		Version:     strings.TrimPrefix(res.TagName, "v"),
		Assets:      make([]ReleaseAsset, len(res.Assets)),
	}

	for i, asset := range res.Assets {
		release.Assets[i] = ReleaseAsset{
			Url:                asset.Url,
			Id:                 asset.ID,
			NodeId:             asset.NodeID,
			Name:               asset.Name,
			ContentType:        asset.ContentType,
			Uploaded:           asset.State == "uploaded",
			Size:               asset.Size,
			BrowserDownloadUrl: asset.BrowserDownloadURL,
		}
	}

	return release, nil
}

func (u *Updater) fetchLatestReleaseFromDocs() (*Release, error) {

	response, err := u.client.Get(websiteUrl)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()

	statusCode := response.StatusCode

	if statusCode == 429 {
		return nil, errors.New("rate limited, try again later")
	}

	if !((statusCode >= 200) && (statusCode <= 299)) {
		return nil, fmt.Errorf("http error code: %d\n", statusCode)
	}

	byteArr, readErr := io.ReadAll(response.Body)
	if readErr != nil {
		return nil, fmt.Errorf("error reading response: %w", readErr)
	}

	var res DocsResponse
	err = json.Unmarshal(byteArr, &res)
	if err != nil {
		return nil, err
	}

	res.Release.Version = strings.TrimPrefix(res.Release.TagName, "v")

	return &res.Release, nil
}
