package providers

import (
	"context"
	"fmt"
	"io"
	"os"
	"strings"

	aai "github.com/AssemblyAI/assemblyai-go-sdk"
)

type AssemblyAIProvider struct{}

func assemblyAISpeakerToInt(speaker *string) int {
	if speaker == nil || len(*speaker) == 0 {
		return -1
	}
	return int([]rune(*speaker)[0] - 'A')
}

func fixLanguageCodeFormat(code string) string {
	return strings.Replace(strings.ToLower(code), "-", "_", -1)
}

func (AssemblyAIProvider) Transcribe(
	_ context.Context,
	reader io.Reader,
	opts TranscriptionInputOptions,
) (*TranscriptionResult, error) {
	client := aai.NewClient(os.Getenv("ASSEMBLYAI_API_KEY"))

	language := "en_US"
	if opts.Language != nil {
		language = fixLanguageCodeFormat(*(opts.Language))
	}

	speakerLabel := true
	fmt.Println("KEYWORDS:", opts.Keywords)
	params := aai.TranscriptOptionalParams{
		LanguageCode:  aai.TranscriptLanguageCode(language),
		SpeakerLabels: &speakerLabel,
		WordBoost:     opts.Keywords,
	}

	transcript, err := client.Transcripts.TranscribeFromReader(context.Background(), reader, &params)
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

	text = *transcript.Text
	for _, word := range transcript.Words {
		speaker := assemblyAISpeakerToInt(word.Speaker)
		start := float64(*word.Start) / 1000.0
		end := float64(*word.End) / 1000.0
		if lastSentence.Start == 0 {
			lastSentence.Start = start
		}

		sentenceSpeakersConfidence[speaker] += 0.8

		lastSentence.Text += " " + *word.Text
		lastSentence.End = end

		if rune((*word.Text)[len(*word.Text)-1]) == '.' {
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
			Word:              *word.Text,
			PunctuatedWord:    *word.Text,
			Start:             start,
			End:               end,
			Confidence:        *word.Confidence,
			Speaker:           &speaker,
			SpeakerConfidence: 0.8,
		})
	}
	dialogue = append(dialogue, dialogueElem)

	response := TranscriptionResult{
		Text:     text,
		Words:    words,
		Dialogue: dialogue,
	}

	return &response, nil
}
