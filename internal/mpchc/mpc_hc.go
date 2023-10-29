package mpchc

import (
	"fmt"
	"github.com/rs/zerolog"
	"io"
	"net/http"
	neturl "net/url"
)

type MpcHc struct {
	Host   string
	Port   int
	Path   string
	Logger *zerolog.Logger
}

func (api *MpcHc) url() string {
	return fmt.Sprintf("http://%s:%d", api.Host, api.Port)
}

// Execute sends a command to MPC and returns the response.
func (api *MpcHc) Execute(command int, data map[string]interface{}) (string, error) {
	url := fmt.Sprintf("%s/command.html?wm_command=%d", api.url(), command)

	if data != nil {
		queryParams := neturl.Values{}
		for key, value := range data {
			queryParams.Add(key, fmt.Sprintf("%v", value))
		}
		url += "&" + queryParams.Encode()
	}

	println(url)

	response, err := http.Get(url)
	if err != nil {
		return "", err
	}
	defer response.Body.Close()

	// Check HTTP status code and errors
	statusCode := response.StatusCode
	if !((statusCode >= 200) && (statusCode <= 299)) {
		err = fmt.Errorf("http error code: %d\n", statusCode)
		return "", err
	}

	// Get byte response and http status code
	byteArr, readErr := io.ReadAll(response.Body)
	if readErr != nil {
		err = fmt.Errorf("error reading response: %s\n", readErr)
		return "", err
	}

	// Write response
	res := string(byteArr)

	return res, nil
}

// OpenAndPlay opens a video file in MPC.
func (api *MpcHc) OpenAndPlay(filePath string) (string, error) {
	url := fmt.Sprintf("%s/browser.html?path=%s", api.url(), filePath)

	response, err := http.Get(url)
	if err != nil {
		return "", err
	}
	defer response.Body.Close()

	// Check HTTP status code and errors
	statusCode := response.StatusCode
	if !((statusCode >= 200) && (statusCode <= 299)) {
		err = fmt.Errorf("http error code: %d\n", statusCode)
		return "", err
	}

	// Get byte response and http status code
	byteArr, readErr := io.ReadAll(response.Body)
	if readErr != nil {
		err = fmt.Errorf("error reading response: %s\n", readErr)
		return "", err
	}

	// Write response
	res := string(byteArr)

	return res, nil
}

// GetVariables retrieves player variables from MPC.
func (api *MpcHc) GetVariables() (*Variables, error) {
	url := fmt.Sprintf("%s/variables.html", api.url())

	response, err := http.Get(url)
	if err != nil {
		return &Variables{}, err
	}
	defer response.Body.Close()

	// Check HTTP status code and errors
	statusCode := response.StatusCode
	if !((statusCode >= 200) && (statusCode <= 299)) {
		err = fmt.Errorf("http error code: %d\n", statusCode)
		return &Variables{}, err
	}

	// Get byte response and http status code
	byteArr, readErr := io.ReadAll(response.Body)
	if readErr != nil {
		err = fmt.Errorf("error reading response: %s\n", readErr)
		return &Variables{}, err
	}

	// Write response
	res := string(byteArr)
	vars := parseVariables(res)

	return vars, nil
}
