package image_generation

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
)

type Bucket struct {
	BucketId string
	BaseURL  string
	APIKey   string
}

type FileUploadOptions struct {
	CacheControl string
	ContentType  string
	Update       bool
}

type FileResponse struct {
	Key string `json:"key"`
}

func DefaultFileUploadOptions() FileUploadOptions {
	return FileUploadOptions{
		CacheControl: "3600",
		ContentType:  "text/plain;charset=UTF-8",
		Update:       false,
	}
}

func (f *Bucket) Upload(path string, data io.Reader, opts FileUploadOptions) (string, error) {
	body := bufio.NewReader(data)
	_path := f.BucketId + "/" + path
	client := &http.Client{}

	var method string

	if opts.Update {
		method = http.MethodPut
	} else {
		method = http.MethodPost
	}

	reqURL := fmt.Sprintf("%s/storage/v1/object/%s", f.BaseURL, _path)
	req, err := http.NewRequest(method, reqURL, body)
	if err != nil {
		return "", err
	}

	req.Header.Set("Authorization", "Bearer "+f.APIKey)
	req.Header.Set("cache-control", opts.CacheControl)
	req.Header.Set("content-type", opts.ContentType)
	req.Header.Set("x-upsert", strconv.FormatBool(opts.Update))

	res, err := client.Do(req)
	if err != nil {
		return "", err
	}

	resBody, err := io.ReadAll(res.Body)
	if err != nil {
		return "", err
	}

	var response FileResponse
	if err = json.Unmarshal(resBody, &response); err != nil {
		return "", err
	}

	return fmt.Sprintf("%s/storage/v1/object/public/%s", f.BaseURL, response.Key), nil
}
