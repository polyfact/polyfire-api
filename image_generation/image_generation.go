package image_generation

import (
	"encoding/json"
	"io"
	"net/http"
	"os"

	"github.com/google/uuid"
	router "github.com/julienschmidt/httprouter"
	"github.com/polyfact/api/db"
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
	user_id := r.Context().Value(utils.ContextKeyUserID).(string)
	request, _ := json.Marshal(r.URL.Query())
	prompt := r.URL.Query().Get("p")
	provider := r.URL.Query().Get("provider")
	recordEventRequest := r.Context().Value(utils.ContextKeyRecordEventRequest).(utils.RecordRequestFunc)

	var record utils.RecordFunc = func(response string, props ...utils.KeyValue) {
		recordEventRequest(string(request), response, user_id, props...)
	}
	rateLimitStatus := r.Context().Value(utils.ContextKeyRateLimitStatus)

	if rateLimitStatus == db.RateLimitStatusReached {
		utils.RespondError(w, record, "rate_limit_reached")
		return
	}

	if rateLimitStatus == db.RateLimitStatusProjectReached {
		utils.RespondError(w, record, "project_rate_limit_reached")
		return
	}

	if provider == "" {
		provider = "openai"
	}

	// premium, _ := db.ProjectIsPremium(user_id)

	// if !premium {
	// 	utils.RespondError(w, record, "project_not_premium_model")
	// 	return
	// }

	if provider != "openai" {
		utils.RespondError(w, record, "unknown_model_provider")
		return
	}

	reader, err := DALLEGenerate(prompt)

	db.LogRequests(user_id, "openai", "dalle-2", 0, 0, "image", true)

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

		_ = json.NewEncoder(w).Encode(image)
	} else {
		record("[Raw image]")
		_, _ = io.Copy(w, reader)
	}
}
