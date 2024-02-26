package providers

import (
	"context"
	"fmt"
	"io"

	"github.com/polyfire/api/utils"
)

type GoogleProvider struct{}

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

	fmt.Printf("GCS_URI: %s\n", gcsURI)

	panic("Unimplemented")
}
