package providers

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/polyfact/api/tokens"
)

type ReplicateInput struct {
	Prompt      string   `json:"prompt"`
	Temperature *float32 `json:"temperature,omitempty"`
}

type ReplicateRequestBody struct {
	Version string         `json:"version"`
	Input   ReplicateInput `json:"input"`
	Stream  bool           `json:"stream"`
}

type ReplicateStartResponse struct {
	ID     string `json:"id"`
	Status string `json:"status"`
	URLs   struct {
		Stream string `json:"stream"`
	} `json:"urls"`
}

func ReplicateStart(reqBody ReplicateRequestBody) (ReplicateStartResponse, error) {
	input, err := json.Marshal(reqBody)
	if err != nil {
		return ReplicateStartResponse{}, err
	}

	req, err := http.NewRequest("POST", "https://api.replicate.com/v1/predictions", strings.NewReader(string(input)))
	if err != nil {
		return ReplicateStartResponse{}, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Token "+os.Getenv("REPLICATE_API_KEY"))

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return ReplicateStartResponse{}, err
	}

	res, err := io.ReadAll(resp.Body)
	if err != nil {
		return ReplicateStartResponse{}, err
	}

	var startResponse ReplicateStartResponse
	err = json.Unmarshal(res, &startResponse)
	if err != nil {
		return ReplicateStartResponse{}, err
	}

	return startResponse, nil
}

type ReplicateEvent struct {
	Event string  `json:"event"`
	ID    string  `json:"id"`
	Data  *string `json:"data"`
}

func ParseReplicateEvent(str string) (ReplicateEvent, error) {
	if strings.HasPrefix(str, "data: ") {
		nextLine := strings.Index(str, "\n")
		if nextLine == -1 {
			nextLine = len(str)
		}
		event, err := ParseReplicateEvent(str[nextLine+1:])
		if err != nil {
			return ReplicateEvent{}, err
		}
		var data string
		if event.Data == nil {
			data = str[6:nextLine]
		} else {
			data = str[6:nextLine] + "\n" + *event.Data
		}
		event.Data = &data
		return event, nil
	}

	if strings.HasPrefix(str, "event: ") {
		nextLine := strings.Index(str, "\n")
		if nextLine == -1 {
			return ReplicateEvent{}, errors.New("Invalid event")
		}
		eventName := str[7:nextLine]
		event, err := ParseReplicateEvent(str[nextLine+1:])
		if err != nil {
			return ReplicateEvent{}, err
		}

		event.Event = eventName
		return event, nil
	}

	if strings.HasPrefix(str, "id: ") {
		nextLine := strings.Index(str, "\n")
		if nextLine == -1 {
			return ReplicateEvent{}, errors.New("Invalid event")
		}
		eventID := str[4:nextLine]
		event, err := ParseReplicateEvent(str[nextLine+1:])
		if err != nil {
			return ReplicateEvent{}, err
		}

		event.ID = eventID
		return event, nil
	}

	if strings.HasPrefix(str, ":") {
		return ReplicateEvent{}, nil
	}

	if strings.HasPrefix(str, "\n") {
		return ReplicateEvent{}, nil
	}

	return ReplicateEvent{}, errors.New("Invalid event \"" + str + "\"")
}

type ReplicateProvider struct {
	Model string
}

func (m ReplicateProvider) GetCreditsPerSecond() float64 {
	switch m.Model {
	case "llama-2-70b-chat":
		return 14000.0
	case "replit-code-v1-3b":
		return 11500.0
	default:
		fmt.Printf("Invalid model: %v\n", m.Model)
		return 0.0
	}
}

func (m ReplicateProvider) GetVersion() (string, error) {
	switch m.Model {
	case "llama-2-70b-chat":
		return "02e509c789964a7ea8736978a43525956ef40397be9033abf9fd2badfe68c9e3", nil
	case "replit-code-v1-3b":
		return "b84f4c074b807211cd75e3e8b1589b6399052125b4c27106e43d47189e8415ad", nil
	default:
		return "", errors.New("Invalid model")
	}
}

func (m ReplicateProvider) Generate(task string, c ProviderCallback, opts *ProviderOptions) chan Result {
	chan_res := make(chan Result)

	go func() {
		defer close(chan_res)
		tokenUsage := TokenUsage{Input: 0, Output: 0}
		version, err := m.GetVersion()
		if err != nil {
			fmt.Printf("%v\n", err)
			return
		}

		body := ReplicateRequestBody{
			Input:   ReplicateInput{Prompt: task},
			Version: version,
			Stream:  true,
		}

		if opts != nil && opts.Temperature != nil {
			body.Input.Temperature = opts.Temperature
		}

		replicateStartTime := time.Now()

		startResponse, err := ReplicateStart(body)
		if err != nil {
			fmt.Printf("%v\n", err)
			return
		}

		req, err := http.NewRequest("GET", startResponse.URLs.Stream, nil)
		if err != nil {
			fmt.Printf("%v\n", err)
			return
		}

		req.Header.Set("Authorization", "Token "+os.Getenv("REPLICATE_API_KEY"))
		req.Header.Set("Accept", "text/event-stream")

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			fmt.Printf("%v\n", err)
			return
		}

		tokenUsage.Input += tokens.CountTokens("gpt-2", task)

		totalCompletion := ""
		var p []byte = make([]byte, 128)
		eventBuffer := ""

	receiver:
		for {
			var eventString string
			for {
				if strings.Contains(eventBuffer, "\n\n") {
					eventString = eventBuffer[:strings.Index(eventBuffer, "\n\n")+2]
					eventBuffer = eventBuffer[strings.Index(eventBuffer, "\n\n")+2:]
					break
				}

				nb, err := resp.Body.Read(p)
				if errors.Is(err, io.EOF) || err != nil {
					break receiver
				}
				eventBuffer += string(p[:nb])
			}

			event, err := ParseReplicateEvent(eventString)
			if err != nil {
				fmt.Printf("%v\n", err)
				continue
			}

			if event.Event == "" {
				continue
			}

			if event.Event == "done" {
				break
			}

			if event.Event == "output" {
				if event.Data == nil {
					continue
				}

				result := *event.Data

				totalCompletion += result

				tokenUsage.Output += tokens.CountTokens("gpt-2", result)

				chan_res <- Result{Result: result, TokenUsage: tokenUsage}
			}
		}

		replicateEndTime := time.Now()

		duration := replicateEndTime.Sub(replicateStartTime)

		credits := int(duration.Seconds()*m.GetCreditsPerSecond()) + 1
		if c != nil {
			(*c)("replicate", m.Model, tokenUsage.Input, tokenUsage.Output, totalCompletion, &credits)
		}
	}()

	return chan_res
}

func (m ReplicateProvider) UserAllowed(_user_id string) bool {
	return true
}

func (m ReplicateProvider) Name() string {
	return "replicate"
}

func (m ReplicateProvider) DoesFollowRateLimit() bool {
	return true
}
