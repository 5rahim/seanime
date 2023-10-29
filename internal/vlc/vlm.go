package vlc

import "net/url"

// Vlm returns the full list of VLM elements
func (vlc *VLC) Vlm() (response string, err error) {
	response, err = vlc.RequestMaker("/requests/vlm.xml")
	return
}

// VlmCmd executes a VLM Command and returns the response. Command is internally URL percent-encoded
func (vlc *VLC) VlmCmd(cmd string) (response string, err error) {
	response, err = vlc.RequestMaker("/requests/vlm_cmd.xml?command=" + url.QueryEscape(cmd))
	return
}

// VlmCmdErr returns the last VLM Error
func (vlc *VLC) VlmCmdErr() (response string, err error) {
	response, err = vlc.RequestMaker("/requests/vlm_cmd.xml")
	return
}
