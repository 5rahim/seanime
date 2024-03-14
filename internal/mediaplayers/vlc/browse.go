package vlc

import "github.com/goccy/go-json"

// File struct represents a single item in the browsed directory. Can be a file or a dir
type File struct {
	Type             string `json:"type"` // file or dir
	Path             string `json:"path"`
	Name             string `json:"name"`
	AccessTime       uint   `json:"access_time"`
	UID              uint   `json:"uid"`
	CreationTime     uint   `json:"creation_time"`
	GID              uint   `json:"gid"`
	ModificationTime uint   `json:"modification_time"`
	Mode             uint   `json:"mode"`
	URI              string `json:"uri"`
	Size             uint   `json:"size"`
}

// ParseBrowse parses Browse() responses to []File
func ParseBrowse(browseResponse string) (files []File, err error) {
	var temp struct {
		Files []File `json:"element"`
	}
	err = json.Unmarshal([]byte(browseResponse), &temp)
	files = temp.Files
	return
}

// Browse returns a File array with the items of the provided directory URI
func (vlc *VLC) Browse(uri string) (files []File, err error) {
	var response string
	response, err = vlc.RequestMaker("/requests/browse.json?uri=" + uri)
	if err != nil {
		return
	}
	files, err = ParseBrowse(response)
	return
}
