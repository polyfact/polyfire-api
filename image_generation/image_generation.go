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

	gen, err := Generate(prompt)
	if err != nil {
		utils.RespondError(w, "image_generation_error")
		return
	}

	format := r.URL.Query().Get("format")

	if format == "json" {
		reader, err := (gen.Data[0]).Reader()
		if err != nil {
			utils.RespondError(w, "read_error")
			return
		}

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
		reader, err := (gen.Data[0]).Reader()
		if err != nil {
			utils.RespondError(w, "read_error")
			return
		}

		io.Copy(w, reader)
	}
}
