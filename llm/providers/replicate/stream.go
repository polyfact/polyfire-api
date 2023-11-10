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

func (m ReplicateProvider) SendRequest(streamURL string) (io.ReadCloser, error) {
	req, err := http.NewRequest("GET", streamURL, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Token "+m.ReplicateAPIKey)
	req.Header.Set("Accept", "text/event-stream")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}

	return resp.Body, nil
}

type ReplicateStreamEventBuffer struct {
	buffer string
	Reader io.ReadCloser
}

var ErrEndOfStream = errors.New("End of stream")

func (eb *ReplicateStreamEventBuffer) ReadReplicateEvent() (ReplicateEvent, error) {
	p := make([]byte, 128)
	var eventString string
	for {
		if strings.Contains(eb.buffer, "\n\n") {
			eventString = eb.buffer[:strings.Index(eb.buffer, "\n\n")+2]
			eb.buffer = eb.buffer[strings.Index(eb.buffer, "\n\n")+2:]
			break
		}

		nb, err := eb.Reader.Read(p)
		if errors.Is(err, io.EOF) || err != nil {
			return ReplicateEvent{}, ErrEndOfStream
		}
		fmt.Printf("%s", p[:nb])
		eb.buffer += string(p[:nb])
	}

	return ParseReplicateEvent(eventString)
}

type StopWords struct {
	StopWords      *[]string `json:"stop_words"`
	stopWordsCache string
}

var ErrStopWordFound = errors.New("Stop word found")

func (sw *StopWords) CacheStopWords(data string) (string, error) {
	if sw.StopWords == nil {
		return data, nil
	}
	sw.stopWordsCache += data
	for _, stopWord := range *sw.StopWords {
		if stopWord == strings.TrimSpace(sw.stopWordsCache) ||
			strings.HasPrefix(strings.TrimSpace(sw.stopWordsCache), stopWord) {
			return sw.stopWordsCache, ErrStopWordFound
		}
		if strings.HasPrefix(stopWord, strings.TrimSpace(sw.stopWordsCache)) {
			return "", nil
		}
	}

	result := sw.stopWordsCache
	sw.stopWordsCache = ""
	return result, nil
}

func ReceiveStream(
	chanRes chan options.Result,
	stopWords *StopWords,
	eb *ReplicateStreamEventBuffer,
	replicateAfterBootTime **time.Time,
) (string, bool) {
	completion := ""
	for {
		event, err := eb.ReadReplicateEvent()
		if errors.Is(err, ErrEndOfStream) {
			fmt.Println("End of stream")
			return completion, false
		}
		if err != nil {
			fmt.Printf("%v\n", err)
			continue
		}

		if event.Event == "done" {
			fmt.Println("Done", event)
			return completion, true
		}

		if event.Event == "output" {
			if *replicateAfterBootTime == nil {
				now := time.Now()
				*replicateAfterBootTime = &now
			}

			if event.Data == nil {
				continue
			}

			result, err := stopWords.CacheStopWords(*event.Data)

			if errors.Is(err, ErrStopWordFound) || err != nil {
				fmt.Printf("%v\n", err)
				return completion, true
			}

			if result == "" {
				continue
			}

			completion += result

			tokenUsage := options.TokenUsage{}
			tokenUsage.Output = tokens.CountTokens(result)

			chanRes <- options.Result{Result: result, TokenUsage: tokenUsage}
		}
	}
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
		tokenUsage.Input += tokens.CountTokens(task)
		chanRes <- options.Result{TokenUsage: tokenUsage}

		startResponse, errorCode := m.ReplicateStart(task, opts, true)
		if errorCode != "" {
			chanRes <- options.Result{Err: errorCode}
			return
		}

		var replicateAfterBootTime *time.Time

		totalOutputTokens := 0
		totalCompletion := ""
		stopWords := StopWords{StopWords: opts.StopWords}

		for {
			respBody, err := m.SendRequest(startResponse.URLs.Stream)
			if err != nil {
				chanRes <- options.Result{Err: "generation_error"}
				return
			}

			eb := ReplicateStreamEventBuffer{Reader: respBody}

			completion, done := ReceiveStream(chanRes, &stopWords, &eb, &replicateAfterBootTime)
			totalCompletion += completion
			totalOutputTokens += tokens.CountTokens(completion)
			if done {
				break
			}

			respBody, err = m.SendRequest(startResponse.URLs.Get)
			if err != nil {
				fmt.Println(err)
				chanRes <- options.Result{Err: "generation_error"}
				return
			}
			var output ReplicateStreamPredictionOutput
			err = json.NewDecoder(respBody).Decode(&output)
			if err != nil || (output.Status != "starting" && output.Status != "succeeded") {
				fmt.Println(err)
				chanRes <- options.Result{Err: "generation_error"}
				return
			}

			if output.Status == "succeeded" {
				break
			}

			fmt.Println("Waiting for model to start...", output.Status, output)
		}

		_, err := m.SendRequest(startResponse.URLs.Cancel)
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
