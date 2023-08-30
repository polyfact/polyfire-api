package image_generation

import (
	"encoding/json"
	"io"
	"net/http"
	"os"

	"github.com/google/uuid"
	router "github.com/julienschmidt/httprouter"
	posthog "github.com/polyfact/api/posthog"
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
	user_id := r.Context().Value("user_id").(string)
	posthog.ImageGenerationEvent(user_id)
	prompt := r.URL.Query().Get("p")
	provider := r.URL.Query().Get("provider")

	if provider == "" {
		provider = "openai"
	}

	var reader io.Reader
	var err error
	switch provider {
	case "openai":
		reader, err = DALLEGenerate(prompt)
	case "midjourney":
		reader, err = MJGenerate(prompt)
	default:
		utils.RespondError(w, "unknown_model_provider")
		return
	}
	if err != nil {
		utils.RespondError(w, "image_generation_error")
		return
	}

	format := r.URL.Query().Get("format")

	if format == "json" {
		id := uuid.New().String()

		url, err := storeImageBucket(reader, id+".png")
		if err != nil {
			utils.RespondError(w, "storage_error")
			return
		}

		json.NewEncoder(w).Encode(Image{
			URL: url,
		})
	} else {
		io.Copy(w, reader)
	}
}
