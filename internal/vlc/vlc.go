package vlc

// https://github.com/CedArctic/go-vlc-ctrl/tree/master

import (
	"bytes"
	"fmt"
	"github.com/rs/zerolog"
	"io"
	"net/http"
	"strconv"
)

// VLC struct represents an http interface enabled VLC instance. Build using NewVLC()
type VLC struct {
	Host     string
	Port     int
	Password string
	BaseURL  string // combination of Host and Port
	Logger   *zerolog.Logger
}

type NewVLCOptions struct {
	Host     string
	Port     int
	Password string
	Logger   *zerolog.Logger
}

// NewVLC builds and returns a VLC struct using the Host, Port and Password of the VLC instance
func NewVLC(opts *NewVLCOptions) *VLC {

	// Form instance Base URL
	var BaseURL bytes.Buffer
	BaseURL.WriteString("http://")
	BaseURL.WriteString(opts.Host)
	BaseURL.WriteString(":")
	BaseURL.WriteString(strconv.Itoa(opts.Port))

	// Create and return instance struct
	return &VLC{
		opts.Host,
		opts.Port,
		opts.Password,
		BaseURL.String(),
		opts.Logger,
	}
}

// RequestMaker make requests to VLC using a urlSegment provided by other functions
func (vlc *VLC) RequestMaker(urlSegment string) (response string, err error) {

	// Form a GET Request
	client := &http.Client{}
	request, reqErr := http.NewRequest("GET", vlc.BaseURL+urlSegment, nil)
	if reqErr != nil {
		err = fmt.Errorf("http request error: %s\n", reqErr)
		return
	}

	// Make a GET request
	request.SetBasicAuth("", vlc.Password)
	reqResponse, resErr := client.Do(request)
	if resErr != nil {
		err = fmt.Errorf("http response error: %s\n", resErr)
		return
	}
	defer func() {
		reqResponse.Body.Close()
	}()

	// Check HTTP status code and errors
	statusCode := reqResponse.StatusCode
	if !((statusCode >= 200) && (statusCode <= 299)) {
		err = fmt.Errorf("http error code: %d\n", statusCode)
		return "", err
	}

	// Get byte response and http status code
	byteArr, readErr := io.ReadAll(reqResponse.Body)
	if readErr != nil {
		err = fmt.Errorf("error reading response: %s\n", readErr)
		return
	}

	// Write response
	response = string(byteArr)

	return
}
