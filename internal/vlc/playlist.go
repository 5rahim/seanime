package vlc

import "github.com/goccy/go-json"

// Node structure (node or leaf type) is the basic element of VLC's playlist tree representation.
// Leafs are playlist items. Nodes are playlists or folders inside playlists.
type Node struct {
	Ro       string `json:"ro"`
	Type     string `json:"type"` // node or leaf
	Name     string `json:"name"`
	ID       string `json:"id"`
	Duration int    `json:"duration,omitempty"`
	URI      string `json:"uri,omitempty"`
	Current  string `json:"current,omitempty"`
	Children []Node `json:"children,omitempty"`
}

// ParsePlaylist parses Playlist() responses to Node
func ParsePlaylist(playlistResponse string) (playlist Node, err error) {
	err = json.Unmarshal([]byte(playlistResponse), &playlist)
	if err != nil {
		return
	}
	return
}

// Playlist returns a Node object that is the root node of VLC's Playlist tree
// Playlist tree structure: Level 0 - Root Node (Type="node"), Level 1 - Playlists (Type="node"),
// Level 2+: Playlist Items (Type="leaf") or Folder (Type="node")
func (vlc *VLC) Playlist() (playlist Node, err error) {
	// Make response and check for errors
	response, err := vlc.RequestMaker("/requests/playlist.json")
	if err != nil {
		return
	}
	// Parse to node
	playlist, err = ParsePlaylist(response)
	return
}
