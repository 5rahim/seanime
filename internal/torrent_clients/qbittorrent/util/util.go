package qbittorrent_util

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/textproto"
	"seanime/internal/torrent_clients/qbittorrent/model"
	"strings"
)

func GetInto(client *http.Client, target interface{}, url string, body interface{}) (err error) {
	var buffer bytes.Buffer
	if body != nil {
		if err := json.NewEncoder(&buffer).Encode(body); err != nil {
			return err
		}
	}
	r, err := http.NewRequest("GET", url, &buffer)
	if err != nil {
		return err
	}
	resp, err := client.Do(r)
	if err != nil {
		return err
	}
	defer func() {
		if err2 := resp.Body.Close(); err2 != nil {
			err = err2
		}
	}()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("invalid response status %s", resp.Status)
	}
	buf, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	if err := json.NewDecoder(bytes.NewReader(buf)).Decode(target); err != nil {
		if err2 := json.NewDecoder(strings.NewReader(`"` + string(buf) + `"`)).Decode(target); err2 != nil {
			return err
		}
	}
	return nil
}

func Post(client *http.Client, url string, body interface{}) (err error) {
	var buffer bytes.Buffer
	if err := json.NewEncoder(&buffer).Encode(body); err != nil {
		return err
	}
	r, err := http.NewRequest("POST", url, &buffer)
	if err != nil {
		return err
	}
	resp, err := client.Do(r)
	if err != nil {
		return err
	}
	defer func() {
		if err2 := resp.Body.Close(); err2 != nil {
			err = err2
		}
	}()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("invalid status %s", resp.Status)
	}
	return nil
}

func createFormFileWithHeader(writer *multipart.Writer, name, filename string, headers map[string]string) (io.Writer, error) {
	header := textproto.MIMEHeader{}
	header.Add("Content-Disposition", fmt.Sprintf(`form-data; name="%s"; filename="%s"`, name, filename))
	for key, value := range headers {
		header.Add(key, value)
	}
	return writer.CreatePart(header)
}

func PostMultipartLinks(client *http.Client, url string, options *qbittorrent_model.AddTorrentsOptions, links []string) (err error) {
	var o map[string]interface{}
	if options != nil {
		b, err := json.Marshal(options)
		if err != nil {
			return err
		}
		if err := json.Unmarshal(b, &o); err != nil {
			return err
		}
	}
	buf := bytes.Buffer{}
	form := multipart.NewWriter(&buf)
	if err := form.WriteField("urls", strings.Join(links, "\n")); err != nil {
		return err
	}
	for key, value := range o {
		if err := form.WriteField(key, fmt.Sprintf("%v", value)); err != nil {
			return err
		}
	}
	if err := form.Close(); err != nil {
		return err
	}
	req, err := http.NewRequest("POST", url, &buf)
	if err != nil {
		return err
	}
	req.Header.Add("content-type", "multipart/form-data; boundary="+form.Boundary())
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer func() {
		if err2 := resp.Body.Close(); err2 != nil {
			err = err2
		}
	}()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("invalid status %s", resp.Status)
	}
	return nil
}

func PostMultipartFiles(client *http.Client, url string, options *qbittorrent_model.AddTorrentsOptions, files map[string][]byte) (err error) {
	var o map[string]interface{}
	if options != nil {
		b, err := json.Marshal(options)
		if err != nil {
			return err
		}
		if err := json.Unmarshal(b, &o); err != nil {
			return err
		}
	}
	buf := bytes.Buffer{}
	form := multipart.NewWriter(&buf)
	for filename, file := range files {
		writer, err := createFormFileWithHeader(form, "torrents", filename, map[string]string{
			"content-type": "application/x-bittorrent",
		})
		if err != nil {
			return err
		}
		if _, err := writer.Write(file); err != nil {
			return err
		}
	}
	for key, value := range o {
		if err := form.WriteField(key, fmt.Sprintf("%v", value)); err != nil {
			return err
		}
	}
	if err := form.Close(); err != nil {
		return err
	}
	req, err := http.NewRequest("POST", url, &buf)
	if err != nil {
		return err
	}
	req.Header.Add("content-type", "multipart/form-data; boundary="+form.Boundary())
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer func() {
		if err2 := resp.Body.Close(); err2 != nil {
			err = err2
		}
	}()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("invalid status %s", resp.Status)
	}
	return nil
}

func PostWithContentType(client *http.Client, url string, body io.Reader, contentType string) (err error) {
	r, err := http.NewRequest("POST", url, body)
	if err != nil {
		return err
	}
	r.Header.Add("content-type", contentType)
	resp, err := client.Do(r)
	if err != nil {
		return err
	}
	defer func() {
		if err2 := resp.Body.Close(); err2 != nil {
			err = err2
		}
	}()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("invalid status %s", resp.Status)
	}
	return nil
}

func GetIntoWithContentType(client *http.Client, target interface{}, url string, body io.Reader, contentType string) (err error) {
	r, err := http.NewRequest("GET", url, body)
	if err != nil {
		return err
	}
	r.Header.Add("content-type", contentType)
	resp, err := client.Do(r)
	if err != nil {
		return err
	}
	defer func() {
		if err2 := resp.Body.Close(); err2 != nil {
			err = err2
		}
	}()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("invalid response status %s", resp.Status)
	}
	buf, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	if err := json.NewDecoder(bytes.NewReader(buf)).Decode(target); err != nil {
		if err2 := json.NewDecoder(strings.NewReader(`"` + string(buf) + `"`)).Decode(target); err2 != nil {
			return err
		}
	}
	return nil
}
