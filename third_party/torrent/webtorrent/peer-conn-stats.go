//go:build !js
// +build !js

package webtorrent

func GetPeerConnStats(pc *wrappedPeerConnection) (stats webrtc.StatsReport) {
	if pc != nil {
		stats = pc.GetStats()
	}
	return
}
