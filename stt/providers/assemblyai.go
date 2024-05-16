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

func canSpeakerChangeAssemblyAI(words []aai.TranscriptWord, index int) bool {
	currentWord := ""
	nextWord := ""

	if words[index].Text != nil {
		currentWord = *words[index].Text
	}

	if index < len(words)-1 && words[index+1].Text != nil {
		nextWord = *words[index+1].Text
	}

	return canSpeakerChange(currentWord, nextWord)
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

	transcript, err := client.Transcripts.TranscribeFromReader(
		context.Background(),
		reader,
		&params,
	)
	if err != nil {
		fmt.Println("ERROR", err)
		return nil, err
	}

	var text string
	words := make([]Word, 0)
	lastSentence := make([]Word, 0)
	sentenceSpeakersConfidence := make(map[int]float64, 0)

	text = *transcript.Text
	for i, word := range transcript.Words {
		speaker := assemblyAISpeakerToInt(word.Speaker)
		start := float64(*word.Start) / 1000.0
		end := float64(*word.End) / 1000.0

		sentenceSpeakersConfidence[speaker] += 0.8

		lastSentence = append(lastSentence, Word{
			Word:              *word.Text,
			PunctuatedWord:    *word.Text,
			Start:             start,
			End:               end,
			Confidence:        *word.Confidence,
			Speaker:           &speaker,
			SpeakerConfidence: 0.8,
		})

		if len(lastSentence) != 0 &&
			(canSpeakerChangeAssemblyAI(transcript.Words, i) || i == len(transcript.Words)-1) {
			speaker := getHighestConfidenceSpeaker(sentenceSpeakersConfidence)
			for j := range lastSentence {
				lastSentence[j].Speaker = &speaker
			}
			words = append(words, lastSentence...)
			lastSentence = make([]Word, 0)
			sentenceSpeakersConfidence = make(map[int]float64, 0)
		}
	}

	response := TranscriptionResult{
		Text:     text,
		Words:    words,
		Dialogue: wordsToDialogue(words),
	}

	return &response, nil
}
