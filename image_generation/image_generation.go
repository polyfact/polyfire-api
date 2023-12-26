package imagegeneration

import (
	"encoding/json"
	"io"
	"net/http"
	"os"

	"github.com/google/uuid"
	router "github.com/julienschmidt/httprouter"
	database "github.com/polyfire/api/db"
	"github.com/polyfire/api/utils"
)

type Image struct {
	URL string `json:"url"`
}

func storeImageBucket(reader io.Reader, path string) (string, error) {
	supabaseURL := os.Getenv("SUPABASE_URL")
	supabaseKey := os.Getenv("SUPABASE_KEY")

	bucket := Bucket{
		BucketID: "generated_images",
		BaseURL:  supabaseURL,
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
	db := r.Context().Value(utils.ContextKeyDB).(database.Database)
	userID := r.Context().Value(utils.ContextKeyUserID).(string)
	request, _ := json.Marshal(r.URL.Query())
	prompt := r.URL.Query().Get("p")
	model := r.URL.Query().Get("model")
	recordEventRequest := r.Context().Value(utils.ContextKeyRecordEventRequest).(utils.RecordRequestFunc)

	var record utils.RecordFunc = func(response string, props ...utils.KeyValue) {
		recordEventRequest(string(request), response, userID, props...)
	}
	rateLimitStatus := r.Context().Value(utils.ContextKeyRateLimitStatus)

	if rateLimitStatus == database.RateLimitStatusReached {
		utils.RespondError(w, record, "rate_limit_reached")
		return
	}

	creditsStatus := r.Context().Value(utils.ContextKeyCreditsStatus)

	if creditsStatus == database.CreditsStatusUsedUp {
		utils.RespondError(w, record, "credits_used_up")
		return
	}

	if model == "" {
		model = "dall-e-2"
	}

	if model != "dall-e-2" && model != "dall-e-3" {
		utils.RespondError(w, record, "unknown_model")
		return
	}

	reader, err := DALLEGenerate(r.Context(), prompt, model)

	db.LogRequests(
		r.Context().Value(utils.ContextKeyEventID).(string),
		userID, "openai", model, 0, 0, "image", true)

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
