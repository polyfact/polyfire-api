package stt

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime"
	"mime/multipart"
	"net/http"
	"os"
	"strings"
	"sync"

	router "github.com/julienschmidt/httprouter"
	supa "github.com/nedpals/supabase-go"

	database "github.com/polyfire/api/db"
	providers "github.com/polyfire/api/stt/providers"
	"github.com/polyfire/api/utils"
)

func DownloadFromBucket(bucket string, path string) ([]byte, error) {
	supabaseURL := os.Getenv("SUPABASE_URL")
	supabaseKey := os.Getenv("SUPABASE_KEY")

	supabase := supa.CreateClient(supabaseURL, supabaseKey)

	return supabase.Storage.From(bucket).Download(path)
}

type TranscribeRequestBody struct {
	FilePath     string   `json:"file_path"`
	Provider     string   `json:"provider"`
	Language     *string  `json:"language,omitempty"`
	Keywords     []string `json:"keywords,omitempty"`
	OutputFormat *string  `json:"output_format,omitempty"`
}

type Result struct {
	Text string `json:"text"`
}

func Transcribe(w http.ResponseWriter, r *http.Request, _ router.Params) {
	ctx := r.Context()
	db := ctx.Value(utils.ContextKeyDB).(database.Database)
	userID := ctx.Value(utils.ContextKeyUserID).(string)
	record := ctx.Value(utils.ContextKeyRecordEvent).(utils.RecordFunc)

	rateLimitStatus := ctx.Value(utils.ContextKeyRateLimitStatus)

	if rateLimitStatus == database.RateLimitStatusReached {
		utils.RespondError(w, record, "rate_limit_reached")
		return
	}

	creditsStatus := ctx.Value(utils.ContextKeyCreditsStatus)

	if creditsStatus == database.CreditsStatusUsedUp {
		utils.RespondError(w, record, "credits_used_up")
		return
	}

	contentType := r.Header.Get("Content-Type")
	var fileBufReader io.Reader

	var input TranscribeRequestBody
	providerName := ""
	if contentType == "application/json" {

		err := json.NewDecoder(r.Body).Decode(&input)
		if err != nil {
			utils.RespondError(w, record, "invalid_json")
			return
		}

		providerName = input.Provider
		b, err := DownloadFromBucket("audio_transcribes", input.FilePath)
		if err != nil {
			fmt.Println(err)
			utils.RespondError(w, record, "read_error")
			return
		}

		fileBufReader = bytes.NewReader(b)
	} else {
		_, p, _ := mime.ParseMediaType(contentType)
		boundary := p["boundary"]
		reader := multipart.NewReader(r.Body, boundary)
		part, err := reader.NextPart()
		if err == io.EOF {
			utils.RespondError(w, record, "missing_content")
			return
		}
		if err != nil {
			utils.RespondError(w, record, "read_error", err.Error())
			return
		}
		fileBufReader = bufio.NewReader(part)
	}

	var files []io.Reader
	duration := 0
	if providerName == "openai" || providerName == "" {
		var closeFunc func()
		var err error
		// The format doesn't seem to really matter
		files, duration, closeFunc, err = SplitFile(fileBufReader)
		if err != nil {
			fmt.Println(err)
			utils.RespondError(w, record, "splitting_error")
			return
		}
		defer closeFunc()
	} else {
		files = []io.Reader{fileBufReader}
	}

	provider, err := providers.NewProvider(providerName)
	if err != nil {
		utils.RespondError(w, record, "invalid_model_provider")
		return
	}

	var res providers.TranscriptionResult

	texts := make([]string, len(files))
	wordLists := make([][]providers.Word, len(files))
	dialogues := make([][]providers.DialogueElement, len(files))

	var wg sync.WaitGroup
	for i, r := range files {
		fileReader := r
		chunkNb := i
		wg.Add(1)
		go func() {
			defer wg.Done()
			silences, noSilenceFileReader, closeFunc, err := RemoveSilence(fileReader)
			defer closeFunc()
			if err != nil {
				fmt.Printf("%v\n", err)
				utils.RespondError(w, record, "transcription_error")
				return
			}
			resTmp, err := provider.Transcribe(ctx, noSilenceFileReader, providers.TranscriptionInputOptions{
				Format:   "mpeg",
				Language: input.Language,
				Keywords: input.Keywords,
			})
			if err != nil {
				fmt.Printf("%v\n", err)
				utils.RespondError(w, record, "transcription_error")
				return
			}
			texts[chunkNb] = resTmp.Text
			/* WARNING: The timestamp will be wrong if the file has been chunked because
			 * based on the chunked time and not the full time. For now I choose to ignore
			 * it since the chunking is only used with whisper and the timestamp per words
			 * is only available on deepgram. If it causes a problem at some point it might
			 * be a good idea to move the chunking to the whisper provider function and add
			 * an equivalent timestamp translation step */
			wordLists[chunkNb] = AddSilenceToWordTimestamps(silences, resTmp.Words)
			dialogues[chunkNb] = AddSilenceToDialogueTimestamps(silences, resTmp.Dialogue)
			fmt.Printf("Transcription %v/%v\n", chunkNb+1, len(files))
		}()
	}

	wg.Wait()

	res.Words = make([]providers.Word, 0)
	res.Dialogue = make([]providers.DialogueElement, 0)
	for i, l := range wordLists {
		res.Words = append(res.Words, l...)
		for _, k := range dialogues[i] {
			k.Speaker = i*100 + k.Speaker
			res.Dialogue = append(res.Dialogue, k)
		}
	}
	res.Text = strings.Trim(strings.Join(texts, " "), " \t\n")
	db.LogRequestsCredits(
		r.Context().Value(utils.ContextKeyEventID).(string),
		userID, "whisper", duration*1000, 0, 0, "transcription")

	response, _ := json.Marshal(&res)
	record(string(response))

	_ = json.NewEncoder(w).Encode(res)
}
