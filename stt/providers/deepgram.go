package providers

import (
	"context"
	"fmt"
	"io"
	"os"

	"github.com/deepgram-devs/deepgram-go-sdk/deepgram"
)

type DeepgramProvider struct{}

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
		},
	)
	if err != nil {
		fmt.Println("ERROR", err)
		return nil, err
	}

	var text string
	words := make([]Word, 0)

	dialogue := make([]DialogueElement, 0)
	dialogueElem := DialogueElement{
		Speaker: 0,
		Text:    "",
	}

	if len(res.Results.Channels) > 0 {
		if len(res.Results.Channels[0].Alternatives) > 0 {
			text = res.Results.Channels[0].Alternatives[0].Transcript
			lastSpeaker := 0
			lastPunctatedWord := " "
			for _, word := range res.Results.Channels[0].Alternatives[0].Words {
				if (word.Speaker != nil || *(word.Speaker) != lastSpeaker) &&
					(word.SpeakerConfidence < 0.7) && rune(lastPunctatedWord[len(lastPunctatedWord)-1]) != '.' {
					speaker := lastSpeaker
					word.Speaker = &speaker
				}
				lastSpeaker = *word.Speaker
				lastPunctatedWord = word.Punctuated_Word

				if word.Speaker != nil && dialogueElem.Speaker != *(word.Speaker) {
					dialogue = append(dialogue, dialogueElem)
					dialogueElem = DialogueElement{
						Speaker: *(word.Speaker),
						Text:    "",
					}
				}

				dialogueElem.Text += " " + word.Punctuated_Word

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
	}

	response := TranscriptionResult{
		Text:     text,
		Words:    words,
		Dialogue: dialogue,
	}

	fmt.Println(dialogue)

	return &response, nil
}
