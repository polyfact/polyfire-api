package main

import (
	"log"
	"net/http"

	httprouter "github.com/julienschmidt/httprouter"

	auth "github.com/polyfire/api/auth"
	completion "github.com/polyfire/api/completion"
	db "github.com/polyfire/api/db"
	imageGeneration "github.com/polyfire/api/image_generation"
	kv "github.com/polyfire/api/kv"
	memory "github.com/polyfire/api/memory"
	middlewares "github.com/polyfire/api/middlewares"
	stt "github.com/polyfire/api/stt"
	tts "github.com/polyfire/api/tts"
	utils "github.com/polyfire/api/utils"
)

type CORSRouter struct {
	Router *httprouter.Router
}

func (h CORSRouter) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Headers", "*")
	w.Header().Set("Access-Control-Allow-Methods", "*")

	middlewares.AddRecord(r, utils.Unknown)
	defer middlewares.RecoverFromPanic(w, r)

	h.Router.ServeHTTP(w, r)
}

func GlobalMiddleware(router *httprouter.Router) http.Handler {
	return &CORSRouter{Router: router}
}

func main() {
	log.Print("Starting the server on :8080")

	db.InitDB()

	router := httprouter.New()

	// Auth Routes
	router.GET("/project/:id/auth/firebase", middlewares.Record(utils.AuthFirebase, auth.ExternalFirebaseTokenExchangeHandler))
	router.GET("/project/:id/auth/custom", middlewares.Record(utils.AuthCustom, auth.ExternalCustomTokenExchangeHandler))
	router.GET("/project/:id/auth/anonymous", middlewares.Record(utils.AuthAnonymous, auth.AnonymousTokenExchangeHandler))
	router.GET("/project/:id/auth/provider/redirect", middlewares.Record(utils.AuthProviderRedirection, auth.RedirectAuth))
	router.GET("/project/:id/auth/provider/callback", middlewares.Record(utils.AuthProviderCallback, auth.CallbackAuth))
	router.POST("/project/:id/auth/provider/refresh", middlewares.Record(utils.AuthProviderRefresh, auth.RefreshToken))
	router.GET("/auth/id", middlewares.Record(utils.AuthID, middlewares.Auth(auth.GetAuthID)))

	router.GET("/usage", middlewares.Record(utils.Usage, middlewares.Auth(auth.UserRateLimit)))

	// Completion Routes
	router.POST("/generate", middlewares.Record(utils.Generate, middlewares.Auth(completion.Generate)))
	router.GET("/chat/:id/history", middlewares.Record(utils.ChatHistory, middlewares.Auth(completion.GetChatHistory)))
	router.POST("/chats", middlewares.Record(utils.ChatCreate, middlewares.Auth(completion.CreateChat)))
	router.GET("/stream", middlewares.Record(utils.Generate, middlewares.AuthStream(completion.Stream)))

	// Transcription Routes
	router.POST("/transcribe", middlewares.Record(utils.SpeechToText, middlewares.Auth(stt.Transcribe)))

	// TTS Routes
	router.POST("/tts", middlewares.Record(utils.TextToSpeech, middlewares.Auth(tts.Handler)))

	// Image Generation Routes
	router.GET("/image/generate", middlewares.Record(utils.ImageGeneration, middlewares.Auth(imageGeneration.ImageGeneration)))

	// Memory Routes
	router.GET("/memories", middlewares.Record(utils.MemoryList, middlewares.Auth(memory.Get)))
	router.POST("/memory/:id/search", middlewares.Record(utils.MemorySearch, middlewares.Auth(memory.Search)))
	router.POST("/memory", middlewares.Record(utils.MemoryCreate, middlewares.Auth(memory.Create)))
	router.PUT("/memory", middlewares.Record(utils.MemoryAdd, middlewares.Auth(memory.Add)))

	// KV Routes
	router.GET("/kv", middlewares.Record(utils.KVGet, middlewares.Auth(kv.Get)))
	router.GET("/kvs", middlewares.Record(utils.KVList, middlewares.Auth(kv.List)))
	router.PUT("/kv", middlewares.Record(utils.KVSet, middlewares.Auth(kv.Set)))
	router.DELETE("/kv", middlewares.Record(utils.KVDelete, middlewares.Auth(kv.Delete)))

	log.Fatal(http.ListenAndServe(":8080", GlobalMiddleware(router)))
}
