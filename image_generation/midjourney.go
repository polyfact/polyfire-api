package image_generation

import (
	"bytes"
	"encoding/json"
	"fmt"
	"image"
	"image/png"
	"io"
	"net/http"
	"os"
	"strings"
	"time"
)

type MJImagineRequest struct {
	Prompt string `json:"prompt"`
}

type MJImagineResponse struct {
	TaskID string `json:"taskId"`
}

func MJImagine(prompt string) (string, error) {
	client := http.DefaultClient

	req_body_json, err := json.Marshal(MJImagineRequest{Prompt: prompt})
	if err != nil {
		return "", err
	}

	req, err := http.NewRequest(
		"POST",
		"https://api.midjourneyapi.io/v2/imagine",
		strings.NewReader(string(req_body_json)),
	)
	if err != nil {
		return "", err
	}

	req.Header.Add("Authorization", os.Getenv("SLASHIMAGINE_API_KEY"))
	req.Header.Add("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}

	body_resp, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	var imagine_response MJImagineResponse
	json.Unmarshal(body_resp, &imagine_response)

	return imagine_response.TaskID, nil
}

type MJResultRequest struct {
	TaskID   string `json:"taskId"`
	Position int    `json:"position,omitempty"`
}

type MJResultResponse struct {
	Status     *string `json:"status"`
	Percentage *int    `json:"percentage"`
	ImageURL   *string `json:"imageURL"`
}

func MJResult(input MJResultRequest) (*MJResultResponse, error) {
	client := http.DefaultClient

	req_body_json, err := json.Marshal(input)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest(
		"POST",
		"https://api.midjourneyapi.io/v2/result",
		strings.NewReader(string(req_body_json)),
	)
	if err != nil {
		return nil, err
	}

	req.Header.Add("Authorization", os.Getenv("SLASHIMAGINE_API_KEY"))
	req.Header.Add("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	body_resp, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var result_response MJResultResponse
	json.Unmarshal(body_resp, &result_response)

	return &result_response, nil
}

type SubImager interface {
	SubImage(r image.Rectangle) image.Image
}

func MJGenerate(prompt string) (io.Reader, error) {
	taskId, err := MJImagine(prompt)
	if err != nil {
		return nil, err
	}
	fmt.Printf("TaskID: %v\n", taskId)

	var imageURL string
	for {
		result, err := MJResult(MJResultRequest{TaskID: taskId})
		if err != nil {
			return nil, err
		}
		if result.ImageURL != nil {
			imageURL = *result.ImageURL
			break
		}
		time.Sleep(3 * time.Second)
	}

	resp, err := http.Get(imageURL)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	originalImage, _, err := image.Decode(resp.Body)
	if err != nil {
		return nil, err
	}

	bounds := originalImage.Bounds()
	width := bounds.Dx()
	height := bounds.Dy()
	cropSize := image.Rect(0, 0, width/2, height/2)
	croppedImage, ok := originalImage.(SubImager)
	if !ok {
		return nil, fmt.Errorf("image does not support sub-imaging")
	}

	var b bytes.Buffer
	if err := png.Encode(&b, croppedImage.SubImage(cropSize)); err != nil {
		return nil, err
	}

	return bytes.NewReader(b.Bytes()), nil
}

