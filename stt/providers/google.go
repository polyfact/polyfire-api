package providers

import (
	"context"
	"io"
	"log"
	"strings"

	speech "cloud.google.com/go/speech/apiv1p1beta1"
	speechpb "cloud.google.com/go/speech/apiv1p1beta1/speechpb"
	"github.com/polyfire/api/utils"
	wrapperspb "google.golang.org/protobuf/types/known/wrapperspb"
)

type GoogleProvider struct{}

func removePunctuation(word string) string {
	return strings.ToLower(strings.Trim(word, ".,:;"))
}

func (GoogleProvider) Transcribe(
	ctx context.Context,
	reader io.Reader,
	opts TranscriptionInputOptions,
) (*TranscriptionResult, error) {
	gcs := ctx.Value(utils.ContextKeyGCS).(utils.GCSUploader)
	eventID := ctx.Value(utils.ContextKeyEventID).(string)

	gcsFilePath := "transcription-audio/transcribe-" + eventID

	gcsURI, err := gcs.UploadFile(reader, gcsFilePath)
	if err != nil {
		return nil, err
	}

	bgCtx := context.Background()

	// Instantiates a client
	client, err := speech.NewClient(bgCtx)
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}
	defer client.Close()

	if opts.Language == nil {
		language := "en-US"
		opts.Language = &language
	}

	diarizationConfig := speechpb.SpeakerDiarizationConfig{
		EnableSpeakerDiarization: true,
		MinSpeakerCount:          2,
		MaxSpeakerCount:          30,
	}

	config := speechpb.RecognitionConfig{
		Model:                   "latest_long",
		Encoding:                speechpb.RecognitionConfig_MP3,
		SampleRateHertz:         48000,
		AudioChannelCount:       1,
		EnableWordTimeOffsets:   true,
		EnableWordConfidence:    true,
		EnableSpokenPunctuation: wrapperspb.Bool(true),
		LanguageCode:            *opts.Language,
		DiarizationConfig:       &diarizationConfig,
	}

	audio := speechpb.RecognitionAudio{
		AudioSource: &speechpb.RecognitionAudio_Uri{Uri: gcsURI},
	}

	request := speechpb.LongRunningRecognizeRequest{
		Config: &config,
		Audio:  &audio,
	}

	op, err := client.LongRunningRecognize(bgCtx, &request)
	if err != nil {
		log.Fatalf("failed to recognize: %v", err)
	}
	resp, err := op.Wait(bgCtx)
	if err != nil {
		log.Fatalf("failed to wait for long-running operation: %v", err)
	}

	text := ""
	words := make([]Word, 0)

	dialogue := make([]DialogueElement, 0)
	dialogueElem := DialogueElement{
		Speaker: 0,
		Text:    "",
		Start:   0,
		End:     0,
	}

	for _, result := range resp.Results {
		if len(result.Alternatives) > 0 {
			alt := result.Alternatives[0]
			text += alt.Transcript
			for _, word := range alt.Words {
				speaker := int(word.SpeakerTag)
				if speaker == 0 {
					continue
				}
				if dialogueElem.Start == 0 {
					dialogueElem.Start = word.StartTime.AsDuration().Seconds()
					dialogueElem.Speaker = speaker
				}

				if dialogueElem.Speaker != speaker {
					dialogue = append(dialogue, dialogueElem)
					dialogueElem = DialogueElement{
						Speaker: speaker,
						Text:    "",
						Start:   word.StartTime.AsDuration().Seconds(),
					}
				}

				dialogueElem.Text += " " + word.Word
				dialogueElem.End = word.EndTime.AsDuration().Seconds()

				words = append(words, Word{
					Word:           removePunctuation(word.Word),
					PunctuatedWord: word.Word,
					Start:          word.StartTime.AsDuration().Seconds(),
					End:            word.EndTime.AsDuration().Seconds(),
					Confidence:     float64(word.Confidence),
					Speaker:        &speaker,

					// Google doesn't provide a value for this so we assume it's 80% correct
					SpeakerConfidence: 0.8,
				})
			}
		}
	}

	dialogue = append(dialogue, dialogueElem)

	response := TranscriptionResult{
		Text:     text,
		Words:    words,
		Dialogue: dialogue,
	}

	return &response, nil
}
