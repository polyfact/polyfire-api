package providers

import (
	"context"
	"fmt"
	"io"
	"os"

	"github.com/deepgram-devs/deepgram-go-sdk/deepgram"
)

type DeepgramProvider struct{}

func getHighestConfidenceSpeaker(speakersConfidence map[int]float64) int {
	maxKey := -1
	maxValue := -1.0
	for k, v := range speakersConfidence {
		if v > maxValue {
			maxKey = k
			maxValue = v
		}
	}
	return maxKey
}

func (DeepgramProvider) Transcribe(
	_ context.Context,
	reader io.Reader,
	opts TranscriptionInputOptions,
) (*TranscriptionResult, error) {
	credentials := os.Getenv("DEEPGRAM_API_KEY")
	dg := deepgram.NewClient(credentials)

	language := "en-US"
	if opts.Language != nil {
		language = *(opts.Language)
	}

	res, err := dg.PreRecordedFromStream(
		deepgram.ReadStreamSource{
			Stream:   reader,
			Mimetype: "audio/mp3",
		},
		deepgram.PreRecordedTranscriptionOptions{
			Punctuate:  true,
			Diarize:    true,
			Language:   language,
			Utterances: true,
			Model:      "nova-2",
			Keywords:   opts.Keywords,
		},
	)
	if err != nil {
		fmt.Println("ERROR", err)
		return nil, err
	}

	var text string
	words := make([]Word, 0)

	dialogue := make([]DialogueElement, 0)
	sentenceSpeakersConfidence := make(map[int]float64, 0)
	dialogueElem := DialogueElement{
		Speaker: 0,
		Text:    "",
		Start:   0,
		End:     0,
	}
	lastSentence := DialogueElement{
		Speaker: 0,
		Text:    "",
		Start:   0,
		End:     0,
	}

	if len(res.Results.Channels) > 0 && len(res.Results.Channels[0].Alternatives) > 0 {
		text = res.Results.Channels[0].Alternatives[0].Transcript
		for _, word := range res.Results.Channels[0].Alternatives[0].Words {
			if lastSentence.Start == 0 {
				lastSentence.Start = word.Start
			}

			if word.Speaker != nil {
				sentenceSpeakersConfidence[*word.Speaker] += word.SpeakerConfidence
			}

			lastSentence.Text += " " + word.Punctuated_Word
			lastSentence.End = word.End

			if rune(word.Punctuated_Word[len(word.Punctuated_Word)-1]) == '.' {
				if dialogueElem.Start == 0 {
					dialogueElem = lastSentence
				} else {
					speaker := getHighestConfidenceSpeaker(sentenceSpeakersConfidence)
					lastSentence.Speaker = speaker
					if dialogueElem.Speaker == lastSentence.Speaker {
						dialogueElem.Text = dialogueElem.Text + " " + lastSentence.Text
						dialogueElem.End = lastSentence.End
					} else {
						dialogue = append(dialogue, dialogueElem)
						dialogueElem = lastSentence
					}
				}
				sentenceSpeakersConfidence = make(map[int]float64, 0)
				lastSentence = DialogueElement{
					Speaker: 0,
					Text:    "",
					Start:   0,
					End:     0,
				}
			}

			words = append(words, Word{
				Word:              word.Word,
				PunctuatedWord:    word.Punctuated_Word,
				Start:             word.Start,
				End:               word.End,
				Confidence:        word.Confidence,
				Speaker:           word.Speaker,
				SpeakerConfidence: word.SpeakerConfidence,
			})
		}
		dialogue = append(dialogue, dialogueElem)
	}

	response := TranscriptionResult{
		Text:     text,
		Words:    words,
		Dialogue: dialogue,
	}

	return &response, nil
}
