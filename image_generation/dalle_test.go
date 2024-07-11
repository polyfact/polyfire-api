package imagegeneration

import (
	"context"
	"crypto/sha256"
	"fmt"
	"io"
	"log"
	"testing"

	"github.com/polyfire/api/utils"
)

func TestOpenAIProvider(t *testing.T) {
	utils.SetLogLevel("WARN")
	ctx := utils.MockOpenAIServer(context.Background())
	reader, err := DALLEGenerate(ctx, "A red circle", "test-model")
	if err != nil {
		t.Fatalf(`DALLEGenerate returned an error "%v"`, err)
	}
	h := sha256.New()
	if _, err := io.Copy(h, reader); err != nil {
		log.Fatal(err)
	}
	if fmt.Sprintf(
		"%x",
		h.Sum(nil),
	) != "8d0171c19f9aae986bf4526226ba8d943447af59653d0cd1d8d89e605241f890" {
		t.Fatalf("Image returned doesn't match checksum of red circle test image")
	}
}
