package providers

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/polyfire/api/llm/providers/options"
	"github.com/polyfire/api/tokens"
)

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

func (m ReplicateProvider) Stream(
	task string,
	c options.ProviderCallback,
	opts *options.ProviderOptions,
) chan options.Result {
	chanRes := make(chan options.Result)

	go func() {
		defer close(chanRes)
		tokenUsage := options.TokenUsage{Input: 0, Output: 0}

		replicateStartTime := time.Now()

		startResponse, errorCode := m.ReplicateStart(task, opts, true)
		if errorCode != "" {
			chanRes <- options.Result{Err: errorCode}
			return
		}

		req, err := http.NewRequest("GET", startResponse.URLs.Stream, nil)
		if err != nil {
			chanRes <- options.Result{Err: "generation_error"}
			return
		}

		req.Header.Set("Authorization", "Token "+m.ReplicateAPIKey)
		req.Header.Set("Accept", "text/event-stream")

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			chanRes <- options.Result{Err: "generation_error"}
			return
		}

		tokenUsage.Input += tokens.CountTokens(task)
		totalOutputTokens := 0

		totalCompletion := ""
		p := make([]byte, 128)
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

				tokenUsage.Output = tokens.CountTokens(result)
				totalOutputTokens += tokenUsage.Output

				chanRes <- options.Result{Result: result, TokenUsage: tokenUsage}
			}
		}

		replicateEndTime := time.Now()

		duration := replicateEndTime.Sub(replicateStartTime)

		if c != nil {
			credits := int(duration.Seconds()*m.CreditsPerSecond) + 1
			(*c)("replicate", m.Model, tokenUsage.Input, totalOutputTokens, totalCompletion, &credits)
		}
	}()

	return chanRes
}
