package image_generation

import (
	"encoding/json"
	"io"
	"net/http"
	"os"

	"github.com/google/uuid"
	router "github.com/julienschmidt/httprouter"
	"github.com/polyfact/api/utils"
)

type Image struct {
	URL string `json:"url"`
}

func storeImageBucket(reader io.Reader, path string) (string, error) {
	supabaseUrl := os.Getenv("SUPABASE_URL")
	supabaseKey := os.Getenv("SUPABASE_KEY")

	bucket := Bucket{
		BucketId: "generated_images",
		BaseURL:  supabaseUrl,
		APIKey:   supabaseKey,
	}

	opts := DefaultFileUploadOptions()

	opts.ContentType = "image/png"

	res, err := bucket.Upload(path, reader, opts)
	if err != nil {
		return "", err
	}
	return res, nil
}

func ImageGeneration(w http.ResponseWriter, r *http.Request, _ router.Params) {
	prompt := r.URL.Query().Get("p")
	provider := r.URL.Query().Get("provider")
	record := r.Context().Value("recordEvent").(utils.RecordFunc)

	if provider == "" {
		provider = "openai"
	}

	if provider != "openai" {
		utils.RespondError(w, record, "unknown_model_provider")
		return
	}

	reader, err := DALLEGenerate(prompt)
	if err != nil {
		utils.RespondError(w, record, "image_generation_error")
		return
	}

	format := r.URL.Query().Get("format")

	if format == "json" {
		id := uuid.New().String()

		url, err := storeImageBucket(reader, id+".png")
		if err != nil {
			utils.RespondError(w, record, "storage_error")
			return
		}

		image := Image{URL: url}

		response, _ := json.Marshal(&image)
		record(string(response))

		json.NewEncoder(w).Encode(image)
	} else {
		record("[Raw image]")
		io.Copy(w, reader)
	}
}
