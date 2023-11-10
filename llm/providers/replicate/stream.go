package providers

import (
	"encoding/json"
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

type ReplicateStreamPredictionOutput struct {
	ID     string `json:"id"`
	Status string `json:"status"`
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

		startResponse, errorCode := m.ReplicateStart(task, opts, true)
		if errorCode != "" {
			chanRes <- options.Result{Err: errorCode}
			return
		}

		var replicateAfterBootTime *time.Time

		totalOutputTokens := 0
		totalCompletion := ""
		stopWords := opts.StopWords
		stopWordsCache := ""
	mainloop:
		for {
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

				if event.Event == "done" {
					break
				}

				if event.Event == "output" {
					if replicateAfterBootTime == nil {
						now := time.Now()
						replicateAfterBootTime = &now
					}

					if event.Data == nil {
						continue
					}

					stopWordsCache += *event.Data

					if stopWords != nil {
						fmt.Printf("stopWordsCache: %v\n", stopWordsCache)
						for _, stopWord := range *stopWords {
							if stopWord == strings.TrimSpace(stopWordsCache) || strings.HasPrefix(strings.TrimSpace(stopWordsCache), stopWord) {
								fmt.Println("stopWord:", stopWord, "is equal to stopWordsCache:", stopWordsCache)
								break mainloop
							}
							if strings.HasPrefix(stopWord, strings.TrimSpace(stopWordsCache)) {
								fmt.Println("stopWord:", stopWord, "has prefix stopWordsCache:", stopWordsCache)
								continue receiver
							}
						}
					}
					fmt.Println("stopWordsCache:", stopWordsCache, "is not a stop word")

					result := stopWordsCache
					stopWordsCache = ""

					totalCompletion += result

					tokenUsage.Output = tokens.CountTokens(result)
					totalOutputTokens += tokenUsage.Output

					chanRes <- options.Result{Result: result, TokenUsage: tokenUsage}
				}
			}

			req, err = http.NewRequest("GET", startResponse.URLs.Get, nil)
			if err != nil {
				fmt.Println(err)
				chanRes <- options.Result{Err: "generation_error"}
				return
			}

			req.Header.Set("Authorization", "Token "+m.ReplicateAPIKey)
			req.Header.Set("Accept", "text/event-stream")

			resp, err = http.DefaultClient.Do(req)
			if err != nil {
				fmt.Println(err)
				chanRes <- options.Result{Err: "generation_error"}
				return
			}

			respBody, err := io.ReadAll(resp.Body)
			if err != nil {
				fmt.Println(err)
				chanRes <- options.Result{Err: "generation_error"}
				return
			}

			var output ReplicateStreamPredictionOutput
			err = json.Unmarshal(respBody, &output)
			if err != nil {
				fmt.Println(err)
				chanRes <- options.Result{Err: "generation_error"}
				return
			}

			if output.Status != "starting" && output.Status != "succeeded" {
				chanRes <- options.Result{Err: "generation_error"}
				return
			}

			if output.Status == "succeeded" {
				break
			}

			fmt.Println("Waiting for model to start...", output.Status, output)
		}

		// Cancel the task
		req, err := http.NewRequest("POST", startResponse.URLs.Cancel, nil)
		if err != nil {
			fmt.Println(err)
			chanRes <- options.Result{Err: "generation_error"}
			return
		}

		req.Header.Set("Authorization", "Token "+m.ReplicateAPIKey)
		req.Header.Set("Accept", "text/event-stream")

		_, err = http.DefaultClient.Do(req)
		if err != nil {
			fmt.Println(err)
			chanRes <- options.Result{Err: "generation_error"}
			return
		}

		replicateEndTime := time.Now()

		var duration time.Duration
		if replicateAfterBootTime != nil {
			duration = replicateEndTime.Sub(*replicateAfterBootTime)
		}

		if c != nil {
			credits := int(duration.Seconds()*m.CreditsPerSecond) + 1
			(*c)("replicate", m.Model, tokenUsage.Input, totalOutputTokens, totalCompletion, &credits)
		}
	}()

	return chanRes
}
