package utility

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"mime/multipart"
	"net/http"
	"net/textproto"
	"os"
	"path/filepath"
	"strings"
	"time"
)

var ServerToken = ""

func HTTPPostRequest(url string, contentType string, post []byte) ([]byte, int, error) {

	return HTTPPostRequestInternal(url, contentType, post, 15, nil)
}

func HTTPPostRequestAuth(url string, contentType string, post []byte, token string) ([]byte, int, error) {

	headers := &map[string]string{}
	if token != "" {
		(*headers)["Authorization"] = token
	}

	return HTTPPostRequestInternal(url, contentType, post, 15, headers)
}

func HTTPPostRequestInternal(url string, contentType string, post []byte, timeout int, headers *map[string]string) ([]byte, int, error) {

	httpClient := &http.Client{Timeout: time.Second * time.Duration(timeout)}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(post))
	if err != nil {
		return nil, 0, err
	}
	req.Header.Set("Content-Type", contentType)
	if headers != nil {
		for key, value := range *headers {
			// Use this to avoid golang change key cases.
			req.Header[key] = []string{value}
		}
	}

	resp, err := httpClient.Do(req)
	//resp, err := httpClient.Post(url, contentType, bytes.NewBuffer(post))
	if err != nil {
		log.Println(req.Header)
		log.Println("failed to request origin: ", err)
		return nil, 0, err
	}
	defer resp.Body.Close()

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Println("failed to read response: ", err)
		return nil, 0, fmt.Errorf("failed to read response")
	}

	if resp.StatusCode >= 400 {
		return nil, resp.StatusCode, fmt.Errorf("HTTP error code: %d Response: %s", resp.StatusCode, string(data))
	}

	return data, resp.StatusCode, nil
}

func HTTPGetRequest(url string) ([]byte, error) {

	return HTTPGetRequestInternal(url, nil)
}

func HTTPGetRequestAuth(url string, token string) ([]byte, error) {

	headers := &map[string]string{}
	if token != "" {
		(*headers)["Authorization"] = token
	}

	return HTTPGetRequestInternal(url, headers)
}

func HTTPGetRequestInternal(url string, headers *map[string]string) ([]byte, error) {

	httpClient := &http.Client{Timeout: time.Second * 15}

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	if headers != nil {
		for key, value := range *headers {
			// Use this to avoid golang change key cases.
			req.Header[key] = []string{value}
		}
	}

	resp, err := httpClient.Do(req)
	//resp, err := httpClient.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("HTTP error code: %d", resp.StatusCode)
	}

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return data, nil
}

// Set the Content-Type of upload file.
func CreateFormFile(fieldname, filename string) textproto.MIMEHeader {

	var quoteEscaper = strings.NewReplacer("\\", "\\\\", `"`, "\\\"")
	ext := filepath.Ext(filename)
	mime := GetContentType(strings.ReplaceAll(ext, ".", ""))

	h := make(textproto.MIMEHeader)
	h.Set("Content-Disposition",
		fmt.Sprintf(`form-data; name="%s"; filename="%s"`,
			quoteEscaper.Replace(fieldname), quoteEscaper.Replace(filename)))
	h.Set("Content-Type", mime)
	return h
}

func HTTPUploadFile(filename, url string, fieldname string) ([]byte, error) {

	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	body := &bytes.Buffer{}
	mw := multipart.NewWriter(body)
	fw, err := mw.CreatePart(CreateFormFile(fieldname, filename))
	if err != nil {
		return nil, err
	}

	_, err = io.Copy(fw, file)
	if err != nil {
		return nil, err
	}

	contentType := mw.FormDataContentType()
	err = mw.Close()
	if err != nil {
		return nil, err
	}

	resp, err := http.DefaultClient.Post(url, contentType, body)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("HTTP error code: %d", resp.StatusCode)
	}

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return data, nil
}

func HTTPGetFile(url string, file string) error {

	httpClient := &http.Client{Timeout: time.Second * 15}
	resp, err := httpClient.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return fmt.Errorf("HTTP error code: %d", resp.StatusCode)
	}

	writer, err := os.OpenFile(file, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0664)
	if err != nil {
		return err
	}
	defer writer.Close()

	_, err = io.Copy(writer, resp.Body)
	if err != nil {
		return err
	}

	return nil
}

func HTTPDeleteFile(url string) ([]byte, error) {

	req, err := http.NewRequest(http.MethodDelete, url, nil)
	if err != nil {
		return nil, err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("HTTP error code: %d", resp.StatusCode)
	}

	data, err := ioutil.ReadAll(resp.Body)
	return data, err
}

func PostRequest(url string, request interface{}) (string, int, error) {

	return PostRequestAuth(url, request, "")
}

func PostRequestAuth(url string, request interface{}, token string) (string, int, error) {

	data, err := json.Marshal(request)
	if err != nil {
		return "", 0, err
	}

	out, code, err := HTTPPostRequestAuth(url, "application/json", data, token)
	if err != nil {
		return "", code, err
	}
	if code >= 400 {
		return "", code, fmt.Errorf(string(out))
	}

	return string(out), code, nil
	/*
		err = json.Unmarshal(out, response)
		if err != nil {
			return 0, err
		}

		return code, nil
	*/
}
