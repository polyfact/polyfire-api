package providers

import (
	"context"
	"fmt"
	"io"
	"os"
	"strings"

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

func canSpeakerChangeDeepgram(words []deepgram.WordBase, index int) bool {
	currentWord := words[index].Punctuated_Word
	nextWord := ""

	if index < len(words)-1 {
		nextWord = words[index+1].Punctuated_Word
	}

	return canSpeakerChange(currentWord, nextWord)
}

func canSpeakerChange(currentWord string, nextWord string) bool {
	if len(currentWord) > 0 && (rune(currentWord[len(currentWord)-1]) == '.' ||
		rune(currentWord[len(currentWord)-1]) == '?' ||
		rune(currentWord[len(currentWord)-1]) == '!') {
		return true
	}

	if strings.ToLower(currentWord) == "monsieur" || strings.ToLower(currentWord) == "madame" ||
		strings.ToLower(currentWord) == "m." ||
		strings.ToLower(currentWord) == "mr." ||
		strings.ToLower(currentWord) == "mme." {
		return false
	}

	if len(nextWord) > 0 && rune(nextWord[0]) >= 'A' && rune(nextWord[0]) <= 'Z' {
		return true
	}

	return false
}

func wordsToDialogue(words []Word) []DialogueElement {
	dialogue := make([]DialogueElement, 0)
	var dialogueElement DialogueElement
	text := ""
	for _, wordRaw := range words {
		speakerID := *wordRaw.Speaker

		if dialogueElement.Text == "" || dialogueElement.Speaker != speakerID {
			if text != "" {
				dialogueElement.Text = text
				dialogue = append(dialogue, dialogueElement)
				text = ""
			}

			start := wordRaw.Start
			dialogueElement = DialogueElement{
				Start:   start,
				Speaker: speakerID,
			}
		}
		end := wordRaw.End
		text += " " + wordRaw.PunctuatedWord
		dialogueElement.End = end
	}

	textToInsert := text
	dialogueElement.Text = textToInsert
	dialogue = append(dialogue, dialogueElement)

	return dialogue
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
	lastSentence := make([]Word, 0)

	sentenceSpeakersConfidence := make(map[int]float64, 0)

	if len(res.Results.Channels) > 0 && len(res.Results.Channels[0].Alternatives) > 0 {
		text = res.Results.Channels[0].Alternatives[0].Transcript
		deepgramWords := res.Results.Channels[0].Alternatives[0].Words
		for i, word := range deepgramWords {
			if word.Speaker != nil {
				sentenceSpeakersConfidence[*word.Speaker] += word.SpeakerConfidence
			}

			lastSentence = append(lastSentence, Word{
				Word:              word.Word,
				PunctuatedWord:    word.Punctuated_Word,
				Start:             word.Start,
				End:               word.End,
				Confidence:        word.Confidence,
				Speaker:           word.Speaker,
				SpeakerConfidence: word.SpeakerConfidence,
			})

			if len(lastSentence) != 0 &&
				(canSpeakerChangeDeepgram(deepgramWords, i) || i == len(deepgramWords)-1) {
				speaker := getHighestConfidenceSpeaker(sentenceSpeakersConfidence)
				for j := range lastSentence {
					lastSentence[j].Speaker = &speaker
				}
				words = append(words, lastSentence...)
				sentenceSpeakersConfidence = make(map[int]float64, 0)
			}
		}
	}

	response := TranscriptionResult{
		Text:     text,
		Words:    words,
		Dialogue: wordsToDialogue(words),
	}

	return &response, nil
}
