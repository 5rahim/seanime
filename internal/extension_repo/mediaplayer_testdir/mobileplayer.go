package mediaplayer_testdir

//import (
//	"fmt"
//	"strings"
//
//	hibikemediaplayer "seanime/internal/extension/hibike/mediaplayer"
//)
//
//type (
//	// MobilePlayer is an extension that sends media links the mobile device's media player.
//	MobilePlayer struct {
//		config mobilePlayerConfig
//	}
//
//	mobilePlayerConfig struct {
//		iosPlayer     string
//		androidPlayer string
//	}
//)
//
//func NewMediaPlayer() hibikemediaplayer.MediaPlayer {
//	return &MobilePlayer{}
//}
//
//func (m *MobilePlayer) InitConfig(config map[string]interface{}) {
//	iosPlayer, _ := config["iosPlayer"].(string)
//	androidPlayer, _ := config["androidPlayer"].(string)
//
//	m.config = mobilePlayerConfig{
//		iosPlayer:     iosPlayer,
//		androidPlayer: androidPlayer,
//	}
//}
//
//func (m *MobilePlayer) GetSettings() hibikemediaplayer.Settings {
//	return hibikemediaplayer.Settings{
//		CanTrackProgress: false,
//	}
//}
//
//func (m *MobilePlayer) Play(req hibikemediaplayer.PlayRequest) (*hibikemediaplayer.PlayResponse, error) {
//	return m.getPlayResponse(req)
//}
//
//func (m *MobilePlayer) Stream(req hibikemediaplayer.PlayRequest) (*hibikemediaplayer.PlayResponse, error) {
//	return m.getPlayResponse(req)
//}
//
//func (m *MobilePlayer) getPlayResponse(req hibikemediaplayer.PlayRequest) (*hibikemediaplayer.PlayResponse, error) {
//	var url string
//	if req.ClientInfo.Platform == "ios" {
//		// Play on iOS
//		switch m.config.iosPlayer {
//		case "outplayer":
//			url = getOutplayerUrl(req.Path)
//		}
//	}
//
//	if url == "" {
//		return nil, fmt.Errorf("no player found for platform %s", req.ClientInfo.Platform)
//	}
//
//	return &hibikemediaplayer.PlayResponse{
//		OpenURL: url,
//	}, nil
//}
//
//func getOutplayerUrl(url string) (ret string) {
//	ret = strings.Replace(url, "http://", "outplayer://", 1)
//	ret = strings.Replace(ret, "https://", "outplayer://", 1)
//	return
//}
//
//func (m *MobilePlayer) GetPlaybackStatus() (*hibikemediaplayer.PlaybackStatus, error) {
//	return nil, fmt.Errorf("not implemented")
//}
//
//func (m *MobilePlayer) Start() error {
//	return nil
//}
//
//func (m *MobilePlayer) Stop() error {
//	return nil
//}
