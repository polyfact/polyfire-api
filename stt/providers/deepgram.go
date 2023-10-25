package providers

import (
	"context"
	"fmt"
	"io"
	"os"

	"github.com/deepgram-devs/deepgram-go-sdk/deepgram"
)

type DeepgramProvider struct{}

func (DeepgramProvider) Transcribe(ctx context.Context, reader io.Reader, format string) (*TranscriptionResult, error) {
	credentials := os.Getenv("DEEPGRAM_API_KEY")
	dg := deepgram.NewClient(credentials)

	res, err := dg.PreRecordedFromStream(
		deepgram.ReadStreamSource{
			Stream:   reader,
			Mimetype: "audio/mp3",
		},
		deepgram.PreRecordedTranscriptionOptions{
			Punctuate:  true,
			Diarize:    true,
			Language:   "en-US",
			Utterances: true,
		},
	)
	if err != nil {
		fmt.Println("ERROR", err)
		return nil, err
	}

	var text string
	var words []Word = make([]Word, 0)

	if len(res.Results.Channels) > 0 {
		if len(res.Results.Channels[0].Alternatives) > 0 {
			text = res.Results.Channels[0].Alternatives[0].Transcript
			for _, word := range res.Results.Channels[0].Alternatives[0].Words {
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
		}
	}

	response := TranscriptionResult{
		Text:  text,
		Words: words,
	}

	return &response, nil
}
